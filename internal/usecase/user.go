package usecase

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/fazriegi/fintrack-be/internal/domain"
	"github.com/fazriegi/fintrack-be/internal/repository"
	"github.com/fazriegi/fintrack-be/pkg"
	"github.com/fazriegi/fintrack-be/pkg/constant"
	"github.com/fazriegi/fintrack-be/pkg/password"
	"github.com/fazriegi/fintrack-be/pkg/token"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type userUsecase struct {
	db   *sqlx.DB
	log  *log.Logger
	repo repository.UserRepository
}

type UserUsecase interface {
	Register(ctx context.Context, req *domain.RegisterRequest) (resp pkg.Response)
	Login(ctx context.Context, req *domain.LoginRequest) (resp pkg.Response)
	RefreshToken(ctx context.Context, refreshToken string) (resp pkg.Response)
	Profile(ctx context.Context, accessToken string) (resp pkg.Response)
	Logout(ctx context.Context, accessToken, refreshToken string) (resp pkg.Response)
}

func NewUserUsecase(db *sqlx.DB, log *log.Logger, repo repository.UserRepository) UserUsecase {
	return &userUsecase{db, log, repo}
}

func (uc *userUsecase) Register(ctx context.Context, req *domain.RegisterRequest) pkg.Response {
	existingUser, err := uc.repo.GetByEmail(ctx, req.Email, uc.db)

	if err != nil && err != errors.New(constant.ErrUserNotFound) {
		uc.log.Printf("[ERROR] repo.GetByEmail: %s", err.Error())
		return pkg.NewResponse(http.StatusInternalServerError, constant.ErrServer, nil, nil)
	}

	if existingUser != nil {
		return pkg.NewResponse(http.StatusBadRequest, constant.ErrEmailExists, nil, nil)
	}

	hash, err := password.Hash(req.Password)
	if err != nil {
		uc.log.Printf("[ERROR] password.Hash: %s", err.Error())
		return pkg.NewResponse(http.StatusInternalServerError, constant.ErrServer, nil, nil)
	}

	tx, err := uc.db.Beginx()
	if err != nil {
		uc.log.Printf("[ERROR] error start transaction: %s", err.Error())
		return pkg.NewResponse(http.StatusInternalServerError, constant.ErrServer, nil, nil)
	}
	defer tx.Rollback()

	user := &domain.User{
		Email:    req.Email,
		Password: hash,
		FullName: req.FullName,
	}

	userId, err := uc.repo.Create(ctx, user, tx)
	if err != nil {
		uc.log.Printf("[ERROR] repo.Create: %s", err.Error())
		return pkg.NewResponse(http.StatusInternalServerError, constant.ErrServer, nil, nil)
	}

	if err := uc.repo.SeedDefaultCategories(ctx, tx, userId); err != nil {
		uc.log.Printf("[ERROR] repo.SeedDefaultCategories: %s", err.Error())
		return pkg.NewResponse(http.StatusInternalServerError, constant.ErrServer, nil, nil)
	}

	if err := tx.Commit(); err != nil {
		uc.log.Printf("[ERROR] commit transaction: %s", err.Error())
		return pkg.NewResponse(http.StatusInternalServerError, constant.ErrServer, nil, nil)
	}

	return pkg.NewResponse(http.StatusCreated, "User created successfully", nil, nil)
}

