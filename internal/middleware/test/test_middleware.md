# Middleware Module Test Plan

## Purpose & Scope
- Test HTTP middleware components in `internal/middleware`
- Verify authentication, authorization, context management, and selective path protection
- Ensure proper error handling and request flow control

## Component Map
- **AuthMiddleware (`auth.go`)**: Required and optional authentication with token validation and user context injection
- **SelectiveAuthMiddleware (`selective_auth.go`)**: Path-based authentication requirement with public route exceptions
- **Context utilities (`context.go`)**: User context management helpers and extraction utilities

## Requirements & Constraints
1. **Authentication**: Bearer token validation with proper error responses
2. **User Context**: Inject authenticated user into request context
3. **Path Selection**: Apply auth selectively based on route patterns
4. **Error Handling**: Return appropriate HTTP status codes and error messages
5. **Token Integration**: Work with token service for JWT validation
6. **User Resolution**: Resolve user from database using token claims

## Test Strategy

### Unit Tests — HTTP Middleware
- Framework: `testing`, `testify/assert`, `testify/require`, `net/http/httptest`
- Dependencies: Mock token service and user service implementations
- Test HTTP request/response patterns with real middleware handlers

### Mock Dependencies Design

#### Mock Token Service
```go
type MockTokenService struct {
    ValidateTokenFunc func(ctx context.Context, params token.ValidateTokenParams) (token.ValidateTokenResult, error)
    ValidateCalls     []token.ValidateTokenParams
}
```

#### Mock User Service
```go
type MockUserService struct {
    GetUserByIDFunc func(ctx context.Context, userID pgtype.UUID) (*db.User, error)
    GetUserByIDCalls []pgtype.UUID
}
```

## Test Matrix

### AuthMiddleware Tests

#### Constructor
- **`NewAuthMiddleware`**
  - Returns non-nil middleware with provided dependencies

#### RequireAuth Method
- **Valid Authentication**
  - Bearer token with valid JWT → user added to context, next handler called
  - Token service returns valid claims → user service resolves user successfully
  - Request proceeds with authenticated context

- **Missing Authentication**
  - No Authorization header → 401 Unauthorized response
  - Empty Authorization header → 401 Unauthorized response

- **Invalid Token Format**
  - Authorization header without "Bearer " prefix → 401 Unauthorized
  - Authorization header with wrong format (no space, extra parts) → 401 Unauthorized
  - Authorization header with only "Bearer" (no token) → 401 Unauthorized

- **Token Validation Failures**
  - Token service returns validation error → 401 Unauthorized with error message
  - Expired token → 401 Unauthorized
  - Invalid signature → 401 Unauthorized

- **User Resolution Failures**
  - Token claims contain empty UserID → 401 Unauthorized
  - Token claims contain invalid UUID format → 401 Unauthorized
  - User service fails to find user → 401 Unauthorized with error message
  - Database error during user lookup → 401 Unauthorized

#### OptionalAuth Method
- **Valid Authentication**
  - Bearer token with valid JWT → user added to context, next handler called
  - Same success path as RequireAuth

- **Missing Authentication**
  - No Authorization header → next handler called without user context
  - Request proceeds normally

- **Invalid Token Format**
  - Invalid Authorization header format → next handler called without user context
  - Does not return error, continues without authentication

- **Token Validation Failures**
  - Invalid token → next handler called without user context
  - Token service error → next handler called without user context
  - Does not block request flow

- **User Resolution Failures**
  - Empty UserID in valid token → next handler called without user context
  - Invalid UUID format → 401 Unauthorized (current bug in code?)
  - User service error → 401 Unauthorized (current bug in code?)

### SelectiveAuthMiddleware Tests

#### Constructor
- **`NewSelectiveAuthMiddleware`**
  - Returns non-nil middleware with auth middleware and public paths

#### Handler Method
- **Public Path Access**
  - Request to exact public path → next handler called without auth
  - Request to path starting with public prefix → next handler called without auth
  - Multiple public paths configuration → all paths bypass auth

- **Protected Path Access**
  - Request to non-public path → RequireAuth middleware applied
  - Authentication required for protected routes
  - Same auth validation behavior as AuthMiddleware.RequireAuth

- **Path Matching Logic**
  - Path prefix matching works correctly
  - Case sensitivity behavior
  - Root path and nested path handling
  - Query parameters don't affect path matching

### Context Utilities Tests

#### WithUser Function
- **User Context Addition**
  - Add user to context → context contains user
  - Nil user → context accepts nil value
  - User persists through context chain

#### GetUserFromContext Function
- **User Retrieval**
  - Context with user → returns user and true
  - Context without user → returns nil and false
  - Context with wrong type → returns nil and false

#### MustGetUserFromContext Function
- **User Extraction**
  - Context with user → returns user
  - Context without user → panics with descriptive message
  - Panic message indicates middleware issue

## Test Utilities & Layout

```
internal/middleware/test/
├── auth_middleware_test.go         # AuthMiddleware unit tests
├── selective_auth_middleware_test.go # SelectiveAuthMiddleware unit tests
├── context_test.go                 # Context utilities tests
├── mocks_test.go                   # Mock implementations for dependencies
└── test_helpers_test.go            # HTTP test utilities and helpers
```

### Helper Functions
- `createTestRequest(path, authHeader string) *http.Request`
- `createTestUser(id string) *db.User`
- `createValidToken(userID string) string`
- `assertUnauthorized(t *testing.T, w *httptest.ResponseRecorder)`
- `assertUserInContext(t *testing.T, r *http.Request, expectedUser *db.User)`

### Test Data Setup
- Sample user objects with various UUIDs
- Valid and invalid JWT tokens
- Various path configurations for selective auth
- Error scenarios for dependency failures

## Edge Cases & Error Scenarios

### Authentication Edge Cases
- Very long Authorization headers
- Special characters in tokens
- Unicode characters in error messages
- Concurrent request handling

### Path Matching Edge Cases
- Empty public paths list
- Duplicate public paths
- Overlapping path prefixes
- Root path ("/") in public paths

### Context Edge Cases
- Context cancellation during middleware execution
- Context timeout scenarios
- Nested context modifications

## Performance Considerations
- Token validation should not be called unnecessarily
- User lookup should be cached or optimized
- Context operations should be lightweight
- Middleware ordering affects performance

## Running Tests
- Unit tests: `go test ./internal/middleware/test`
- Coverage: `go test -cover ./internal/middleware/test`
- Race detection: `go test -race ./internal/middleware/test`

## Success Criteria
- ✅ All authentication flows work correctly
- ✅ Proper HTTP status codes and error messages
- ✅ User context injection and extraction
- ✅ Path-based authentication selection
- ✅ Error handling without information leakage
- ✅ Middleware chain integration

## Security Considerations
- Error messages don't expose sensitive information
- Token validation failures are handled securely
- User context access is type-safe
- Authorization bypass paths are explicitly configured

## Future Enhancements
- Add role-based authorization middleware
- Implement request rate limiting middleware
- Add request logging and monitoring middleware
- Support for API key authentication alongside JWT
