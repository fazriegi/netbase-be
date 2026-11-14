package middleware

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequestSecurityBodyLimit(t *testing.T) {
	sec := RequestSecurity([]string{"http://localhost:5173"})

	handler := sec(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]string
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))

	// 1. Valid small JSON request
	smallBody := bytes.NewBufferString(`{"username":"test"}`)
	req1 := httptest.NewRequest(http.MethodPost, "/v1/login", smallBody)
	req1.Header.Set("Content-Type", "application/json")
	req1.Header.Set("Origin", "http://localhost:5173")
	w1 := httptest.NewRecorder()
	handler.ServeHTTP(w1, req1)
	if w1.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for valid body, got %d", w1.Code)
	}

	// 2. Oversized body (> 1 MB) with unknown ContentLength (-1) should fail decoding
	largeData := bytes.Repeat([]byte("a"), MaxRequestBodyBytes+100)
	req2 := httptest.NewRequest(http.MethodPost, "/v1/login", bytes.NewReader(largeData))
	req2.ContentLength = -1
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, req2)
	if w2.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 BadRequest when streaming oversized body, got %d", w2.Code)
	}

	// 3. Oversized declared Content-Length should be rejected with 413 immediately
	req3 := httptest.NewRequest(http.MethodPost, "/v1/login", bytes.NewReader(largeData))
	req3.Header.Set("Content-Type", "application/json")
	w3 := httptest.NewRecorder()
	handler.ServeHTTP(w3, req3)
	if w3.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected 413 RequestEntityTooLarge for declared oversized ContentLength, got %d", w3.Code)
	}
}

