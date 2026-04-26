package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/fazriegi/fintrack-be/internal/delivery/http/middleware"
	"github.com/fazriegi/fintrack-be/internal/domain"
	"github.com/fazriegi/fintrack-be/internal/usecase"
	"github.com/fazriegi/fintrack-be/pkg"
	"github.com/fazriegi/fintrack-be/pkg/constant"
	"github.com/fazriegi/fintrack-be/pkg/validator"
	"github.com/google/uuid"
)

type AssetHandler struct {
	usecase usecase.AssetUsecase
	logger  *log.Logger
}

func NewAssetHandler(mux *http.ServeMux, uc usecase.AssetUsecase, logger *log.Logger) {
	handler := &AssetHandler{
		usecase: uc,
		logger:  logger,
	}

	mux.Handle("GET /v1/assets", middleware.MiddlewareAuth(http.HandlerFunc(handler.ListAsset)))
	mux.Handle("GET /v1/assets/categories", middleware.MiddlewareAuth(http.HandlerFunc(handler.ListAssetCategory)))
	mux.Handle("GET /v1/assets/{id}", middleware.MiddlewareAuth(http.HandlerFunc(handler.GetByID)))
	mux.Handle("DELETE /v1/assets/{id}", middleware.MiddlewareAuth(http.HandlerFunc(handler.Delete)))
	mux.Handle("POST /v1/assets", middleware.MiddlewareAuth(http.HandlerFunc(handler.Create)))
}

func (h *AssetHandler) ListAsset(w http.ResponseWriter, r *http.Request) {
	var req domain.ListAssetRequest

	if err := pkg.ParseQueryParam(r, &req); err != nil {
		h.logger.Printf("[ERROR] parsing query params: %s", err.Error())
		pkg.NewResponse(http.StatusBadRequest, constant.ErrParseQueryParam, nil, nil).HTTP(w)
		return
	}

	response := h.usecase.ListAsset(r.Context(), &req)

	response.HTTP(w)
}

func (h *AssetHandler) ListAssetCategory(w http.ResponseWriter, r *http.Request) {
	response := h.usecase.ListAssetCategory(r.Context())
	response.HTTP(w)
}

func (h *AssetHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	parsedID, err := uuid.Parse(id)
	if err != nil {
		h.logger.Printf("[ERROR] uuid.Parse - invalid UUID format: %s", err.Error())
		pkg.NewResponse(http.StatusBadRequest, constant.ErrInvalidParam, nil, nil).HTTP(w)
		return
	}

	h.usecase.GetByID(r.Context(), parsedID).HTTP(w)
}

func (h *AssetHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	parsedID, err := uuid.Parse(id)
	if err != nil {
		h.logger.Printf("[ERROR] uuid.Parse - invalid UUID format: %s", err.Error())
		pkg.NewResponse(http.StatusBadRequest, constant.ErrInvalidParam, nil, nil).HTTP(w)
		return
	}

	h.usecase.Delete(r.Context(), parsedID).HTTP(w)
}

func (h *AssetHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateAsset

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

	response := h.usecase.Create(r.Context(), &req)
	response.HTTP(w)
}
