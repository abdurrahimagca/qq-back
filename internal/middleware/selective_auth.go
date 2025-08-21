package middleware

import (
	"net/http"
	"strings"
)

type SelectiveAuthMiddleware struct {
	authMiddleware *AuthMiddleware
	publicPaths    []string
}

func NewSelectiveAuthMiddleware(authMiddleware *AuthMiddleware, publicPaths []string) *SelectiveAuthMiddleware {
	return &SelectiveAuthMiddleware{
		authMiddleware: authMiddleware,
		publicPaths:    publicPaths,
	}
}

func (m *SelectiveAuthMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the path should be public
		for _, publicPath := range m.publicPaths {
			if strings.HasPrefix(r.URL.Path, publicPath) {
				// Allow public access
				next.ServeHTTP(w, r)
				return
			}
		}
		
		// Apply auth middleware for protected routes
		m.authMiddleware.RequireAuth(next).ServeHTTP(w, r)
	})
}