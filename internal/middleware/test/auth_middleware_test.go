package middleware_test

import (
	"net/http/httptest"
	"testing"

	"github.com/abdurrahimagca/qq-back/internal/middleware"
	"github.com/stretchr/testify/assert"
)

func TestNewAuthMiddleware(t *testing.T) {
	tokenService := NewMockTokenService()
	userService := NewMockUserService()

	authMiddleware := middleware.NewAuthMiddleware(tokenService, userService)

	assert.NotNil(t, authMiddleware, "AuthMiddleware should not be nil")
}

func TestAuthMiddleware_RequireAuth_ValidToken(t *testing.T) {
	tokenService := NewMockTokenService()
	userService := NewMockUserService()
	authMiddleware := middleware.NewAuthMiddleware(tokenService, userService)

	// Setup successful token validation
	user := createTestUser(TestUserID1)
	tokenService.SetValidateTokenResult(TestUserID1, nil)
	userService.SetGetUserByIDResult(user, nil)

	// Create test handler
	handler := NewTestHandler()
	protectedHandler := authMiddleware.RequireAuth(handler)

	// Create request with valid token
	req := createTestRequest("/protected", createValidToken(TestUserID1))
	w := httptest.NewRecorder()

	protectedHandler.ServeHTTP(w, req)

	// Verify success
	assertOK(t, w)
	assert.True(t, handler.WasCalled(), "Next handler should be called")
	assertUserInContext(t, handler.GetRequest(), user)

	// Verify service calls
	assert.Equal(t, 1, tokenService.GetValidateTokenCallCount(), "Token service should be called once")
	assert.Equal(t, 1, userService.GetGetUserByIDCallCount(), "User service should be called once")
}

func TestAuthMiddleware_RequireAuth_NoAuthHeader(t *testing.T) {
	tokenService := NewMockTokenService()
	userService := NewMockUserService()
	authMiddleware := middleware.NewAuthMiddleware(tokenService, userService)

	handler := NewTestHandler()
	protectedHandler := authMiddleware.RequireAuth(handler)

	// Create request without Authorization header
	req := createTestRequest("/protected", "")
	w := httptest.NewRecorder()

	protectedHandler.ServeHTTP(w, req)

	// Verify unauthorized response
	assertUnauthorized(t, w)
	assert.False(t, handler.WasCalled(), "Next handler should not be called")
	assert.Contains(t, w.Body.String(), "Authorization header required")

	// Verify no service calls
	assert.Equal(t, 0, tokenService.GetValidateTokenCallCount(), "Token service should not be called")
	assert.Equal(t, 0, userService.GetGetUserByIDCallCount(), "User service should not be called")
}

func TestAuthMiddleware_RequireAuth_EmptyAuthHeader(t *testing.T) {
	tokenService := NewMockTokenService()
	userService := NewMockUserService()
	authMiddleware := middleware.NewAuthMiddleware(tokenService, userService)

	handler := NewTestHandler()
	protectedHandler := authMiddleware.RequireAuth(handler)

	// Create request with empty Authorization header
	req := createTestRequest("/protected", "")
	req.Header.Set("Authorization", "")
	w := httptest.NewRecorder()

	protectedHandler.ServeHTTP(w, req)

	// Verify unauthorized response
	assertUnauthorized(t, w)
	assert.False(t, handler.WasCalled(), "Next handler should not be called")
	assert.Contains(t, w.Body.String(), "Authorization header required")
}

