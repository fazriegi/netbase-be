package pkg

import (
	"encoding/json"
	"net/http"
)

type Response struct {
	Status
	Data           any             `json:"data,omitempty"`
	PaginationMeta *PaginationMeta `json:"pagination_meta,omitempty"`
}

type Status struct {
	Code      int    `json:"code"`
	Status    string `json:"status"`
	Message   string `json:"message"`
	IsSuccess bool   `json:"is_success"`
}

type PaginationMeta struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

func buildStatus(code int, msg string) Status {
	statusText := http.StatusText(code)
	if statusText == "" {
		statusText = "Unknown Status"
	}

	if msg == "" {
		msg = statusText
	}

	return Status{
		Code:      code,
		Message:   msg,
		Status:    statusText,
		IsSuccess: code >= 200 && code <= 299,
	}
}

func NewResponse(code int, message string, data any, paginationMeta *PaginationMeta) Response {
	return Response{
		Status:         buildStatus(code, message),
		Data:           data,
		PaginationMeta: paginationMeta,
	}
}

func (r Response) HTTP(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(r.Status.Code)
	json.NewEncoder(w).Encode(r)
}
