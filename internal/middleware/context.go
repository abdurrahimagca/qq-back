package middleware

import (
	"context"

	"github.com/abdurrahimagca/qq-back/internal/db"
)

type contextKey string

const (
	UserContextKey contextKey = "user"
)

func WithUser(ctx context.Context, user *db.User) context.Context {
	return context.WithValue(ctx, UserContextKey, user)
}

func GetUserFromContext(ctx context.Context) (*db.User, bool) {
	user, ok := ctx.Value(UserContextKey).(*db.User)
	return user, ok
}

func MustGetUserFromContext(ctx context.Context) *db.User {
	user, ok := GetUserFromContext(ctx)
	if !ok {
		panic("user not found in context - middleware not applied?")
	}
	return user
}