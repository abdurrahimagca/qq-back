package middleware

import (
	"net/http"
	"strings"

	tokenport "github.com/abdurrahimagca/qq-back/internal/platform/token"
	"github.com/abdurrahimagca/qq-back/internal/user"
	"github.com/jackc/pgx/v5/pgtype"
)

type AuthMiddleware struct {
	tokenService tokenport.Service
	userService  user.Service
}

func NewAuthMiddleware(tokenService tokenport.Service, userService user.Service) *AuthMiddleware {
	return &AuthMiddleware{
		tokenService: tokenService,
		userService:  userService,
	}
}

func (m *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		// Extract Bearer token
		parts := strings.Fields(authHeader)
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
			return
		}

		token := strings.TrimSpace(parts[1])
		if token == "" {
			http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
			return
		}

		// Validate token
		tokenResult, err := m.tokenService.ValidateToken(r.Context(), tokenport.ValidateTokenParams{
			Token: token,
		})
		if err != nil {
			http.Error(w, "Invalid token: "+err.Error(), http.StatusUnauthorized)
			return
		}

		// Get user from database
		userID := tokenResult.Claims.UserID
		if userID == "" {
			http.Error(w, "Invalid token: missing user ID", http.StatusUnauthorized)
			return
		}

		var userUUID pgtype.UUID
		err = userUUID.Scan(userID)
		if err != nil {
			http.Error(w, "Invalid user ID: "+err.Error(), http.StatusUnauthorized)
			return
		}
		retrievedUser, userErr := m.userService.GetUserByID(r.Context(), userUUID)
		if userErr != nil {
			http.Error(w, "User not found: "+userErr.Error(), http.StatusUnauthorized)
			return
		}

		// Add user to context
		ctx := WithUser(r.Context(), retrievedUser)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *AuthMiddleware) OptionalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			// No auth provided, continue without user
			next.ServeHTTP(w, r)
			return
		}

		// Extract Bearer token
		parts := strings.Fields(authHeader)
		if len(parts) != 2 || parts[0] != "Bearer" {
			// Invalid format but optional, continue without user
			next.ServeHTTP(w, r)
			return
		}

		token := strings.TrimSpace(parts[1])
		if token == "" {
			// Empty token, continue without user
			next.ServeHTTP(w, r)
			return
		}

		// Try to validate token
		tokenResult, err := m.tokenService.ValidateToken(r.Context(), tokenport.ValidateTokenParams{
			Token: token,
		})
		if err != nil {
			// Invalid token but optional, continue without user
			next.ServeHTTP(w, r)
			return
		}

		// Get user from database
		userID := tokenResult.Claims.UserID
		if userID != "" {
			var userUUID pgtype.UUID
			err = userUUID.Scan(userID)
			if err != nil {
				http.Error(w, "Invalid user ID: "+err.Error(), http.StatusUnauthorized)
				return
			}
			retrievedUser, userErr := m.userService.GetUserByID(r.Context(), userUUID)
			if userErr != nil {
				http.Error(w, "User not found: "+userErr.Error(), http.StatusUnauthorized)
				return
			}

			// Add user to context
			ctx := WithUser(r.Context(), retrievedUser)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		// Continue without user if anything fails
		next.ServeHTTP(w, r)
	})
}