func TestAuthMiddleware_RequireAuth_InvalidAuthHeaderFormat(t *testing.T) {
	tokenService := NewMockTokenService()
	userService := NewMockUserService()
	authMiddleware := middleware.NewAuthMiddleware(tokenService, userService)

	handler := NewTestHandler()
	protectedHandler := authMiddleware.RequireAuth(handler)

	tests := []struct {
		name       string
		authHeader string
	}{
		{
			name:       "Missing Bearer prefix",
			authHeader: "token-without-bearer",
		},
		{
			name:       "Wrong prefix",
			authHeader: "Basic dXNlcjpwYXNz",
		},
		{
			name:       "Only Bearer",
			authHeader: "Bearer",
		},
		{
			name:       "Too many parts",
			authHeader: "Bearer token extra part",
		},
		{
			name:       "Empty token",
			authHeader: "Bearer ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler.Reset()
			req := createTestRequest("/protected", tt.authHeader)
			w := httptest.NewRecorder()

			protectedHandler.ServeHTTP(w, req)

			assertUnauthorized(t, w)
			assert.False(t, handler.WasCalled(), "Next handler should not be called")
			assert.Contains(t, w.Body.String(), "Invalid authorization header format")
		})
	}
}

func TestAuthMiddleware_RequireAuth_TokenValidationError(t *testing.T) {
	tokenService := NewMockTokenService()
	userService := NewMockUserService()
	authMiddleware := middleware.NewAuthMiddleware(tokenService, userService)

	// Setup token validation error
	tokenService.SetValidateTokenError(ErrInvalidToken)

	handler := NewTestHandler()
	protectedHandler := authMiddleware.RequireAuth(handler)

	req := createTestRequest("/protected", createInvalidToken())
	w := httptest.NewRecorder()

	protectedHandler.ServeHTTP(w, req)

	// Verify unauthorized response
	assertUnauthorized(t, w)
	assert.False(t, handler.WasCalled(), "Next handler should not be called")
	assert.Contains(t, w.Body.String(), "Invalid token:")
	assert.Contains(t, w.Body.String(), ErrInvalidToken.Error())

	// Verify service calls
	assert.Equal(t, 1, tokenService.GetValidateTokenCallCount(), "Token service should be called")
	assert.Equal(t, 0, userService.GetGetUserByIDCallCount(), "User service should not be called")
}

func TestAuthMiddleware_RequireAuth_EmptyUserID(t *testing.T) {
	tokenService := NewMockTokenService()
	userService := NewMockUserService()
	authMiddleware := middleware.NewAuthMiddleware(tokenService, userService)

	// Setup token with empty UserID
	tokenService.SetValidateTokenResult("", nil)

	handler := NewTestHandler()
	protectedHandler := authMiddleware.RequireAuth(handler)

	req := createTestRequest("/protected", createValidToken(""))
	w := httptest.NewRecorder()

	protectedHandler.ServeHTTP(w, req)

	// Verify unauthorized response
	assertUnauthorized(t, w)
	assert.False(t, handler.WasCalled(), "Next handler should not be called")
	assert.Contains(t, w.Body.String(), "Invalid token: missing user ID")

	// Verify service calls
	assert.Equal(t, 1, tokenService.GetValidateTokenCallCount(), "Token service should be called")
	assert.Equal(t, 0, userService.GetGetUserByIDCallCount(), "User service should not be called")
}

func TestAuthMiddleware_RequireAuth_InvalidUserIDFormat(t *testing.T) {
	tokenService := NewMockTokenService()
	userService := NewMockUserService()
	authMiddleware := middleware.NewAuthMiddleware(tokenService, userService)

	// Setup token with invalid UUID format
	tokenService.SetValidateTokenResult("not-a-uuid", nil)

	handler := NewTestHandler()
	protectedHandler := authMiddleware.RequireAuth(handler)

	req := createTestRequest("/protected", createValidToken("not-a-uuid"))
	w := httptest.NewRecorder()

	protectedHandler.ServeHTTP(w, req)

	// Verify unauthorized response
	assertUnauthorized(t, w)
	assert.False(t, handler.WasCalled(), "Next handler should not be called")
	assert.Contains(t, w.Body.String(), "Invalid user ID:")

	// Verify service calls
	assert.Equal(t, 1, tokenService.GetValidateTokenCallCount(), "Token service should be called")
	assert.Equal(t, 0, userService.GetGetUserByIDCallCount(), "User service should not be called")
}

