package token

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"
)

var (
	SecretKey          = []byte(os.Getenv("JWT_SECRET"))
	ErrSecretNotSet    = errors.New("JWT_SECRET environment variable is not set")
	ErrInvalidTokenUse = errors.New("invalid token type")
)

type Claims struct {
	UserID    string `json:"user_id"`
	TokenType string `json:"token_type"`
	jwt.RegisteredClaims
}

func InitSecretKey() error {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return ErrSecretNotSet
	}
	SecretKey = []byte(secret)
	return nil
}

func getSecretKey() ([]byte, error) {
	if len(SecretKey) == 0 {
		secret := os.Getenv("JWT_SECRET")
		if secret != "" {
			SecretKey = []byte(secret)
		}
	}
	if len(SecretKey) == 0 {
		return nil, ErrSecretNotSet
	}
	return SecretKey, nil
}

func GenerateAccessToken(userID string) (string, error) {
	secret, err := getSecretKey()
	if err != nil {
		return "", err
	}

	now := time.Now()
	claims := &Claims{
		UserID:    userID,
		TokenType: TokenTypeAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.NewString(),
			ExpiresAt: jwt.NewNumericDate(now.Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

func GenerateRefreshToken(userID string) (string, error) {
	secret, err := getSecretKey()
	if err != nil {
		return "", err
	}

	now := time.Now()
	claims := &Claims{
		UserID:    userID,
		TokenType: TokenTypeRefresh,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.NewString(),
			ExpiresAt: jwt.NewNumericDate(now.Add(7 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

func ValidateToken(tokenString string) (*Claims, error) {
	secret, err := getSecretKey()
	if err != nil {
		return nil, err
	}

	token, err := jwt.ParseWithClaims(
		tokenString,
		&Claims{},
		func(t *jwt.Token) (interface{}, error) {
			return secret, nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
	)
	if err != nil || token == nil {
		if err != nil {
			return nil, err
		}
		return nil, jwt.ErrTokenMalformed
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, jwt.ErrTokenInvalidClaims
}

func ValidateAccessToken(tokenString string) (*Claims, error) {
	claims, err := ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}
	if claims.TokenType != TokenTypeAccess {
		return nil, ErrInvalidTokenUse
	}
	return claims, nil
}

func ValidateRefreshToken(tokenString string) (*Claims, error) {
	claims, err := ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}
	if claims.TokenType != TokenTypeRefresh {
		return nil, ErrInvalidTokenUse
	}
	return claims, nil
}
