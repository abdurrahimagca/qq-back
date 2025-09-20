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
			if pathMatchesPublic(r.URL.Path, publicPath) {
				// Allow public access
				next.ServeHTTP(w, r)
				return
			}
		}

		// Apply auth middleware for protected routes
		m.authMiddleware.RequireAuth(next).ServeHTTP(w, r)
	})
}

func pathMatchesPublic(requestPath, publicPath string) bool {
	if publicPath == "" {
		return false
	}

	if requestPath == publicPath {
		return true
	}

	if publicPath == "/" {
		return strings.HasPrefix(requestPath, "/")
	}

	if strings.HasSuffix(publicPath, "/") {
		return strings.HasPrefix(requestPath, publicPath)
	}

	if strings.HasPrefix(requestPath, publicPath) {
		remainder := strings.TrimPrefix(requestPath, publicPath)
		if remainder == "" {
			return true
		}

		switch remainder[0] {
		case '/', '?', '#':
			return true
		}
	}

	return false
}
