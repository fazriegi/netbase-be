package middleware

import (
	"context"
	"net/http"

	"github.com/fazriegi/fintrack-be/pkg"
	"github.com/fazriegi/fintrack-be/pkg/constant"
	"github.com/fazriegi/fintrack-be/pkg/token"
)

func MiddlewareAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("access_token")
		if err != nil || cookie.Value == "" {
			pkg.NewResponse(http.StatusUnauthorized, constant.ErrInvalidToken, nil, nil).HTTP(w)
			return
		}

		claims, err := token.ValidateToken(cookie.Value)
		if err != nil {
			pkg.NewResponse(http.StatusUnauthorized, constant.ErrInvalidToken, nil, nil).HTTP(w)
			return
		}

		ctx := context.WithValue(r.Context(), "user_id", claims.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
