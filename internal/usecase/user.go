package usecase

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/fazriegi/fintrack-be/internal/domain"
	"github.com/fazriegi/fintrack-be/pkg"
	"github.com/fazriegi/fintrack-be/pkg/constant"
	"github.com/fazriegi/fintrack-be/pkg/password"
	"github.com/fazriegi/fintrack-be/pkg/token"
	"github.com/google/uuid"
)

type userUsecase struct {
	log  *log.Logger
	repo domain.UserRepository
	tx   domain.TransactionManager
}

type UserUsecase interface {
	Register(ctx context.Context, req *domain.RegisterRequest) (resp pkg.Response)
	Login(ctx context.Context, req *domain.LoginRequest) (resp pkg.Response)
	RefreshToken(ctx context.Context, refreshToken, remoteAddr string) (resp pkg.Response)
	Profile(ctx context.Context, accessToken string) (resp pkg.Response)
	Logout(ctx context.Context, accessToken, refreshToken string) (resp pkg.Response)
	CleanupExpiredTokens(ctx context.Context) error
}

func NewUserUsecase(log *log.Logger, repo domain.UserRepository, tx domain.TransactionManager) UserUsecase {
	return &userUsecase{log, repo, tx}
}

func (uc *userUsecase) Register(ctx context.Context, req *domain.RegisterRequest) pkg.Response {
	existingUser, err := uc.repo.GetByUsername(ctx, req.Username)
	if err != nil && err.Error() != constant.ErrUserNotFound {
		uc.log.Printf("[ERROR] repo.GetByUsername: %s", err.Error())
		return pkg.NewResponse(http.StatusInternalServerError, constant.ErrServer, nil, nil)
	}

	if existingUser != nil {
		return pkg.NewResponse(http.StatusBadRequest, constant.ErrUsernameExists, nil, nil)
	}

	hash, err := password.Hash(req.Password)
	if err != nil {
		uc.log.Printf("[ERROR] password.Hash: %s", err.Error())
		return pkg.NewResponse(http.StatusInternalServerError, constant.ErrServer, nil, nil)
	}

	user := &domain.User{
		Username: req.Username,
		Email:    req.Email,
		Password: hash,
		FullName: req.FullName,
	}

	var userId uuid.UUID
	err = uc.tx.WithTransaction(ctx, func(txCtx context.Context) error {
		var err error
		userId, err = uc.repo.Create(txCtx, user)
		if err != nil {
			return err
		}

		if err := uc.repo.SeedDefaultCategories(txCtx, userId); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		uc.log.Printf("[ERROR] Register transaction failed: %s", err.Error())
		return pkg.NewResponse(http.StatusInternalServerError, constant.ErrServer, nil, nil)
	}

	return pkg.NewResponse(http.StatusCreated, "User created successfully", nil, nil)
}

func (uc *userUsecase) Login(ctx context.Context, req *domain.LoginRequest) (resp pkg.Response) {
	user, err := uc.repo.GetByUsername(ctx, req.Username)
	if err != nil {
		uc.log.Printf("[ERROR] repo.GetByUsername: %s", err.Error())
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

	err = uc.tx.WithTransaction(ctx, func(txCtx context.Context) error {
		return uc.repo.InsertRefreshToken(txCtx, domain.RefreshToken{
			UserID:     user.ID,
			Token:      refreshToken,
			ExpiresAt:  time.Now().Add(7 * 24 * time.Hour),
			DeviceInfo: "",
			IPAddress:  req.RemoteAddr,
		})
	})
	if err != nil {
		uc.log.Printf("[ERROR] InsertRefreshToken transaction failed: %s", err.Error())
		return pkg.NewResponse(http.StatusInternalServerError, constant.ErrServer, nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Login successful", map[string]any{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"user":          user,
	}, nil)
}

func (uc *userUsecase) RefreshToken(ctx context.Context, refreshToken, remoteAddr string) (resp pkg.Response) {
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

	tokenExp, err := uc.repo.CheckRefreshToken(ctx, parsedUserID, refreshToken)
	if err != nil {
		if err.Error() != constant.ErrNotFound {
			uc.log.Printf("[ERROR] repo.CheckRefreshToken: %s", err.Error())
		}

		return pkg.NewResponse(http.StatusUnauthorized, constant.ErrInvalidToken, nil, nil)
	}

	threeHour := time.Now().Add(3 * time.Hour)
	if threeHour.After(tokenExp) {
		err = uc.tx.WithTransaction(ctx, func(txCtx context.Context) error {
			if err := uc.repo.RevokeRefreshToken(txCtx, parsedUserID, refreshToken); err != nil {
				return err
			}

			var err error
			refreshToken, err = token.GenerateRefreshToken(parsedUserID.String())
			if err != nil {
				return err
			}

			if err := uc.repo.InsertRefreshToken(txCtx, domain.RefreshToken{
				UserID:     parsedUserID,
				Token:      refreshToken,
				ExpiresAt:  time.Now().Add(7 * 24 * time.Hour),
				DeviceInfo: "",
				IPAddress:  remoteAddr,
			}); err != nil {
				return err
			}

			return nil
		})
		if err != nil {
			uc.log.Printf("[ERROR] RefreshToken transaction failed: %s", err.Error())
			return pkg.NewResponse(http.StatusUnauthorized, constant.ErrInvalidToken, nil, nil)
		}
	}

	newAccessToken, err := token.GenerateAccessToken(claims.UserID)
	if err != nil {
		uc.log.Printf("[ERROR] token.GenerateAccessToken: %s", err.Error())
		return pkg.NewResponse(http.StatusUnauthorized, constant.ErrInvalidToken, nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Success", map[string]any{
		"access_token":  newAccessToken,
		"refresh_token": refreshToken,
	}, nil)
}

func (uc *userUsecase) Profile(ctx context.Context, accessToken string) (resp pkg.Response) {
	userId := ctx.Value("user_id").(uuid.UUID)

	user, err := uc.repo.GetByID(ctx, userId)
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

	_ = uc.tx.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := uc.repo.RevokeRefreshToken(txCtx, parsedUserID, refreshToken); err != nil {
			uc.log.Printf("[ERROR] repo.RevokeRefreshToken: %s", err.Error())
			return err
		}
		return nil
	})

	return pkg.NewResponse(http.StatusOK, "Success", nil, nil)
}

func (uc *userUsecase) CleanupExpiredTokens(ctx context.Context) error {
	return uc.repo.RemoveExpiredToken(ctx, nil)
}