func TestAuthMiddleware_RequireAuth_UserNotFound(t *testing.T) {
	tokenService := NewMockTokenService()
	userService := NewMockUserService()
	authMiddleware := middleware.NewAuthMiddleware(tokenService, userService)

	// Setup successful token validation but user not found
	tokenService.SetValidateTokenResult(TestUserID1, nil)
	userService.SetGetUserByIDError(ErrUserNotFound)

	handler := NewTestHandler()
	protectedHandler := authMiddleware.RequireAuth(handler)

	req := createTestRequest("/protected", createValidToken(TestUserID1))
	w := httptest.NewRecorder()

	protectedHandler.ServeHTTP(w, req)

	// Verify unauthorized response
	assertUnauthorized(t, w)
	assert.False(t, handler.WasCalled(), "Next handler should not be called")
	assert.Contains(t, w.Body.String(), "User not found:")
	assert.Contains(t, w.Body.String(), ErrUserNotFound.Error())

	// Verify service calls
	assert.Equal(t, 1, tokenService.GetValidateTokenCallCount(), "Token service should be called")
	assert.Equal(t, 1, userService.GetGetUserByIDCallCount(), "User service should be called")
}

func TestAuthMiddleware_OptionalAuth_ValidToken(t *testing.T) {
	tokenService := NewMockTokenService()
	userService := NewMockUserService()
	authMiddleware := middleware.NewAuthMiddleware(tokenService, userService)

	// Setup successful authentication
	user := createTestUser(TestUserID1)
	tokenService.SetValidateTokenResult(TestUserID1, nil)
	userService.SetGetUserByIDResult(user, nil)

	handler := NewTestHandler()
	optionalHandler := authMiddleware.OptionalAuth(handler)

	req := createTestRequest("/optional", createValidToken(TestUserID1))
	w := httptest.NewRecorder()

	optionalHandler.ServeHTTP(w, req)

	// Verify success with user in context
	assertOK(t, w)
	assert.True(t, handler.WasCalled(), "Next handler should be called")
	assertUserInContext(t, handler.GetRequest(), user)

	// Verify service calls
	assert.Equal(t, 1, tokenService.GetValidateTokenCallCount(), "Token service should be called")
	assert.Equal(t, 1, userService.GetGetUserByIDCallCount(), "User service should be called")
}

func TestAuthMiddleware_OptionalAuth_NoAuthHeader(t *testing.T) {
	tokenService := NewMockTokenService()
	userService := NewMockUserService()
	authMiddleware := middleware.NewAuthMiddleware(tokenService, userService)

	handler := NewTestHandler()
	optionalHandler := authMiddleware.OptionalAuth(handler)

	req := createTestRequest("/optional", "")
	w := httptest.NewRecorder()

	optionalHandler.ServeHTTP(w, req)

	// Verify success without user in context
	assertOK(t, w)
	assert.True(t, handler.WasCalled(), "Next handler should be called")
	assertNoUserInContext(t, handler.GetRequest())

	// Verify no service calls
	assert.Equal(t, 0, tokenService.GetValidateTokenCallCount(), "Token service should not be called")
	assert.Equal(t, 0, userService.GetGetUserByIDCallCount(), "User service should not be called")
}

func TestAuthMiddleware_OptionalAuth_InvalidAuthHeaderFormat(t *testing.T) {
	tokenService := NewMockTokenService()
	userService := NewMockUserService()
	authMiddleware := middleware.NewAuthMiddleware(tokenService, userService)

	handler := NewTestHandler()
	optionalHandler := authMiddleware.OptionalAuth(handler)

	req := createTestRequest("/optional", "Invalid-Format")
	w := httptest.NewRecorder()

	optionalHandler.ServeHTTP(w, req)

	// Verify success without user (invalid format ignored)
	assertOK(t, w)
	assert.True(t, handler.WasCalled(), "Next handler should be called")
	assertNoUserInContext(t, handler.GetRequest())

	// Verify no service calls
	assert.Equal(t, 0, tokenService.GetValidateTokenCallCount(), "Token service should not be called")
	assert.Equal(t, 0, userService.GetGetUserByIDCallCount(), "User service should not be called")
}

