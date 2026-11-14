package middleware

import (
	"context"
	"errors"
	"io"
	"mime"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/fazriegi/netbase-be/pkg"
	"github.com/fazriegi/netbase-be/pkg/constant"
	"github.com/fazriegi/netbase-be/pkg/token"
	"github.com/google/uuid"
)

const MaxRequestBodyBytes = 1 << 20 // 1 MB

var errRequestBodyTooLarge = errors.New("http: request body too large")

// bodyLimitReadCloser caps request body reads without touching http.ResponseWriter headers,
// avoiding the net/http MaxBytesReader header race (GO-2026-6859).
type bodyLimitReadCloser struct {
	rc        io.ReadCloser
	remaining int64
}

func (b *bodyLimitReadCloser) Read(p []byte) (int, error) {
	if b.remaining <= 0 {
		var scratch [1]byte
		n, err := b.rc.Read(scratch[:])
		if n > 0 {
			return 0, errRequestBodyTooLarge
		}
		return 0, err
	}
	if int64(len(p)) > b.remaining {
		p = p[:b.remaining]
	}
	n, err := b.rc.Read(p)
	b.remaining -= int64(n)
	return n, err
}

func (b *bodyLimitReadCloser) Close() error {
	return b.rc.Close()
}

func MiddlewareAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("access_token")
		if err != nil || cookie.Value == "" {
			pkg.NewResponse(http.StatusUnauthorized, constant.ErrInvalidToken, nil, nil).HTTP(w)
			return
		}

		claims, err := token.ValidateAccessToken(cookie.Value)
		if err != nil {
			pkg.NewResponse(http.StatusUnauthorized, constant.ErrInvalidToken, nil, nil).HTTP(w)
			return
		}

		parsedUserID, err := uuid.Parse(claims.UserID)
		if err != nil {
			pkg.NewResponse(http.StatusUnauthorized, constant.ErrInvalidToken, nil, nil).HTTP(w)
			return
		}

		ctx := context.WithValue(r.Context(), "user_id", parsedUserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequestSecurity enforces request body size limits and CSRF / Content-Type checks on state-changing methods.
func RequestSecurity(allowedOrigins []string) func(http.Handler) http.Handler {
	originSet := make(map[string]struct{}, len(allowedOrigins))
	for _, o := range allowedOrigins {
		trimmed := strings.TrimSpace(o)
		if trimmed != "" && trimmed != "*" {
			originSet[strings.TrimRight(trimmed, "/")] = struct{}{}
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.ContentLength > MaxRequestBodyBytes {
				pkg.NewResponse(http.StatusRequestEntityTooLarge, constant.ErrInvalidRequest, nil, nil).HTTP(w)
				return
			}
			if r.Body != nil {
				r.Body = &bodyLimitReadCloser{
					rc:        r.Body,
					remaining: MaxRequestBodyBytes,
				}
			}

			switch r.Method {
			case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
				if origin := r.Header.Get("Origin"); origin != "" && len(originSet) > 0 {
					normalized := strings.TrimRight(origin, "/")
					if _, ok := originSet[normalized]; !ok {
						pkg.NewResponse(http.StatusForbidden, constant.ErrNotAuthorized, nil, nil).HTTP(w)
						return
					}
				}

				if r.ContentLength > 0 {
					ct := r.Header.Get("Content-Type")
					mediaType, _, err := mime.ParseMediaType(ct)
					if err != nil || mediaType != "application/json" {
						pkg.NewResponse(http.StatusUnsupportedMediaType, constant.ErrInvalidJson, nil, nil).HTTP(w)
						return
					}
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

type ipVisitor struct {
	count       int
	windowStart time.Time
}

var (
	visitorsMu sync.Mutex
	visitors   = make(map[string]*ipVisitor)
)

// RateLimitAuth limits authentication endpoints to maxRequests per window per IP address.
func RateLimitAuth(maxRequests int, window time.Duration, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil || ip == "" {
			ip = r.RemoteAddr
		}

		now := time.Now()
		visitorsMu.Lock()
		if len(visitors) > 10000 {
			for k, v := range visitors {
				if now.Sub(v.windowStart) > window {
					delete(visitors, k)
				}
			}
		}

		v, exists := visitors[ip]
		if !exists || now.Sub(v.windowStart) > window {
			visitors[ip] = &ipVisitor{count: 1, windowStart: now}
			visitorsMu.Unlock()
			next(w, r)
			return
		}

		if v.count >= maxRequests {
			visitorsMu.Unlock()
			w.Header().Set("Retry-After", "60")
			pkg.NewResponse(http.StatusTooManyRequests, "too many requests, please try again later", nil, nil).HTTP(w)
			return
		}

		v.count++
		visitorsMu.Unlock()
		next(w, r)
	}
}
