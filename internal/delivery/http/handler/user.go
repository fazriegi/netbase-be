package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/fazriegi/fintrack-be/internal/delivery/http/middleware"
	"github.com/fazriegi/fintrack-be/internal/domain"
	"github.com/fazriegi/fintrack-be/internal/usecase"
	"github.com/fazriegi/fintrack-be/pkg"
	"github.com/fazriegi/fintrack-be/pkg/constant"
	"github.com/fazriegi/fintrack-be/pkg/validator"
)

type UserHandler struct {
	usecase usecase.UserUsecase
	logger  *log.Logger
}

func NewUserHandler(mux *http.ServeMux, uc usecase.UserUsecase, logger *log.Logger) {
	handler := &UserHandler{
		usecase: uc,
		logger:  logger,
	}

	mux.HandleFunc("POST /v1/register", handler.Register)
	mux.HandleFunc("POST /v1/login", handler.Login)
	mux.HandleFunc("POST /v1/refresh_token", handler.RefreshToken)
	mux.HandleFunc("POST /v1/logout", handler.Logout)

	mux.Handle("GET /v1/profile", middleware.MiddlewareAuth(http.HandlerFunc(handler.Profile)))
}

func (h *UserHandler) setAuthCookies(w http.ResponseWriter, accessToken, refreshToken string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   true, // Set true in production (HTTPS)
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(15 * time.Minute),
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(7 * 24 * time.Hour),
	})
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req domain.RegisterRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		pkg.NewResponse(http.StatusBadRequest, constant.ErrInvalidJson, nil, nil).HTTP(w)
		return
	}

	// validation
	validationErr := validator.ValidateRequest(&req)
	if len(validationErr) > 0 {
		errResponse := map[string]any{
			"errors": validationErr,
		}

		pkg.NewResponse(http.StatusUnprocessableEntity, constant.ErrValidation, errResponse, nil).HTTP(w)
		return
	}

	response := h.usecase.Register(r.Context(), &req)
	response.HTTP(w)
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req domain.LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		pkg.NewResponse(http.StatusBadRequest, constant.ErrInvalidJson, nil, nil).HTTP(w)
		return
	}

	// validation
	validationErr := validator.ValidateRequest(&req)
	if len(validationErr) > 0 {
		errResponse := map[string]any{
			"errors": validationErr,
		}

		pkg.NewResponse(http.StatusUnprocessableEntity, constant.ErrValidation, errResponse, nil).HTTP(w)
		return
	}

	req.RemoteAddr = r.RemoteAddr

	response := h.usecase.Login(r.Context(), &req)
	if response.Data != nil {
		data, ok := response.Data.(map[string]any)
		if !ok {
			h.logger.Printf("[ERROR] failed to convert data")
			pkg.NewResponse(http.StatusInternalServerError, constant.ErrServer, nil, nil).HTTP(w)

			return
		}

		h.setAuthCookies(w, fmt.Sprint(data["access_token"]), fmt.Sprint(data["refresh_token"]))

		delete(data, "access_token")
		delete(data, "refresh_token")
	}

	response.HTTP(w)
}

func (h *UserHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		pkg.NewResponse(http.StatusUnauthorized, constant.ErrInvalidToken, nil, nil).HTTP(w)
		return
	}

	if cookie.Value == "" {
		pkg.NewResponse(http.StatusUnauthorized, constant.ErrInvalidToken, nil, nil).HTTP(w)
		return
	}

	response := h.usecase.RefreshToken(r.Context(), cookie.Value, r.RemoteAddr)
	if response.Data != nil {
		data, ok := response.Data.(map[string]any)
		if !ok {
			h.logger.Printf("[ERROR] failed to convert data")
			pkg.NewResponse(http.StatusInternalServerError, constant.ErrServer, nil, nil).HTTP(w)

			return
		}

		h.setAuthCookies(w, fmt.Sprint(data["access_token"]), fmt.Sprint(data["refresh_token"]))

		delete(data, "access_token")
		delete(data, "refresh_token")
	}
	response.HTTP(w)
}

func (h *UserHandler) Profile(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("access_token")
	if err != nil {
		pkg.NewResponse(http.StatusUnauthorized, constant.ErrInvalidToken, nil, nil).HTTP(w)
		return
	}

	response := h.usecase.Profile(r.Context(), cookie.Value)

	response.HTTP(w)
}

func (h *UserHandler) Logout(w http.ResponseWriter, r *http.Request) {
	accessToken, err := r.Cookie("access_token")
	if err != nil {
		pkg.NewResponse(http.StatusOK, "Success", nil, nil).HTTP(w)
		return
	}

	refreshToken, err := r.Cookie("refresh_token")
	if err != nil {
		pkg.NewResponse(http.StatusOK, "Success", nil, nil).HTTP(w)
		return
	}

	if accessToken.Value == "" || refreshToken.Value == "" {
		pkg.NewResponse(http.StatusOK, "Success", nil, nil).HTTP(w)
		return
	}

	response := h.usecase.Logout(r.Context(), accessToken.Value, refreshToken.Value)

	h.setAuthCookies(w, "", "")

	response.HTTP(w)
}
