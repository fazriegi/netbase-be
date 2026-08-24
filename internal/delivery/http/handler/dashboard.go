package handler

import (
	"log"
	"net/http"

	"github.com/fazriegi/netbase-be/internal/delivery/http/middleware"
	"github.com/fazriegi/netbase-be/internal/domain"
	"github.com/fazriegi/netbase-be/internal/usecase"
	"github.com/fazriegi/netbase-be/pkg"
	"github.com/fazriegi/netbase-be/pkg/constant"
	"github.com/fazriegi/netbase-be/pkg/validator"
)

type DashboardHandler struct {
	usecase usecase.DashboardUsecase
	logger  *log.Logger
}

func NewDashboardHandler(mux *http.ServeMux, uc usecase.DashboardUsecase, logger *log.Logger) {
	handler := &DashboardHandler{
		usecase: uc,
		logger:  logger,
	}

	mux.Handle("GET /v1/dashboard/cashflow", middleware.MiddlewareAuth(http.HandlerFunc(handler.GetCashflow)))
	mux.Handle("GET /v1/dashboard/milestone", middleware.MiddlewareAuth(http.HandlerFunc(handler.GetActiveMilestone)))
	mux.Handle("GET /v1/dashboard/networth", middleware.MiddlewareAuth(http.HandlerFunc(handler.GetNetworthSummary)))
	mux.Handle("GET /v1/dashboard/networth/history", middleware.MiddlewareAuth(http.HandlerFunc(handler.GetNetworthHistory)))
}

func (h *DashboardHandler) GetCashflow(w http.ResponseWriter, r *http.Request) {
	var req domain.DashboardCashflowRequest

	if err := pkg.ParseQueryParam(r, &req); err != nil {
		h.logger.Printf("[ERROR] parsing query params: %s", err.Error())
		pkg.NewResponse(http.StatusBadRequest, constant.ErrParseQueryParam, nil, nil).HTTP(w)
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

	h.usecase.GetCashflow(r.Context(), req.Period).HTTP(w)
}

func (h *DashboardHandler) GetActiveMilestone(w http.ResponseWriter, r *http.Request) {
	h.usecase.GetActiveMilestone(r.Context()).HTTP(w)
}

func (h *DashboardHandler) GetNetworthSummary(w http.ResponseWriter, r *http.Request) {
	h.usecase.GetNetworthSummary(r.Context()).HTTP(w)
}

func (h *DashboardHandler) GetNetworthHistory(w http.ResponseWriter, r *http.Request) {
	var req domain.DashboardNetworthHistoryRequest

	if err := pkg.ParseQueryParam(r, &req); err != nil {
		h.logger.Printf("[ERROR] parsing query params: %s", err.Error())
		pkg.NewResponse(http.StatusBadRequest, constant.ErrParseQueryParam, nil, nil).HTTP(w)
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

	h.usecase.GetNetworthHistory(r.Context(), &req).HTTP(w)
}
