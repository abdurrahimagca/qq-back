package environment

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	APIKey string
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
	}, nil
}
