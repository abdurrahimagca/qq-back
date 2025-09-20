package middleware_test

import (
	"net/http/httptest"
	"testing"

	"github.com/abdurrahimagca/qq-back/internal/middleware"
	"github.com/stretchr/testify/assert"
)

func TestNewSelectiveAuthMiddleware(t *testing.T) {
	tokenService := NewMockTokenService()
	userService := NewMockUserService()
	authMiddleware := middleware.NewAuthMiddleware(tokenService, userService)
	publicPaths := []string{"/health", "/public"}

	selectiveAuth := middleware.NewSelectiveAuthMiddleware(authMiddleware, publicPaths)

	assert.NotNil(t, selectiveAuth, "SelectiveAuthMiddleware should not be nil")
}

func TestSelectiveAuthMiddleware_PublicPath_ExactMatch(t *testing.T) {
	tokenService := NewMockTokenService()
	userService := NewMockUserService()
	authMiddleware := middleware.NewAuthMiddleware(tokenService, userService)
	publicPaths := []string{"/health", "/public"}

	selectiveAuth := middleware.NewSelectiveAuthMiddleware(authMiddleware, publicPaths)
	handler := NewTestHandler()
	protectedHandler := selectiveAuth.Handler(handler)

	tests := []struct {
		name string
		path string
	}{
		{
			name: "Health endpoint",
			path: "/health",
		},
		{
			name: "Public endpoint",
			path: "/public",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler.Reset()
			req := createTestRequest(tt.path, "")
			w := httptest.NewRecorder()

			protectedHandler.ServeHTTP(w, req)

			// Should allow access without authentication
			assertOK(t, w)
			assert.True(t, handler.WasCalled(), "Next handler should be called")
			assertNoUserInContext(t, handler.GetRequest())

			// Verify no service calls
			assert.Equal(t, 0, tokenService.GetValidateTokenCallCount(), "Token service should not be called")
			assert.Equal(t, 0, userService.GetGetUserByIDCallCount(), "User service should not be called")
		})
	}
}

func TestSelectiveAuthMiddleware_PublicPath_PrefixMatch(t *testing.T) {
	tokenService := NewMockTokenService()
	userService := NewMockUserService()
	authMiddleware := middleware.NewAuthMiddleware(tokenService, userService)
	publicPaths := []string{"/api/public", "/docs"}

	selectiveAuth := middleware.NewSelectiveAuthMiddleware(authMiddleware, publicPaths)
	handler := NewTestHandler()
	protectedHandler := selectiveAuth.Handler(handler)

	tests := []struct {
		name string
		path string
	}{
		{
			name: "Public API endpoint",
			path: "/api/public/users",
		},
		{
			name: "Public API nested",
			path: "/api/public/health/check",
		},
		{
			name: "Docs root",
			path: "/docs",
		},
		{
			name: "Docs nested",
			path: "/docs/openapi.json",
		},
		{
			name: "Docs deep nested",
			path: "/docs/v1/swagger/index.html",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler.Reset()
			tokenService.Reset()
			userService.Reset()

			req := createTestRequest(tt.path, "")
			w := httptest.NewRecorder()

			protectedHandler.ServeHTTP(w, req)

			// Should allow access without authentication
			assertOK(t, w)
			assert.True(t, handler.WasCalled(), "Next handler should be called")
			assertNoUserInContext(t, handler.GetRequest())

			// Verify no service calls
			assert.Equal(t, 0, tokenService.GetValidateTokenCallCount(), "Token service should not be called")
			assert.Equal(t, 0, userService.GetGetUserByIDCallCount(), "User service should not be called")
		})
	}
}

func TestSelectiveAuthMiddleware_ProtectedPath_RequiresAuth(t *testing.T) {
	tokenService := NewMockTokenService()
	userService := NewMockUserService()
	authMiddleware := middleware.NewAuthMiddleware(tokenService, userService)
	publicPaths := []string{"/health", "/public"}

	selectiveAuth := middleware.NewSelectiveAuthMiddleware(authMiddleware, publicPaths)
	handler := NewTestHandler()
	protectedHandler := selectiveAuth.Handler(handler)

	tests := []struct {
		name string
		path string
	}{
		{
			name: "Private API",
			path: "/api/private",
		},
		{
			name: "User endpoint",
			path: "/users/123",
		},
		{
			name: "Admin endpoint",
			path: "/admin/dashboard",
		},
		{
			name: "Root path",
			path: "/",
		},
		{
			name: "Similar to public but not exact",
			path: "/publics",
		},
		{
			name: "Similar to public but different",
			path: "/public-but-not",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler.Reset()
			tokenService.Reset()
			userService.Reset()

			req := createTestRequest(tt.path, "")
			w := httptest.NewRecorder()

			protectedHandler.ServeHTTP(w, req)

			// Should require authentication and fail without token
			assertUnauthorized(t, w)
			assert.False(t, handler.WasCalled(), "Next handler should not be called")
			assert.Contains(t, w.Body.String(), "Authorization header required")

			// Verify no service calls (auth fails before token validation)
			assert.Equal(t, 0, tokenService.GetValidateTokenCallCount(), "Token service should not be called")
			assert.Equal(t, 0, userService.GetGetUserByIDCallCount(), "User service should not be called")
		})
	}
}

