package handler

import (
	"log"
	"net/http"

	"github.com/fazriegi/netbase-be/internal/delivery/http/middleware"
	"github.com/fazriegi/netbase-be/internal/usecase"
)

type NetworthHandler struct {
	usecase usecase.NetworthUsecase
	logger  *log.Logger
}

func NewNetworthHandler(mux *http.ServeMux, uc usecase.NetworthUsecase, logger *log.Logger) {
	handler := &NetworthHandler{
		usecase: uc,
		logger:  logger,
	}

	mux.Handle("GET /v1/net-worth/current", middleware.MiddlewareAuth(http.HandlerFunc(handler.GetCurrent)))
}

func (h *NetworthHandler) GetCurrent(w http.ResponseWriter, r *http.Request) {
	h.usecase.GetCurrent(r.Context()).HTTP(w)
}
