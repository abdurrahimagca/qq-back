package token

import (
	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	jwt.RegisteredClaims

	UserID string `json:"user_id"`
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
