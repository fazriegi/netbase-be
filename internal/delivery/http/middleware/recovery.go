package middleware

import (
	"log"
	"net/http"
	"runtime/debug"

	"github.com/fazriegi/netbase-be/pkg"
	"github.com/fazriegi/netbase-be/pkg/constant"
)

func Recovery(logger *log.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					logger.Printf("[PANIC RECOVERED] %v\nStack trace:\n%s", err, debug.Stack())
					pkg.NewResponse(http.StatusInternalServerError, constant.ErrServer, nil, nil).HTTP(w)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
