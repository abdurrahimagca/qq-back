package tokens

import (
	"time"

	"github.com/abdurrahimagca/qq-back/internal/environment"
	"github.com/golang-jwt/jwt/v5"
)

type Service interface {
	GenerateTokens(GenerateTokenParams) (GenerateTokenResult, error)
	ValidateToken(ValidateTokenParams) (ValidateTokenResult, error)
}

type service struct {
	environment *environment.Environment
}

func NewService(environment *environment.Environment) Service {
	return &service{
		environment: environment,
	}
}

func (s *service) GenerateTokens(params GenerateTokenParams) (GenerateTokenResult, error) {
	claims := Claims{
		UserID: params.UserID,
		Exp:    time.Now().Add(time.Duration(s.environment.Token.AccessTokenExpireTime) * time.Minute).Unix(),
		Iat:    time.Now().Unix(),
		Iss:    s.environment.Token.Issuer,
		Aud:    s.environment.Token.Audience,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   params.UserID,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(s.environment.Token.AccessTokenExpireTime) * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    s.environment.Token.Issuer,
			Audience:  jwt.ClaimStrings{s.environment.Token.Audience},
		},
	}
	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(s.environment.Token.Secret))
	if err != nil {
		return GenerateTokenResult{}, err
	}
	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(s.environment.Token.Secret))
	if err != nil {
		return GenerateTokenResult{}, err
	}
	return GenerateTokenResult{
		Tokens: &Tokens{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		},
	}, nil
}
func (s *service) ValidateToken(params ValidateTokenParams) (ValidateTokenResult, error) {
	claims := Claims{}
	_, err := jwt.ParseWithClaims(params.Token, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.environment.Token.Secret), nil
	})
	if err != nil {
		return ValidateTokenResult{}, err
	}
	return ValidateTokenResult{
		Claims: &claims,
	}, nil
}
