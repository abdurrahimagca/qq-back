package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/abdurrahimagca/qq-back/internal/api"
	"github.com/abdurrahimagca/qq-back/internal/ports"
	"github.com/abdurrahimagca/qq-back/internal/user"
)

type StrictAuthMiddleware struct {
	tokenService ports.TokenPort
	userService  user.Service
	publicPaths  []string
}

func NewStrictAuthMiddleware(tokenService ports.TokenPort, userService user.Service, publicPaths []string) *StrictAuthMiddleware {
	return &StrictAuthMiddleware{
		tokenService: tokenService,
		userService:  userService,
		publicPaths:  publicPaths,
	}
}

func (m *StrictAuthMiddleware) Middleware(f api.StrictHandlerFunc, operationID string) api.StrictHandlerFunc {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request, request interface{}) (response interface{}, err error) {
		// Check if this is a public endpoint
		for _, publicPath := range m.publicPaths {
			if strings.Contains(r.URL.Path, publicPath) || operationID == "PostAuthOtp" || operationID == "PostAuthOtpVerify" || operationID == "PostAuthRefreshToken" {
				return f(ctx, w, r, request)
			}
		}

		// Extract and validate auth token
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return map[string]interface{}{
				"error":   "Authorization header required",
				"success": false,
			}, nil
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			w.WriteHeader(http.StatusUnauthorized)
			return map[string]interface{}{
				"error":   "Invalid authorization header format", 
				"success": false,
			}, nil
		}

		token := parts[1]
		tokenResult, err := m.tokenService.ValidateToken(ctx, ports.ValidateTokenParams{
			Token: token,
		})
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return map[string]interface{}{
				"error":   "Invalid token: " + err.Error(),
				"success": false,
			}, nil
		}

		userID := tokenResult.Claims.UserID
		if userID == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return map[string]interface{}{
				"error":   "Invalid token: missing user ID",
				"success": false,
			}, nil
		}

		user, err := m.userService.GetUserByID(ctx, userID)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return map[string]interface{}{
				"error":   "User not found: " + err.Error(),
				"success": false,
			}, nil
		}

		// Add user to context
		ctx = WithUser(ctx, &user)
		
		return f(ctx, w, r, request)
	}
}