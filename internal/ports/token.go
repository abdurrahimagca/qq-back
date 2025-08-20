package ports

import (
	"context"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

type GenerateTokenParams struct {
	UserID string
}

type GenerateTokenResult struct {
	AccessToken  string
	RefreshToken string
}

type ValidateTokenParams struct {
	Token string
}

type ValidateTokenResult struct {
	Claims *Claims
}

type TokenPort interface {
	GenerateTokens(ctx context.Context, params GenerateTokenParams) (GenerateTokenResult, error)
	ValidateToken(ctx context.Context, params ValidateTokenParams) (ValidateTokenResult, error)
}