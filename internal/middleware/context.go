package middleware

import (
	"context"

	"github.com/abdurrahimagca/qq-back/internal/user"
)

type contextKey string

const (
	UserContextKey contextKey = "user"
)

func WithUser(ctx context.Context, user *user.User) context.Context {
	return context.WithValue(ctx, UserContextKey, user)
}

func GetUserFromContext(ctx context.Context) (*user.User, bool) {
	user, ok := ctx.Value(UserContextKey).(*user.User)
	return user, ok
}

func MustGetUserFromContext(ctx context.Context) *user.User {
	user, ok := GetUserFromContext(ctx)
	if !ok {
		panic("user not found in context - middleware not applied?")
	}
	return user
}