func TestAuthMiddleware_OptionalAuth_TokenValidationError(t *testing.T) {
	tokenService := NewMockTokenService()
	userService := NewMockUserService()
	authMiddleware := middleware.NewAuthMiddleware(tokenService, userService)

	// Setup token validation error
	tokenService.SetValidateTokenError(ErrInvalidToken)

	handler := NewTestHandler()
	optionalHandler := authMiddleware.OptionalAuth(handler)

	req := createTestRequest("/optional", createInvalidToken())
	w := httptest.NewRecorder()

	optionalHandler.ServeHTTP(w, req)

	// Verify success without user (error ignored)
	assertOK(t, w)
	assert.True(t, handler.WasCalled(), "Next handler should be called")
	assertNoUserInContext(t, handler.GetRequest())

	// Verify service calls
	assert.Equal(t, 1, tokenService.GetValidateTokenCallCount(), "Token service should be called")
	assert.Equal(t, 0, userService.GetGetUserByIDCallCount(), "User service should not be called")
}

func TestAuthMiddleware_OptionalAuth_EmptyUserID(t *testing.T) {
	tokenService := NewMockTokenService()
	userService := NewMockUserService()
	authMiddleware := middleware.NewAuthMiddleware(tokenService, userService)

	// Setup token with empty UserID
	tokenService.SetValidateTokenResult("", nil)

	handler := NewTestHandler()
	optionalHandler := authMiddleware.OptionalAuth(handler)

	req := createTestRequest("/optional", createValidToken(""))
	w := httptest.NewRecorder()

	optionalHandler.ServeHTTP(w, req)

	// Verify success without user (empty UserID handled gracefully)
	assertOK(t, w)
	assert.True(t, handler.WasCalled(), "Next handler should be called")
	assertNoUserInContext(t, handler.GetRequest())

	// Verify service calls
	assert.Equal(t, 1, tokenService.GetValidateTokenCallCount(), "Token service should be called")
	assert.Equal(t, 0, userService.GetGetUserByIDCallCount(), "User service should not be called")
}

// Note: The current implementation has a bug in OptionalAuth where invalid UUID format
// and user service errors return 401 instead of continuing without user.
// These tests document the current behavior but should be fixed in the implementation.

func TestAuthMiddleware_OptionalAuth_InvalidUserIDFormat_CurrentBehavior(t *testing.T) {
	tokenService := NewMockTokenService()
	userService := NewMockUserService()
	authMiddleware := middleware.NewAuthMiddleware(tokenService, userService)

	// Setup token with invalid UUID format
	tokenService.SetValidateTokenResult("not-a-uuid", nil)

	handler := NewTestHandler()
	optionalHandler := authMiddleware.OptionalAuth(handler)

	req := createTestRequest("/optional", createValidToken("not-a-uuid"))
	w := httptest.NewRecorder()

	optionalHandler.ServeHTTP(w, req)

	// Current behavior: returns 401 (should be fixed to continue without user)
	assertUnauthorized(t, w)
	assert.False(t, handler.WasCalled(), "Next handler should not be called")
	assert.Contains(t, w.Body.String(), "Invalid user ID:")
}

func TestAuthMiddleware_OptionalAuth_UserServiceError_CurrentBehavior(t *testing.T) {
	tokenService := NewMockTokenService()
	userService := NewMockUserService()
	authMiddleware := middleware.NewAuthMiddleware(tokenService, userService)

	// Setup token validation success but user service error
	tokenService.SetValidateTokenResult(TestUserID1, nil)
	userService.SetGetUserByIDError(ErrUserNotFound)

	handler := NewTestHandler()
	optionalHandler := authMiddleware.OptionalAuth(handler)

	req := createTestRequest("/optional", createValidToken(TestUserID1))
	w := httptest.NewRecorder()

	optionalHandler.ServeHTTP(w, req)

	// Current behavior: returns 401 (should be fixed to continue without user)
	assertUnauthorized(t, w)
	assert.False(t, handler.WasCalled(), "Next handler should not be called")
	assert.Contains(t, w.Body.String(), "User not found:")
}
