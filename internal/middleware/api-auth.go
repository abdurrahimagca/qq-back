package middleware

import (
	"net/http"
	"strings"

	"github.com/abdurrahimagca/qq-back/internal/config/environment"
)

func ApiAuth(env *environment.Config) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			apiKey := strings.TrimSpace(r.Header.Get("X-API-Key"))

			if len(apiKey) == 0 {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			if apiKey != env.APIKey {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
