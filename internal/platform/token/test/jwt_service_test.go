package token_test

import (
	"context"
	"testing"
	"time"

	"github.com/abdurrahimagca/qq-back/internal/platform/token"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewJWTTokenService(t *testing.T) {
	env := BuildEnv("test-secret", 15, 24, "test-issuer", "test-audience")

	service := token.NewJWTTokenService(env)

	assert.NotNil(t, service, "Service should not be nil")
}

func TestJWTTokenService_GenerateTokens(t *testing.T) {
	env := BuildEnv("test-secret-123", 15, 24, "test-issuer", "test-audience")
	service := token.NewJWTTokenService(env)
	ctx := context.Background()

	userID := "user-123"
	params := token.GenerateTokenParams{UserID: userID}

	result, err := service.GenerateTokens(ctx, params)

	require.NoError(t, err)
	assert.NotEmpty(t, result.AccessToken, "Access token should not be empty")
	assert.NotEmpty(t, result.RefreshToken, "Refresh token should not be empty")

	// Parse and validate access token claims
	accessClaims, accessHeader, err := ParseClaims(result.AccessToken)
	require.NoError(t, err)

	assert.Equal(t, userID, accessClaims.UserID, "Access token UserID should match")
	assert.Equal(t, userID, accessClaims.Subject, "Access token Subject should match UserID")
	assert.Equal(t, "test-issuer", accessClaims.Issuer, "Access token Issuer should match")
	assert.Contains(t, accessClaims.Audience, "test-audience", "Access token Audience should match")
	assert.Equal(t, "HS256", accessHeader["alg"], "Access token algorithm should be HS256")

	// Validate access token expiration (15 minutes)
	now := time.Now()
	expectedAccessExp := now.Add(15 * time.Minute)
	actualAccessExp := accessClaims.ExpiresAt.Time
	assert.True(t, IsWithinTolerance(actualAccessExp, expectedAccessExp, 2*time.Second),
		"Access token expiration should be ~15 minutes from now")

	// Parse and validate refresh token claims
	refreshClaims, refreshHeader, err := ParseClaims(result.RefreshToken)
	require.NoError(t, err)

	assert.Equal(t, userID, refreshClaims.UserID, "Refresh token UserID should match")
	assert.Equal(t, userID, refreshClaims.Subject, "Refresh token Subject should match UserID")
	assert.Equal(t, "test-issuer", refreshClaims.Issuer, "Refresh token Issuer should match")
	assert.Contains(t, refreshClaims.Audience, "test-audience", "Refresh token Audience should match")
	assert.Equal(t, "HS256", refreshHeader["alg"], "Refresh token algorithm should be HS256")

	// Validate refresh token expiration (24 hours)
	expectedRefreshExp := now.Add(24 * time.Hour)
	actualRefreshExp := refreshClaims.ExpiresAt.Time
	assert.True(t, IsWithinTolerance(actualRefreshExp, expectedRefreshExp, 2*time.Second),
		"Refresh token expiration should be ~24 hours from now")
}

func TestJWTTokenService_GenerateTokens_EmptyUserID(t *testing.T) {
	env := BuildEnv("test-secret", 15, 24, "test-issuer", "test-audience")
	service := token.NewJWTTokenService(env)
	ctx := context.Background()

	params := token.GenerateTokenParams{UserID: ""}

	result, err := service.GenerateTokens(ctx, params)

	require.NoError(t, err, "Should still generate tokens with empty UserID")
	assert.NotEmpty(t, result.AccessToken)
	assert.NotEmpty(t, result.RefreshToken)

	// Verify claims structure
	claims, _, err := ParseClaims(result.AccessToken)
	require.NoError(t, err)
	assert.Empty(t, claims.UserID)
	assert.Empty(t, claims.Subject)
}

func TestJWTTokenService_GenerateTokens_ShortExpiration(t *testing.T) {
	// Test with very short expiration times
	env := BuildEnv("test-secret", 1, 1, "test-issuer", "test-audience") // 1 minute, 1 hour
	service := token.NewJWTTokenService(env)
	ctx := context.Background()

	params := token.GenerateTokenParams{UserID: "user-123"}

	result, err := service.GenerateTokens(ctx, params)

	require.NoError(t, err)

	// Verify short expiration
	claims, _, err := ParseClaims(result.AccessToken)
	require.NoError(t, err)

	now := time.Now()
	expectedExp := now.Add(1 * time.Minute)
	actualExp := claims.ExpiresAt.Time
	assert.True(t, IsWithinTolerance(actualExp, expectedExp, 2*time.Second),
		"Should handle short expiration times correctly")
}

func TestJWTTokenService_ValidateToken_ValidToken(t *testing.T) {
	env := BuildEnv("test-secret-456", 15, 24, "test-issuer", "test-audience")
	service := token.NewJWTTokenService(env)
	ctx := context.Background()

	// Generate a token first
	genParams := token.GenerateTokenParams{UserID: "user-456"}
	genResult, err := service.GenerateTokens(ctx, genParams)
	require.NoError(t, err)

	// Validate the generated token
	valParams := token.ValidateTokenParams{Token: genResult.AccessToken}
	valResult, err := service.ValidateToken(ctx, valParams)

	require.NoError(t, err)
	require.NotNil(t, valResult.Claims)
	assert.Equal(t, "user-456", valResult.Claims.UserID)
	assert.Equal(t, "user-456", valResult.Claims.Subject)
	assert.Equal(t, "test-issuer", valResult.Claims.Issuer)
	assert.Contains(t, valResult.Claims.Audience, "test-audience")
}

func TestJWTTokenService_ValidateToken_ExpiredToken(t *testing.T) {
	env := BuildEnv("test-secret", 15, 24, "test-issuer", "test-audience")
	service := token.NewJWTTokenService(env)
	ctx := context.Background()

	// Create an expired token manually
	now := time.Now()
	expiredClaims := &token.Claims{
		UserID: "user-expired",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "user-expired",
			ExpiresAt: jwt.NewNumericDate(now.Add(-1 * time.Hour)), // Expired 1 hour ago
			IssuedAt:  jwt.NewNumericDate(now.Add(-2 * time.Hour)),
			Issuer:    "test-issuer",
			Audience:  jwt.ClaimStrings{"test-audience"},
		},
	}

	expiredToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, expiredClaims).
		SignedString([]byte("test-secret"))
	require.NoError(t, err)

	// Try to validate expired token
	params := token.ValidateTokenParams{Token: expiredToken}
	_, err = service.ValidateToken(ctx, params)

	require.Error(t, err, "Should return error for expired token")
	assert.Contains(t, err.Error(), "failed to parse token", "Error should mention parsing failure")
}

