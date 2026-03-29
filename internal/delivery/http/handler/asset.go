package handler

import (
	"log"
	"net/http"

	"github.com/fazriegi/fintrack-be/internal/delivery/http/middleware"
	"github.com/fazriegi/fintrack-be/internal/domain"
	"github.com/fazriegi/fintrack-be/internal/usecase"
	"github.com/fazriegi/fintrack-be/pkg"
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
}

func (h *AssetHandler) ListAsset(w http.ResponseWriter, r *http.Request) {
	var req domain.ListAssetRequest

	if err := pkg.ParseQueryParam(r, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response := h.usecase.ListAsset(r.Context(), &req)

	response.HTTP(w)
}
