package bootstrap

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v10"
	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
)

const defaultDotEnvPath = ".env"

// Config holds the server configuration loaded from environment variables.
type Config struct {
	Env              string        `env:"ENV" envDefault:"development"`
	Port             int           `env:"PORT" validate:"required"`
	EnableReflection bool          `env:"ENABLE_REFLECTION" envDefault:"false"`
	DatabaseURL      string        `env:"POSTGRES_URL" validate:"required"`
	DBConnectTimeout time.Duration `env:"DB_CONNECT_TIMEOUT" envDefault:"10s"`
	ShutdownTimeout  time.Duration `env:"SHUTDOWN_TIMEOUT"   envDefault:"30s"`
}

// LoadEnvFile loads environment variables from the specified .env file.
func LoadEnvFile(path string) error {
	if path == "" {
		path = defaultDotEnvPath
	}
	err := godotenv.Load(path)
	if err != nil {
		return fmt.Errorf("failed to parse env file %q: %w", path, err)
	}
	return nil
}

// LoadConfig loads the configuration from environment variables and validates it.
func LoadConfig() (Config, error) {
	cfg := Config{}
	if err := env.Parse(&cfg); err != nil {
		return cfg, err
	}

	v := validator.New()
	if err := v.Struct(&cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}
