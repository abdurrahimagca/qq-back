package environment

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type ResendConfig struct {
	Url string
	Key string
}

type Config struct {
	APIKey string
	Resend ResendConfig
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

	return &Config{
		APIKey: os.Getenv("API_KEY"),
		Resend: ResendConfig{
			Url: os.Getenv("RESEND_URL"),
			Key: os.Getenv("RESEND_KEY"),
		},
	}, nil
}
