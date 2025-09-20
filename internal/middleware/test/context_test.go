package middleware_test

import (
	"context"
	"testing"

	"github.com/abdurrahimagca/qq-back/internal/middleware"
	"github.com/stretchr/testify/assert"
)

func TestWithUser(t *testing.T) {
	ctx := context.Background()
	user := createTestUser(TestUserID1)

	newCtx := middleware.WithUser(ctx, user)

	assert.NotEqual(t, ctx, newCtx, "Should return new context")

	// Verify user is in context
	retrievedUser, ok := middleware.GetUserFromContext(newCtx)
	assert.True(t, ok, "Should be able to retrieve user from context")
	assert.Equal(t, user, retrievedUser, "Retrieved user should match original")
}

func TestWithUser_NilUser(t *testing.T) {
	ctx := context.Background()

	newCtx := middleware.WithUser(ctx, nil)

	// Should be able to store nil user
	retrievedUser, ok := middleware.GetUserFromContext(newCtx)
	assert.True(t, ok, "Should be able to retrieve value from context")
	assert.Nil(t, retrievedUser, "Retrieved user should be nil")
}

func TestGetUserFromContext_WithUser(t *testing.T) {
	ctx := context.Background()
	user := createTestUser(TestUserID1)
	ctx = middleware.WithUser(ctx, user)

	retrievedUser, ok := middleware.GetUserFromContext(ctx)

	assert.True(t, ok, "Should successfully retrieve user")
	assert.Equal(t, user, retrievedUser, "Retrieved user should match original")
}

func TestGetUserFromContext_WithoutUser(t *testing.T) {
	ctx := context.Background()

	retrievedUser, ok := middleware.GetUserFromContext(ctx)

	assert.False(t, ok, "Should not find user in context")
	assert.Nil(t, retrievedUser, "Retrieved user should be nil")
}

func TestGetUserFromContext_WrongType(t *testing.T) {
	ctx := context.Background()
	// Add wrong type to context
	ctx = context.WithValue(ctx, middleware.UserContextKey, "not-a-user")

	retrievedUser, ok := middleware.GetUserFromContext(ctx)

	assert.False(t, ok, "Should not find user with wrong type")
	assert.Nil(t, retrievedUser, "Retrieved user should be nil")
}

func TestMustGetUserFromContext_WithUser(t *testing.T) {
	ctx := context.Background()
	user := createTestUser(TestUserID1)
	ctx = middleware.WithUser(ctx, user)

	retrievedUser := middleware.MustGetUserFromContext(ctx)

	assert.Equal(t, user, retrievedUser, "Retrieved user should match original")
}

func TestMustGetUserFromContext_WithoutUser(t *testing.T) {
	ctx := context.Background()

	assert.Panics(t, func() {
		middleware.MustGetUserFromContext(ctx)
	}, "Should panic when user not in context")
}

func TestMustGetUserFromContext_PanicMessage(t *testing.T) {
	ctx := context.Background()

	defer func() {
		if r := recover(); r != nil {
			panicMsg, ok := r.(string)
			assert.True(t, ok, "Panic should be a string")
			assert.Contains(t, panicMsg, "user not found in context", "Panic message should be descriptive")
			assert.Contains(t, panicMsg, "middleware not applied", "Panic message should suggest middleware issue")
		}
	}()

	middleware.MustGetUserFromContext(ctx)
	t.Fatal("Should have panicked")
}

func TestMustGetUserFromContext_WrongType(t *testing.T) {
	ctx := context.Background()
	// Add wrong type to context
	ctx = context.WithValue(ctx, middleware.UserContextKey, "not-a-user")

	assert.Panics(t, func() {
		middleware.MustGetUserFromContext(ctx)
	}, "Should panic when wrong type in context")
}

func TestUserContextKey_IsUnique(t *testing.T) {
	// Ensure the context key is properly typed and unique
	ctx := context.Background()
	user := createTestUser(TestUserID1)

	// Add user with our key
	ctx = middleware.WithUser(ctx, user)

	// Try to access with a string key (should fail)
	wrongValue := ctx.Value("user")
	assert.Nil(t, wrongValue, "Should not be able to access with string key")

	// Access with correct key should work
	correctValue := ctx.Value(middleware.UserContextKey)
	assert.Equal(t, user, correctValue, "Should be able to access with correct key")
}

func TestContextChaining(t *testing.T) {
	ctx := context.Background()
	user1 := createTestUser(TestUserID1)
	user2 := createTestUser(TestUserID2)

	// Chain context modifications
	ctx1 := middleware.WithUser(ctx, user1)
	ctx2 := middleware.WithUser(ctx1, user2)

	// Original context should not have user
	_, ok := middleware.GetUserFromContext(ctx)
	assert.False(t, ok, "Original context should not have user")

	// First context should have user1
	retrievedUser1, ok := middleware.GetUserFromContext(ctx1)
	assert.True(t, ok, "First context should have user")
	assert.Equal(t, user1, retrievedUser1, "First context should have user1")

	// Second context should have user2 (overwritten)
	retrievedUser2, ok := middleware.GetUserFromContext(ctx2)
	assert.True(t, ok, "Second context should have user")
	assert.Equal(t, user2, retrievedUser2, "Second context should have user2")
}
