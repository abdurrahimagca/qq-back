package token_test

import (
	"context"
	"time"

	"github.com/abdurrahimagca/qq-back/internal/environment"
	"github.com/abdurrahimagca/qq-back/internal/platform/token"
	"github.com/golang-jwt/jwt/v5"
)

// BuildEnv creates a test environment with token configuration
func BuildEnv(secret string, accessMins, refreshHours int, issuer, audience string) *environment.Environment {
	return &environment.Environment{
		Ctx: context.Background(),
		Token: environment.TokenEnvironment{
			Secret:                 secret,
			AccessTokenExpireTime:  accessMins,
			RefreshTokenExpireTime: refreshHours,
			Issuer:                 issuer,
			Audience:               audience,
		},
	}
}

// ParseClaims parses a token string and returns claims and header information

func ParseClaims(tokenString string) (*token.Claims, map[string]any, error) {
	var claims token.Claims
	header := make(map[string]any)

	parser := jwt.NewParser(jwt.WithoutClaimsValidation())
	tok, _, err := parser.ParseUnverified(tokenString, &claims)
	if err != nil {
		return nil, nil, err
	}

	for k, v := range tok.Header {
		header[k] = v
	}

	if tok.Claims != nil {
		if c, ok := tok.Claims.(*token.Claims); ok {
			claims = *c
		}
	}

	return &claims, header, nil
}

// IsWithinTolerance checks if the given time is within tolerance of the expected time
func IsWithinTolerance(actual, expected time.Time, tolerance time.Duration) bool {
	diff := actual.Sub(expected)
	if diff < 0 {
		diff = -diff
	}
	return diff <= tolerance
}
