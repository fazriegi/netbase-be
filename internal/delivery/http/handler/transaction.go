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
	"github.com/google/uuid"
)

type TransactionHandler struct {
	usecase usecase.TransactionUsecase
	logger  *log.Logger
}

func NewTransactionHandler(mux *http.ServeMux, uc usecase.TransactionUsecase, logger *log.Logger) {
	h := &TransactionHandler{
		usecase: uc,
		logger:  logger,
	}

	mux.Handle("GET /v1/transactions", middleware.MiddlewareAuth(http.HandlerFunc(h.List)))
	mux.Handle("GET /v1/transactions/summary", middleware.MiddlewareAuth(http.HandlerFunc(h.GetSummary)))
	mux.Handle("GET /v1/transactions/{id}", middleware.MiddlewareAuth(http.HandlerFunc(h.GetByID)))
	mux.Handle("POST /v1/transactions", middleware.MiddlewareAuth(http.HandlerFunc(h.Create)))
	mux.Handle("PUT /v1/transactions/{id}", middleware.MiddlewareAuth(http.HandlerFunc(h.Update)))
	mux.Handle("DELETE /v1/transactions/{id}", middleware.MiddlewareAuth(http.HandlerFunc(h.Delete)))

	mux.Handle("GET /v1/transactions/categories", middleware.MiddlewareAuth(http.HandlerFunc(h.ListCategory)))
	mux.Handle("POST /v1/transactions/categories", middleware.MiddlewareAuth(http.HandlerFunc(h.CreateCategory)))
	mux.Handle("DELETE /v1/transactions/categories/{id}", middleware.MiddlewareAuth(http.HandlerFunc(h.DeleteCategory)))
}

func (h *TransactionHandler) List(w http.ResponseWriter, r *http.Request) {
	var req domain.ListTransactionRequest

	if err := pkg.ParseQueryParam(r, &req); err != nil {
		h.logger.Printf("[ERROR] parsing query params: %s", err.Error())
		pkg.NewResponse(http.StatusBadRequest, constant.ErrParseQueryParam, nil, nil).HTTP(w)
		return
	}

	h.usecase.List(r.Context(), &req).HTTP(w)
}

func (h *TransactionHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	var req domain.ListTransactionRequest

	if err := pkg.ParseQueryParam(r, &req); err != nil {
		h.logger.Printf("[ERROR] parsing query params: %s", err.Error())
		pkg.NewResponse(http.StatusBadRequest, constant.ErrParseQueryParam, nil, nil).HTTP(w)
		return
	}

	h.usecase.GetSummary(r.Context(), &req).HTTP(w)
}

func (h *TransactionHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	parsedID, err := uuid.Parse(id)
	if err != nil {
		h.logger.Printf("[ERROR] uuid.Parse - invalid UUID format: %s", err.Error())
		pkg.NewResponse(http.StatusBadRequest, constant.ErrInvalidParam, nil, nil).HTTP(w)
		return
	}

	h.usecase.GetByID(r.Context(), parsedID).HTTP(w)
}

func (h *TransactionHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateTransaction

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		pkg.NewResponse(http.StatusBadRequest, constant.ErrInvalidJson, nil, nil).HTTP(w)
		return
	}

	validationErr := validator.ValidateRequest(&req)
	if len(validationErr) > 0 {
		errResponse := map[string]any{
			"errors": validationErr,
		}
		pkg.NewResponse(http.StatusUnprocessableEntity, constant.ErrValidation, errResponse, nil).HTTP(w)
		return
	}

	h.usecase.Create(r.Context(), &req).HTTP(w)
}

func (h *TransactionHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	parsedID, err := uuid.Parse(id)
	if err != nil {
		h.logger.Printf("[ERROR] uuid.Parse - invalid UUID format: %s", err.Error())
		pkg.NewResponse(http.StatusBadRequest, constant.ErrInvalidParam, nil, nil).HTTP(w)
		return
	}

	var req domain.CreateTransaction
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		pkg.NewResponse(http.StatusBadRequest, constant.ErrInvalidJson, nil, nil).HTTP(w)
		return
	}

	validationErr := validator.ValidateRequest(&req)
	if len(validationErr) > 0 {
		errResponse := map[string]any{
			"errors": validationErr,
		}
		pkg.NewResponse(http.StatusUnprocessableEntity, constant.ErrValidation, errResponse, nil).HTTP(w)
		return
	}

	req.ID = parsedID
	h.usecase.Update(r.Context(), &req).HTTP(w)
}

func (h *TransactionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	parsedID, err := uuid.Parse(id)
	if err != nil {
		h.logger.Printf("[ERROR] uuid.Parse - invalid UUID format: %s", err.Error())
		pkg.NewResponse(http.StatusBadRequest, constant.ErrInvalidParam, nil, nil).HTTP(w)
		return
	}

	h.usecase.Delete(r.Context(), parsedID).HTTP(w)
}

func (h *TransactionHandler) ListCategory(w http.ResponseWriter, r *http.Request) {
	var req domain.ListCategoryRequest

	if err := pkg.ParseQueryParam(r, &req); err != nil {
		h.logger.Printf("[ERROR] parsing query params: %s", err.Error())
		pkg.NewResponse(http.StatusBadRequest, constant.ErrParseQueryParam, nil, nil).HTTP(w)
		return
	}

	h.usecase.ListCategory(r.Context(), &req).HTTP(w)
}

func (h *TransactionHandler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	var req domain.Category

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		pkg.NewResponse(http.StatusBadRequest, constant.ErrInvalidJson, nil, nil).HTTP(w)
		return
	}

	validationErr := validator.ValidateRequest(&req)
	if len(validationErr) > 0 {
		errResponse := map[string]any{
			"errors": validationErr,
		}
		pkg.NewResponse(http.StatusUnprocessableEntity, constant.ErrValidation, errResponse, nil).HTTP(w)
		return
	}

	h.usecase.CreateCategory(r.Context(), &req).HTTP(w)
}

func (h *TransactionHandler) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	parsedID, err := uuid.Parse(id)
	if err != nil {
		h.logger.Printf("[ERROR] uuid.Parse - invalid UUID format: %s", err.Error())
		pkg.NewResponse(http.StatusBadRequest, constant.ErrInvalidParam, nil, nil).HTTP(w)
		return
	}

	h.usecase.DeleteCategory(r.Context(), parsedID).HTTP(w)
}