func (uc *userUsecase) Login(ctx context.Context, req *domain.LoginRequest) (resp pkg.Response) {
	user, err := uc.repo.GetByEmail(ctx, req.Email, uc.db)
	if err != nil {
		uc.log.Printf("[ERROR] repo.GetByEmail: %s", err.Error())
		return pkg.NewResponse(http.StatusUnauthorized, constant.ErrInvalidCreds, nil, nil)
	}

	if !password.Check(req.Password, user.Password) {
		return pkg.NewResponse(http.StatusUnauthorized, constant.ErrInvalidCreds, nil, nil)
	}

	accessToken, err := token.GenerateAccessToken(user.ID.String())
	if err != nil {
		uc.log.Printf("[ERROR] token.GenerateAccessToken: %s", err.Error())
		return pkg.NewResponse(http.StatusInternalServerError, constant.ErrServer, nil, nil)
	}

	refreshToken, err := token.GenerateRefreshToken(user.ID.String())
	if err != nil {
		uc.log.Printf("[ERROR] token.GenerateRefreshToken: %s", err.Error())
		return pkg.NewResponse(http.StatusInternalServerError, constant.ErrServer, nil, nil)
	}

	tx, err := uc.db.Beginx()
	if err != nil {
		uc.log.Printf("[ERROR] error start transaction: %s", err.Error())
		return pkg.NewResponse(http.StatusInternalServerError, constant.ErrServer, nil, nil)
	}
	defer tx.Rollback()

	if err := uc.repo.InsertRefreshToken(ctx, domain.RefreshToken{
		UserID:     user.ID,
		Token:      refreshToken,
		ExpiresAt:  time.Now().Add(7 * 24 * time.Hour),
		DeviceInfo: "",
		IPAddress:  "",
	}, tx); err != nil {
		uc.log.Printf("[ERROR] repo.InsertRefreshToken: %s", err.Error())
		return pkg.NewResponse(http.StatusInternalServerError, constant.ErrServer, nil, nil)
	}

	if err := tx.Commit(); err != nil {
		uc.log.Printf("[ERROR] commit transaction: %s", err.Error())
		return pkg.NewResponse(http.StatusInternalServerError, constant.ErrServer, nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Login successful", map[string]any{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	}, nil)
}

func (uc *userUsecase) RefreshToken(ctx context.Context, refreshToken string) (resp pkg.Response) {
	claims, err := token.ValidateToken(refreshToken)
	if err != nil {
		uc.log.Printf("[ERROR] token.ValidateToken: %s", err.Error())
		return pkg.NewResponse(http.StatusUnauthorized, constant.ErrInvalidToken, nil, nil)
	}

	parsedUserID, err := uuid.Parse(claims.UserID)
	if err != nil {
		uc.log.Printf("[ERROR] uuid.Parse - invalid UUID format in claims: %s", err.Error())
		return pkg.NewResponse(http.StatusUnauthorized, constant.ErrInvalidToken, nil, nil)
	}

	tx, err := uc.db.Beginx()
	if err != nil {
		uc.log.Printf("[ERROR] error start transaction: %s", err.Error())
		return pkg.NewResponse(http.StatusInternalServerError, constant.ErrServer, nil, nil)
	}
	defer tx.Rollback()

	err = uc.repo.CheckRefreshToken(ctx, parsedUserID, refreshToken, tx)
	if err != nil {
		if err.Error() != constant.ErrNotFound {
			uc.log.Printf("[ERROR] repo.CheckRefreshToken: %s", err.Error())
		}

		return pkg.NewResponse(http.StatusUnauthorized, constant.ErrInvalidToken, nil, nil)
	}

	newAccessToken, err := token.GenerateAccessToken(claims.UserID)
	if err != nil {
		uc.log.Printf("[ERROR] token.GenerateAccessToken: %s", err.Error())
		return pkg.NewResponse(http.StatusInternalServerError, constant.ErrServer, nil, nil)
	}

	newRefreshToken, err := token.GenerateRefreshToken(claims.UserID)
	if err != nil {
		uc.log.Printf("[ERROR] token.GenerateRefreshToken: %s", err.Error())
		return pkg.NewResponse(http.StatusInternalServerError, constant.ErrServer, nil, nil)
	}

	if err := uc.repo.RevokeRefreshToken(ctx, parsedUserID, refreshToken, tx); err != nil {
		uc.log.Printf("[ERROR] repo.RevokeRefreshToken: %s", err.Error())
	}

	if err := uc.repo.InsertRefreshToken(ctx, domain.RefreshToken{
		UserID:     parsedUserID,
		Token:      newRefreshToken,
		ExpiresAt:  time.Now().Add(7 * 24 * time.Hour),
		DeviceInfo: "",
		IPAddress:  "",
	}, tx); err != nil {
		uc.log.Printf("[ERROR] repo.InsertRefreshToken: %s", err.Error())
		return pkg.NewResponse(http.StatusInternalServerError, constant.ErrServer, nil, nil)
	}

	if err := tx.Commit(); err != nil {
		uc.log.Printf("[ERROR] commit transaction: %s", err.Error())
		return pkg.NewResponse(http.StatusInternalServerError, constant.ErrServer, nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Success", map[string]any{
		"access_token":  newAccessToken,
		"refresh_token": newRefreshToken,
	}, nil)
}

func (uc *userUsecase) Profile(ctx context.Context, accessToken string) (resp pkg.Response) {
	claims, err := token.ValidateToken(accessToken)
	if err != nil {
		uc.log.Printf("[ERROR] token.ValidateToken: %s", err.Error())
		return pkg.NewResponse(http.StatusUnauthorized, constant.ErrInvalidToken, nil, nil)
	}

	parsedUserID, err := uuid.Parse(claims.UserID)
	if err != nil {
		uc.log.Printf("[ERROR] uuid.Parse - invalid UUID format in claims: %s", err.Error())
		return pkg.NewResponse(http.StatusUnauthorized, constant.ErrInvalidToken, nil, nil)
	}

	user, err := uc.repo.GetByID(ctx, parsedUserID, uc.db)
	if err != nil {
		if err.Error() != constant.ErrUserNotFound {
			uc.log.Printf("[ERROR] repo.GetByID: %s", err.Error())
			return pkg.NewResponse(http.StatusInternalServerError, constant.ErrServer, nil, nil)
		}

		return pkg.NewResponse(http.StatusInternalServerError, constant.ErrUserNotFound, nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Success", user, nil)
}

func (uc *userUsecase) Logout(ctx context.Context, accessToken, refreshToken string) (resp pkg.Response) {
	claims, err := token.ValidateToken(accessToken)
	if err != nil {
		uc.log.Printf("[ERROR] token.ValidateToken: %s", err.Error())
	}

	parsedUserID, err := uuid.Parse(claims.UserID)
	if err != nil {
		uc.log.Printf("[ERROR] uuid.Parse - invalid UUID format in claims: %s", err.Error())
	}

	tx, err := uc.db.Beginx()
	if err != nil {
		uc.log.Printf("[ERROR] error start transaction: %s", err.Error())
	}
	defer tx.Rollback()

	if err := uc.repo.RevokeRefreshToken(ctx, parsedUserID, refreshToken, tx); err != nil {
		uc.log.Printf("[ERROR] repo.RevokeRefreshToken: %s", err.Error())
	}

	if err := tx.Commit(); err != nil {
		uc.log.Printf("[ERROR] commit transaction: %s", err.Error())
	}

	return pkg.NewResponse(http.StatusOK, "Success", nil, nil)
}
