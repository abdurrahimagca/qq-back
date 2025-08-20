package token

import (
	"context"
	"fmt"
	"time"

	"github.com/abdurrahimagca/qq-back/internal/environment"
	"github.com/abdurrahimagca/qq-back/internal/ports"
	"github.com/golang-jwt/jwt/v5"
)

type jwtTokenService struct {
	environment *environment.Environment
}

func NewJWTTokenService(conf *environment.Environment) ports.TokenPort {
	return &jwtTokenService{
		environment: conf,
	}
}

func (j *jwtTokenService) GenerateTokens(ctx context.Context, params ports.GenerateTokenParams) (ports.GenerateTokenResult, error) {
	accessTokenClaims := &ports.Claims{
		UserID: params.UserID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   params.UserID,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(j.environment.Token.AccessTokenExpireTime) * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    j.environment.Token.Issuer,
			Audience:  jwt.ClaimStrings{j.environment.Token.Audience},
		},
	}
	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, accessTokenClaims).SignedString([]byte(j.environment.Token.Secret))
	if err != nil {
		return ports.GenerateTokenResult{}, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshTokenClaims := &ports.Claims{
		UserID: params.UserID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   params.UserID,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(j.environment.Token.RefreshTokenExpireTime) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    j.environment.Token.Issuer,
			Audience:  jwt.ClaimStrings{j.environment.Token.Audience},
		},
	}
	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshTokenClaims).SignedString([]byte(j.environment.Token.Secret))
	if err != nil {
		return ports.GenerateTokenResult{}, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return ports.GenerateTokenResult{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (j *jwtTokenService) ValidateToken(ctx context.Context, params ports.ValidateTokenParams) (ports.ValidateTokenResult, error) {
	claims := &ports.Claims{}
	token, err := jwt.ParseWithClaims(params.Token, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(j.environment.Token.Secret), nil
	})

	if err != nil {
		return ports.ValidateTokenResult{}, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return ports.ValidateTokenResult{}, fmt.Errorf("invalid token")
	}

	return ports.ValidateTokenResult{
		Claims: claims,
	}, nil
}