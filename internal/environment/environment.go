package environment

import (
	"fmt"
	"os"
	"strconv"

	"golang.org/x/net/context"
)

type ResendEnvironment struct {
	Key string
}
type TokenEnvironment struct {
	Secret                 string
	AccessTokenExpireTime  int
	RefreshTokenExpireTime int
	Issuer                 string
	Audience               string
}
type R2Environment struct {
	BucketName      string
	URL             string
	TokenValue      string
	AccessKeyID     string
	SecretAccessKey string
	AccountID       string
}
type APIEnvironment struct {
	Port    string
	Version string

	Title       string
	Description string
}
type Environment struct {
	Resend      ResendEnvironment
	DatabaseURL string
	Ctx         context.Context
	Token       TokenEnvironment
	R2          R2Environment
	API         APIEnvironment
}

func Load() (*Environment, error) {
	accessTokenExpireTime, err := strconv.Atoi(getOrThrow("ACCESS_TOKEN_EXPIRE_TIME"))
	if err != nil {
		return nil, fmt.Errorf("error converting ACCESS_TOKEN_EXPIRE_TIME to int: %w", err)
	}
	refreshTokenExpireTime, err := strconv.Atoi(getOrThrow("REFRESH_TOKEN_EXPIRE_TIME"))
	if err != nil {
		return nil, fmt.Errorf("error converting REFRESH_TOKEN_EXPIRE_TIME to int: %w", err)
	}

	return &Environment{
		Resend: ResendEnvironment{

			Key: getOrThrow("RESEND_KEY"),
		},
		DatabaseURL: getOrThrow("DATABASE_URL"),
		Ctx:         context.Background(),
		Token: TokenEnvironment{
			Secret:                 getOrThrow("TOKEN_SECRET"),
			AccessTokenExpireTime:  accessTokenExpireTime,
			RefreshTokenExpireTime: refreshTokenExpireTime,
			Issuer:                 getOrThrow("ISSUER"),
			Audience:               getOrThrow("AUDIENCE"),
		},
		R2: R2Environment{
			BucketName:      getOrThrow("R2_BUCKET_NAME"),
			URL:             getOrThrow("R2_URL"),
			TokenValue:      getOrThrow("R2_TOKEN_VALUE"),
			AccessKeyID:     getOrThrow("R2_ACCESS_KEY_ID"),
			SecretAccessKey: getOrThrow("R2_SECRET_ACCESS_KEY"),
			AccountID:       getOrThrow("R2_ACCOUNT_ID"),
		},
		API: APIEnvironment{

			Port:        getOrThrow("API_PORT"),
			Version:     getOrReturnPlaceholder("API_VERSION", "0.0.1"),
			Title:       getOrReturnPlaceholder("API_TITLE", "QQ API"),
			Description: getOrReturnPlaceholder("API_DESCRIPTION", "QQ API"),
		},
	}, nil
}

func getOrThrow(env string) string {
	if os.Getenv(env) == "" {
		panic(fmt.Sprintf("environment variable %s is not set", env))
	}
	return os.Getenv(env)
}

func getOrReturnPlaceholder(env string, placeholder string) string {
	if os.Getenv(env) == "" {
		return placeholder
	}
	return os.Getenv(env)
}
