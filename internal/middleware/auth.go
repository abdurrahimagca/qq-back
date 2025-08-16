package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/abdurrahimagca/qq-back/internal/config/environment"
	"github.com/abdurrahimagca/qq-back/internal/service/auth"
)

type userContextKey string

const UserContextKey userContextKey = "user"

func UserAuth(env *environment.Config) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tx, ok := GetTxFromContext(r.Context())
			if !ok {
				http.Error(w, "Transaction not found in context", http.StatusInternalServerError)
				return
			}

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error":"Authorization header is required","code":"MISSING_AUTH_HEADER"}`))
				return
			}

			if !strings.HasPrefix(authHeader, "Bearer ") {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error":"Authorization header must start with 'Bearer '","code":"INVALID_AUTH_FORMAT"}`))
				return
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString == "" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error":"Token cannot be empty","code":"EMPTY_TOKEN"}`))
				return
			}

			user, err := auth.ValidateAndGetUserFromAccessToken(tokenString, env, tx)
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error":"` + err.Error() + `","code":"TOKEN_VALIDATION_FAILED"}`))
				return
			}

			ctx := context.WithValue(r.Context(), UserContextKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
