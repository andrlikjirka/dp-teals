package bootstrap

import (
	"sync"

	"github.com/caarlos0/env/v10"
	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
)

const defaultDotEnvPath = ".env"

var (
	once     sync.Once
	validate = validator.New()
)

// Config holds the server configuration loaded from environment variables.
type Config struct {
	Env              string `env:"ENV" envDefault:"development"`
	Port             int    `env:"PORT" validate:"required"`
	EnableReflection bool   `env:"ENABLE_REFLECTION" envDefault:"false"`
	DatabaseURL      string `env:"POSTGRES_URL" validate:"required"`
}

// LoadConfig loads the configuration from environment variables and validates it.
func LoadConfig(path string) (Config, error) {
	loadDotEnv(path)

	cfg := Config{}
	if err := env.Parse(&cfg); err != nil {
		return cfg, err
	}

	if err := validate.Struct(&cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}

// MustLoadConfig loads the configuration and panics if there is an error.
func MustLoadConfig(path string) Config {
	cfg, err := LoadConfig(path)
	if err != nil {
		panic(err)
	}
	return cfg
}

// loadDotEnv loads environment variables from the specified .env file.
func loadDotEnv(path string) {
	once.Do(func() {
		if path == "" {
			path = defaultDotEnvPath
		}

		_ = godotenv.Load(path)
		_ = godotenv.Load(path + ".common")
	})
}
