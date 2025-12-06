package config

import (
	"fmt"
	"os"
	"time"
)

type Config struct {
	PostgresURL          string
	RedisURL             string
	JWTSecret            string
	Port                 string
	Environment          string
	JWTExpiration        time.Duration
	JWTRefreshExpiration time.Duration
	AllowedOrigins       string
	LogLevel             string
}

var AppConfig *Config

func Load() (*Config, error) {
	config := &Config{
		PostgresURL:          os.Getenv("POSTGRES_URL"),
		RedisURL:             os.Getenv("REDIS_URL"),
		JWTSecret:            os.Getenv("JWT_SECRET"),
		Port:                 os.Getenv("PORT"),
		Environment:          os.Getenv("ENV"),
		AllowedOrigins:       os.Getenv("ALLOWED_ORIGINS"),
		LogLevel:             os.Getenv("LOG_LEVEL"),
	}

	if config.PostgresURL == "" {
		return nil, fmt.Errorf("POSTGRES_URL environment variable is not set")
	}

	if config.RedisURL == "" {
		return nil, fmt.Errorf("REDIS_URL environment variable is not set")
	}

	if config.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET environment variable is not set")
	}

	if config.Port == "" {
		config.Port = "8080"
	}

	if config.Environment == "" {
		config.Environment = "development"
	}

	if config.AllowedOrigins == "" {
		config.AllowedOrigins = "http://localhost:3000,http://localhost:5173"
	}

	if config.LogLevel == "" {
		config.LogLevel = "info"
	}

	jwtExpStr := os.Getenv("JWT_EXPIRATION")
	if jwtExpStr == "" {
		jwtExpStr = "24h"
	}
	jwtExp, err := time.ParseDuration(jwtExpStr)
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_EXPIRATION duration: %w", err)
	}
	config.JWTExpiration = jwtExp

	jwtRefreshExpStr := os.Getenv("JWT_REFRESH_EXPIRATION")
	if jwtRefreshExpStr == "" {
		jwtRefreshExpStr = "168h"
	}
	jwtRefreshExp, err := time.ParseDuration(jwtRefreshExpStr)
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_REFRESH_EXPIRATION duration: %w", err)
	}
	config.JWTRefreshExpiration = jwtRefreshExp

	AppConfig = config
	return config, nil
}

func Get() *Config {
	if AppConfig == nil {
		panic("config not loaded")
	}
	return AppConfig
}

func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}
