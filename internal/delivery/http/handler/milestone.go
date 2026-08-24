package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/fazriegi/netbase-be/internal/delivery/http/middleware"
	"github.com/fazriegi/netbase-be/internal/domain"
	"github.com/fazriegi/netbase-be/internal/usecase"
	"github.com/fazriegi/netbase-be/pkg"
	"github.com/fazriegi/netbase-be/pkg/constant"
	"github.com/fazriegi/netbase-be/pkg/validator"
)

type MilestoneHandler struct {
	usecase usecase.MilestoneUsecase
	logger  *log.Logger
}

func NewMilestoneHandler(mux *http.ServeMux, uc usecase.MilestoneUsecase, logger *log.Logger) {
	handler := &MilestoneHandler{
		usecase: uc,
		logger:  logger,
	}

	mux.Handle("POST /v1/milestones", middleware.MiddlewareAuth(http.HandlerFunc(handler.Create)))
}

func (h *MilestoneHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateMilestoneRequest

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
