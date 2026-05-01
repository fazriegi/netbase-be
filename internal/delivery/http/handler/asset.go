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
	mux.Handle("PUT /v1/assets/{id}", middleware.MiddlewareAuth(http.HandlerFunc(handler.Update)))
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

	detailValidationErr := validateAssetDetails(req.CategoryType, req.Details)
	if len(detailValidationErr) > 0 {
		validationErr = append(validationErr, detailValidationErr...)
	}

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

func (h *AssetHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	parsedID, err := uuid.Parse(id)
	if err != nil {
		h.logger.Printf("[ERROR] uuid.Parse - invalid UUID format: %s", err.Error())
		pkg.NewResponse(http.StatusBadRequest, constant.ErrInvalidParam, nil, nil).HTTP(w)
		return
	}

	var req domain.UpdateAsset

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		pkg.NewResponse(http.StatusBadRequest, constant.ErrInvalidJson, nil, nil).HTTP(w)
		return
	}

	// validation
	validationErr := validator.ValidateRequest(&req)

	detailValidationErr := validateAssetDetails(req.CategoryType, req.Details)
	if len(detailValidationErr) > 0 {
		validationErr = append(validationErr, detailValidationErr...)
	}

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

func validateAssetDetails(categoryType string, details any) []validator.ValidationErrResponse {
	var detailErrors []validator.ValidationErrResponse

	detailsBytes, err := json.Marshal(details)
	if err != nil {
		return append(detailErrors, validator.ValidationErrResponse{
			FailedField: "details",
			Tag:         "invalid_format",
			TagValue:    "",
		})
	}

	switch categoryType {
	case "liquid":
		var detail domain.LiquidAsset
		if err := json.Unmarshal(detailsBytes, &detail); err != nil {
			return append(detailErrors, validator.ValidationErrResponse{
				FailedField: "details",
				Tag:         "invalid_format",
				TagValue:    "",
			})
		}
		detailErrors = validator.ValidateRequest(&detail)
	case "investment":
		var detail domain.InvestmentAsset
		if err := json.Unmarshal(detailsBytes, &detail); err != nil {
			return append(detailErrors, validator.ValidationErrResponse{
				FailedField: "details",
				Tag:         "invalid_format",
				TagValue:    "",
			})
		}
		detailErrors = validator.ValidateRequest(&detail)
	case "physical":
		var detail domain.PhysicalAsset
		if err := json.Unmarshal(detailsBytes, &detail); err != nil {
			return append(detailErrors, validator.ValidationErrResponse{
				FailedField: "details",
				Tag:         "invalid_format",
				TagValue:    "",
			})
		}
		detailErrors = validator.ValidateRequest(&detail)
	default:
		detailErrors = append(detailErrors, validator.ValidationErrResponse{
			FailedField: "category_type",
			Tag:         "invalid",
			TagValue:    categoryType,
		})
	}

	for i := range detailErrors {
		detailErrors[i].FailedField = "details." + detailErrors[i].FailedField
	}

	return detailErrors
}