func TestSelectiveAuthMiddleware_ProtectedPath_WithValidAuth(t *testing.T) {
	tokenService := NewMockTokenService()
	userService := NewMockUserService()
	authMiddleware := middleware.NewAuthMiddleware(tokenService, userService)
	publicPaths := []string{"/health", "/public"}

	// Setup successful authentication
	user := createTestUser(TestUserID1)
	tokenService.SetValidateTokenResult(TestUserID1, nil)
	userService.SetGetUserByIDResult(user, nil)

	selectiveAuth := middleware.NewSelectiveAuthMiddleware(authMiddleware, publicPaths)
	handler := NewTestHandler()
	protectedHandler := selectiveAuth.Handler(handler)

	req := createTestRequest("/api/private", createValidToken(TestUserID1))
	w := httptest.NewRecorder()

	protectedHandler.ServeHTTP(w, req)

	// Should allow access with valid authentication
	assertOK(t, w)
	assert.True(t, handler.WasCalled(), "Next handler should be called")
	assertUserInContext(t, handler.GetRequest(), user)

	// Verify service calls
	assert.Equal(t, 1, tokenService.GetValidateTokenCallCount(), "Token service should be called")
	assert.Equal(t, 1, userService.GetGetUserByIDCallCount(), "User service should be called")
}

func TestSelectiveAuthMiddleware_EmptyPublicPaths(t *testing.T) {
	tokenService := NewMockTokenService()
	userService := NewMockUserService()
	authMiddleware := middleware.NewAuthMiddleware(tokenService, userService)
	publicPaths := []string{} // No public paths

	selectiveAuth := middleware.NewSelectiveAuthMiddleware(authMiddleware, publicPaths)
	handler := NewTestHandler()
	protectedHandler := selectiveAuth.Handler(handler)

	tests := []string{"/health", "/public", "/", "/api"}

	for _, path := range tests {
		t.Run("Path "+path, func(t *testing.T) {
			handler.Reset()
			tokenService.Reset()
			userService.Reset()

			req := createTestRequest(path, "")
			w := httptest.NewRecorder()

			protectedHandler.ServeHTTP(w, req)

			// All paths should require authentication
			assertUnauthorized(t, w)
			assert.False(t, handler.WasCalled(), "Next handler should not be called")
			assert.Contains(t, w.Body.String(), "Authorization header required")
		})
	}
}

func TestSelectiveAuthMiddleware_NilPublicPaths(t *testing.T) {
	tokenService := NewMockTokenService()
	userService := NewMockUserService()
	authMiddleware := middleware.NewAuthMiddleware(tokenService, userService)
	var publicPaths []string = nil // Nil public paths

	selectiveAuth := middleware.NewSelectiveAuthMiddleware(authMiddleware, publicPaths)
	handler := NewTestHandler()
	protectedHandler := selectiveAuth.Handler(handler)

	req := createTestRequest("/health", "")
	w := httptest.NewRecorder()

	protectedHandler.ServeHTTP(w, req)

	// Should require authentication (no public paths)
	assertUnauthorized(t, w)
	assert.False(t, handler.WasCalled(), "Next handler should not be called")
	assert.Contains(t, w.Body.String(), "Authorization header required")
}

func TestSelectiveAuthMiddleware_RootPathPublic(t *testing.T) {
	tokenService := NewMockTokenService()
	userService := NewMockUserService()
	authMiddleware := middleware.NewAuthMiddleware(tokenService, userService)
	publicPaths := []string{"/"} // Root path is public

	selectiveAuth := middleware.NewSelectiveAuthMiddleware(authMiddleware, publicPaths)
	handler := NewTestHandler()
	protectedHandler := selectiveAuth.Handler(handler)

	tests := []string{"/", "/health", "/api", "/users", "/admin"}

	for _, path := range tests {
		t.Run("Path "+path, func(t *testing.T) {
			handler.Reset()
			tokenService.Reset()
			userService.Reset()

			req := createTestRequest(path, "")
			w := httptest.NewRecorder()

			protectedHandler.ServeHTTP(w, req)

			// All paths should be public (prefix match with "/")
			assertOK(t, w)
			assert.True(t, handler.WasCalled(), "Next handler should be called")
			assertNoUserInContext(t, handler.GetRequest())

			// Verify no service calls
			assert.Equal(t, 0, tokenService.GetValidateTokenCallCount(), "Token service should not be called")
			assert.Equal(t, 0, userService.GetGetUserByIDCallCount(), "User service should not be called")
		})
	}
}

