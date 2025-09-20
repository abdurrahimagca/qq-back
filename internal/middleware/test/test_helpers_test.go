package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/abdurrahimagca/qq-back/internal/db"
	"github.com/abdurrahimagca/qq-back/internal/middleware"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createTestRequest creates a GET HTTP request for testing.
func createTestRequest(path, authHeader string) *http.Request {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}
	return req
}

// createTestUser creates a test user with the given ID and a placeholder username.
func createTestUser(id string) *db.User {
	var userID pgtype.UUID
	userID.Scan(id)

	return &db.User{
		ID:       userID,
		Username: "user-" + id[len(id)-4:],
	}
}

// createValidToken creates a valid Bearer token string.
func createValidToken(userID string) string {
	return "Bearer valid-jwt-token-" + userID
}

// createInvalidToken creates an invalid Bearer token string.
func createInvalidToken() string {
	return "Bearer invalid-jwt-token"
}

// assertUnauthorized checks that the response is 401 Unauthorized.
func assertUnauthorized(t *testing.T, w *httptest.ResponseRecorder) {
	assert.Equal(t, http.StatusUnauthorized, w.Code, "Expected 401 Unauthorized")
}

// assertOK checks that the response is 200 OK.
func assertOK(t *testing.T, w *httptest.ResponseRecorder) {
	assert.Equal(t, http.StatusOK, w.Code, "Expected 200 OK")
}

// assertUserInContext checks that the user is present in the request context.
func assertUserInContext(t *testing.T, r *http.Request, expectedUser *db.User) {
	user, ok := middleware.GetUserFromContext(r.Context())
	require.True(t, ok, "User should be present in context")
	require.NotNil(t, user, "User should not be nil")

	if expectedUser != nil {
		assert.Equal(t, expectedUser.ID, user.ID, "User ID should match")
		assert.Equal(t, expectedUser.Username, user.Username, "Username should match")
	}
}

// assertNoUserInContext checks that no user is present in the request context.
func assertNoUserInContext(t *testing.T, r *http.Request) {
	_, ok := middleware.GetUserFromContext(r.Context())
	assert.False(t, ok, "User should not be present in context")
}

// TestHandler is a simple handler for testing middleware.
type TestHandler struct {
	Called   bool
	Request  *http.Request
	Response http.ResponseWriter
}

func NewTestHandler() *TestHandler {
	return &TestHandler{}
}

func (h *TestHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.Called = true
	h.Request = r
	h.Response = w
	w.WriteHeader(http.StatusOK)
}

func (h *TestHandler) Reset() {
	h.Called = false
	h.Request = nil
	h.Response = nil
}

func (h *TestHandler) WasCalled() bool {
	return h.Called
}

func (h *TestHandler) GetRequest() *http.Request {
	return h.Request
}

// Common test UUIDs.
const (
	TestUserID1 = "550e8400-e29b-41d4-a716-446655440000"
	TestUserID2 = "550e8400-e29b-41d4-a716-446655440001"
	TestUserID3 = "550e8400-e29b-41d4-a716-446655440002"
)
