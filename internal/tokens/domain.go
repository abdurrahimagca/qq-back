package tokens

import (
	"github.com/golang-jwt/jwt/v5"
)

type Tokens struct {
	AccessToken  string
	RefreshToken string
}
type Claims struct {
	UserID string `json:"user_id"`
	Exp    int64  `json:"exp"`
	Iat    int64  `json:"iat"`
	Iss    string `json:"iss"`
	Aud    string `json:"aud"`
	jwt.RegisteredClaims
}
type GenerateTokenParams struct {
	UserID string
}

type GenerateTokenResult struct {
	Tokens *Tokens
}
type ValidateTokenParams struct {
	Token string
}
type ValidateTokenResult struct {
	Claims *Claims
}
