package token

import (
	"context"
)

// Service captures token generation and validation behaviour required by the app layer.
type Service interface {
	GenerateTokens(ctx context.Context, params GenerateTokenParams) (GenerateTokenResult, error)
	ValidateToken(ctx context.Context, params ValidateTokenParams) (ValidateTokenResult, error)
}
