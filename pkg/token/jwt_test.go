package token

import (
	"os"
	"testing"
)

func TestJWTValidationAndTypeSeparation(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret-key-for-unit-testing-123456")
	if err := InitSecretKey(); err != nil {
		t.Fatalf("InitSecretKey failed: %v", err)
	}

	userID := "11111111-2222-3333-4444-555555555555"

	access1, err := GenerateAccessToken(userID)
	if err != nil {
		t.Fatalf("GenerateAccessToken failed: %v", err)
	}

	refresh1, err := GenerateRefreshToken(userID)
	if err != nil {
		t.Fatalf("GenerateRefreshToken failed: %v", err)
	}
	refresh2, err := GenerateRefreshToken(userID)
	if err != nil {
		t.Fatalf("GenerateRefreshToken 2 failed: %v", err)
	}

	if refresh1 == refresh2 {
		t.Errorf("expected two refresh tokens generated in the same second to have unique JTI, got identical tokens")
	}

	// Access token should pass ValidateAccessToken and fail ValidateRefreshToken
	if _, err := ValidateAccessToken(access1); err != nil {
		t.Errorf("ValidateAccessToken rejected valid access token: %v", err)
	}
	if _, err := ValidateRefreshToken(access1); err == nil {
		t.Errorf("ValidateRefreshToken accepted an access token")
	}

	// Refresh token should pass ValidateRefreshToken and fail ValidateAccessToken
	if _, err := ValidateRefreshToken(refresh1); err != nil {
		t.Errorf("ValidateRefreshToken rejected valid refresh token: %v", err)
	}
	if _, err := ValidateAccessToken(refresh1); err == nil {
		t.Errorf("ValidateAccessToken accepted a refresh token")
	}

	// Malformed token must not panic
	if _, err := ValidateToken("malformed.token"); err == nil {
		t.Errorf("expected error for malformed token")
	}
	if _, err := ValidateToken(""); err == nil {
		t.Errorf("expected error for empty token")
	}

	// Empty secret must fail
	SecretKey = nil
	os.Unsetenv("JWT_SECRET")
	if _, err := GenerateAccessToken(userID); err == nil {
		t.Errorf("expected error when JWT_SECRET is unset")
	}
}