func TestSelectiveAuthMiddleware_OverlappingPaths(t *testing.T) {
	tokenService := NewMockTokenService()
	userService := NewMockUserService()
	authMiddleware := middleware.NewAuthMiddleware(tokenService, userService)
	publicPaths := []string{"/api", "/api/public", "/api/public/health"}

	selectiveAuth := middleware.NewSelectiveAuthMiddleware(authMiddleware, publicPaths)
	handler := NewTestHandler()
	protectedHandler := selectiveAuth.Handler(handler)

	tests := []struct {
		name   string
		path   string
		public bool
	}{
		{
			name:   "First matching prefix",
			path:   "/api/private",
			public: true, // Matches "/api"
		},
		{
			name:   "More specific path",
			path:   "/api/public/users",
			public: true, // Matches "/api" (first match)
		},
		{
			name:   "Non-matching path",
			path:   "/admin/dashboard",
			public: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler.Reset()
			tokenService.Reset()
			userService.Reset()

			req := createTestRequest(tt.path, "")
			w := httptest.NewRecorder()

			protectedHandler.ServeHTTP(w, req)

			if tt.public {
				assertOK(t, w)
				assert.True(t, handler.WasCalled(), "Next handler should be called for public path")
				assertNoUserInContext(t, handler.GetRequest())
			} else {
				assertUnauthorized(t, w)
				assert.False(t, handler.WasCalled(), "Next handler should not be called for protected path")
			}
		})
	}
}

func TestSelectiveAuthMiddleware_CaseSensitivity(t *testing.T) {
	tokenService := NewMockTokenService()
	userService := NewMockUserService()
	authMiddleware := middleware.NewAuthMiddleware(tokenService, userService)
	publicPaths := []string{"/public"}

	selectiveAuth := middleware.NewSelectiveAuthMiddleware(authMiddleware, publicPaths)
	handler := NewTestHandler()
	protectedHandler := selectiveAuth.Handler(handler)

	tests := []struct {
		name   string
		path   string
		public bool
	}{
		{
			name:   "Exact case match",
			path:   "/public",
			public: true,
		},
		{
			name:   "Different case",
			path:   "/Public",
			public: false, // Case sensitive
		},
		{
			name:   "All uppercase",
			path:   "/PUBLIC",
			public: false, // Case sensitive
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler.Reset()
			tokenService.Reset()
			userService.Reset()

			req := createTestRequest(tt.path, "")
			w := httptest.NewRecorder()

			protectedHandler.ServeHTTP(w, req)

			if tt.public {
				assertOK(t, w)
				assert.True(t, handler.WasCalled(), "Next handler should be called for public path")
			} else {
				assertUnauthorized(t, w)
				assert.False(t, handler.WasCalled(), "Next handler should not be called for protected path")
			}
		})
	}
}

func TestSelectiveAuthMiddleware_QueryParameters(t *testing.T) {
	tokenService := NewMockTokenService()
	userService := NewMockUserService()
	authMiddleware := middleware.NewAuthMiddleware(tokenService, userService)
	publicPaths := []string{"/health"}

	selectiveAuth := middleware.NewSelectiveAuthMiddleware(authMiddleware, publicPaths)
	handler := NewTestHandler()
	protectedHandler := selectiveAuth.Handler(handler)

	tests := []struct {
		name string
		url  string
	}{
		{
			name: "With query parameters",
			url:  "/health?check=true",
		},
		{
			name: "With multiple query parameters",
			url:  "/health?check=true&format=json",
		},
		{
			name: "With fragment",
			url:  "/health#section1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler.Reset()
			tokenService.Reset()
			userService.Reset()

			req := createTestRequest(tt.url, "")
			w := httptest.NewRecorder()

			protectedHandler.ServeHTTP(w, req)

			// Query parameters should not affect path matching
			assertOK(t, w)
			assert.True(t, handler.WasCalled(), "Next handler should be called")
			assertNoUserInContext(t, handler.GetRequest())

			// Verify no service calls
			assert.Equal(t, 0, tokenService.GetValidateTokenCallCount(), "Token service should not be called")
			assert.Equal(t, 0, userService.GetGetUserByIDCallCount(), "User service should not be called")
		})
	}
}
