package environment

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type ResendConfig struct {
	Url string
	Key string
}
type TokenConfig struct {
	Secret                 string
	AccessTokenExpireTime  int
	RefreshTokenExpireTime int
	Issuer                 string
	Audience               string
}
type R2Config struct {
	BucketName string
	URL         string
	TokenValue    string
	AccessKeyID   string
	SecretAccessKey string
	AccountID       string
}

type Config struct {
	APIKey      string
	Resend      ResendConfig
	DatabaseURL string
	Token       TokenConfig
	R2         R2Config

}

func Load() (*Config, error) {
	env := os.Getenv("ENV")
	if env == "" {
		env = "development"
	}

	envFileName := fmt.Sprintf(".env.%s", env)
	if err := godotenv.Load(envFileName); err != nil {
		return nil, fmt.Errorf("error loading %s: %w", envFileName, err)
	}
	accessTokenExpireTime, err := strconv.Atoi(os.Getenv("ACCESS_TOKEN_EXPIRE_TIME"))
	if err != nil {
		return nil, fmt.Errorf("error converting ACCESS_TOKEN_EXPIRE_TIME to int: %w", err)
	}
	refreshTokenExpireTime, err := strconv.Atoi(os.Getenv("REFRESH_TOKEN_EXPIRE_TIME"))
	if err != nil {
		return nil, fmt.Errorf("error converting REFRESH_TOKEN_EXPIRE_TIME to int: %w", err)
	}

	return &Config{
		APIKey: os.Getenv("API_KEY"),
		Resend: ResendConfig{
			Url: os.Getenv("RESEND_URL"),
			Key: os.Getenv("RESEND_KEY"),
		},
		DatabaseURL: os.Getenv("DATABASE_URL"),
		Token: TokenConfig{
			Secret:                 os.Getenv("TOKEN_SECRET"),
			AccessTokenExpireTime:  accessTokenExpireTime,
			RefreshTokenExpireTime: refreshTokenExpireTime,
			Issuer:                 os.Getenv("ISSUER"),
			Audience:               os.Getenv("AUDIENCE"),
		},
		R2: R2Config{
			BucketName:     os.Getenv("R2_BUCKET_NAME"),
			URL:           os.Getenv("R2_URL"),
			TokenValue:    os.Getenv("R2_TOKEN_VALUE"),
			AccessKeyID:  os.Getenv("R2_ACCESS_KEY_ID"),
			SecretAccessKey: os.Getenv("R2_SECRET_ACCESS_KEY"),
			AccountID: os.Getenv("R2_ACCOUNT_ID"),
		},
	}, nil
}
