package config

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
)

type Config struct {
	Db   DbConfig   `validate:"required"`
	App  AppConfig  `validate:"required"`
	Auth AuthConfig `validate:"required"`
}

type AppConfig struct {
	Port int    `validate:"required,min=1,max=65535"`
	Env  string `validate:"required"`
}

type DbConfig struct {
	User     string `validate:"required"`
	Password string `validate:"required"`
	DbName   string `validate:"required"`
	Host     string `validate:"required"`
	Port     int    `validate:"required,min=1,max=65535"`
}

type AuthConfig struct {
	JwtSecret       string `validate:"required"`
	JwtDuration     int    `validate:"required"`
	RefreshDuration int    `validate:"required"`
}

var validate = validator.New()

// LoadConfig loads and validates configuration from environment variables
func LoadConfig() (*Config, error) {
	// Load .env file (optional - won't error if file doesn't exist)
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: No .env file found, using system environment variables")
	}

	config := &Config{
		Db: DbConfig{
			User:     os.Getenv("DB_USER"),
			Password: os.Getenv("DB_PASSWORD"),
			DbName:   os.Getenv("DB_NAME"),
			Host:     os.Getenv("DB_HOST"),
			Port:     getEnvAsInt("DB_PORT", 5432), // Default to 5432 if not set
		},
		App: AppConfig{
			Port: getEnvAsInt("PORT", 6666),
			Env:  os.Getenv("ENV"),
		},
		Auth: AuthConfig{
			JwtSecret:       os.Getenv("JWT_SECRET"),
			JwtDuration:     getEnvAsInt("JWT_DURATION", 5),
			RefreshDuration: getEnvAsInt("REFRESH_DURATION", 5),
		},
	}

	// Validate the config struct
	if err := validate.Struct(config); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}
	fmt.Println("refersh duration:", config.Auth.RefreshDuration)
	return config, nil
}

// MustLoadConfig loads config and panics if validation fails
func MustLoadConfig() *Config {
	config, err := LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	return config
}

// getEnvAsInt gets environment variable as integer with a fallback value
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