func TestJWTTokenService_ValidateToken_WrongSecret(t *testing.T) {
	// Generate token with one secret
	env1 := BuildEnv("secret-1", 15, 24, "test-issuer", "test-audience")
	service1 := token.NewJWTTokenService(env1)
	ctx := context.Background()

	genParams := token.GenerateTokenParams{UserID: "user-123"}
	genResult, err := service1.GenerateTokens(ctx, genParams)
	require.NoError(t, err)

	// Try to validate with different secret
	env2 := BuildEnv("secret-2", 15, 24, "test-issuer", "test-audience")
	service2 := token.NewJWTTokenService(env2)

	valParams := token.ValidateTokenParams{Token: genResult.AccessToken}
	_, err = service2.ValidateToken(ctx, valParams)

	require.Error(t, err, "Should return error for wrong secret")
	assert.Contains(t, err.Error(), "failed to parse token", "Error should mention parsing failure")
}

func TestJWTTokenService_ValidateToken_WrongAlgorithm(t *testing.T) {
	env := BuildEnv("test-secret", 15, 24, "test-issuer", "test-audience")
	service := token.NewJWTTokenService(env)
	ctx := context.Background()

	// Create a token with RS256 algorithm (not HS256)
	claims := &token.Claims{
		UserID: "user-123",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "user-123",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "test-issuer",
			Audience:  jwt.ClaimStrings{"test-audience"},
		},
	}

	// Create token with different algorithm (this will fail during validation)
	// For testing purposes, we'll create a malformed token that claims to use RS256
	wrongAlgToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	wrongAlgToken.Header["alg"] = "RS256" // Manually set wrong algorithm in header

	tokenString, err := wrongAlgToken.SignedString([]byte("test-secret"))
	require.NoError(t, err)

	// Try to validate
	params := token.ValidateTokenParams{Token: tokenString}
	_, err = service.ValidateToken(ctx, params)

	require.Error(t, err, "Should return error for unexpected algorithm")
	assert.Contains(t, err.Error(), "unexpected signing method", "Error should mention unexpected signing method")
}

func TestJWTTokenService_ValidateToken_TamperedToken(t *testing.T) {
	env := BuildEnv("test-secret", 15, 24, "test-issuer", "test-audience")
	service := token.NewJWTTokenService(env)
	ctx := context.Background()

	// Generate a valid token
	genParams := token.GenerateTokenParams{UserID: "user-123"}
	genResult, err := service.GenerateTokens(ctx, genParams)
	require.NoError(t, err)

	// Tamper with the token by changing a character
	tamperedToken := genResult.AccessToken[:len(genResult.AccessToken)-5] + "XXXXX"

	// Try to validate tampered token
	params := token.ValidateTokenParams{Token: tamperedToken}
	_, err = service.ValidateToken(ctx, params)

	require.Error(t, err, "Should return error for tampered token")
	assert.Contains(t, err.Error(), "failed to parse token", "Error should mention parsing failure")
}

func TestJWTTokenService_ValidateToken_MalformedToken(t *testing.T) {
	env := BuildEnv("test-secret", 15, 24, "test-issuer", "test-audience")
	service := token.NewJWTTokenService(env)
	ctx := context.Background()

	tests := []struct {
		name  string
		token string
	}{
		{
			name:  "Empty token",
			token: "",
		},
		{
			name:  "Invalid format",
			token: "not.a.jwt.token",
		},
		{
			name:  "Random string",
			token: "this-is-not-a-jwt",
		},
		{
			name:  "Incomplete JWT",
			token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.incomplete",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := token.ValidateTokenParams{Token: tt.token}
			_, err := service.ValidateToken(ctx, params)

			require.Error(t, err, "Should return error for malformed token")
			assert.Contains(t, err.Error(), "failed to parse token", "Error should mention parsing failure")
		})
	}
}

func TestJWTTokenService_ValidateToken_MultipleAudiences(t *testing.T) {
	env := BuildEnv("test-secret", 15, 24, "test-issuer", "aud1,aud2,aud3")
	service := token.NewJWTTokenService(env)
	ctx := context.Background()

	// Generate token
	genParams := token.GenerateTokenParams{UserID: "user-123"}
	genResult, err := service.GenerateTokens(ctx, genParams)
	require.NoError(t, err)

	// Validate token
	valParams := token.ValidateTokenParams{Token: genResult.AccessToken}
	valResult, err := service.ValidateToken(ctx, valParams)

	require.NoError(t, err)
	require.NotNil(t, valResult.Claims)

	// Check that audience is handled correctly
	// Note: The current implementation uses jwt.ClaimStrings{env.Token.Audience}
	// which creates a single-element slice, not a comma-separated parse
	assert.Contains(t, valResult.Claims.Audience, "aud1,aud2,aud3")
}
