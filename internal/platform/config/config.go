package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	HTTPAddr        string
	DatabaseURL     string
	LogLevel        string
	CambioURL       string
	CambioTimeoutMS int
	ShutdownTimeout time.Duration
	MigrationPath   string
	APIPrefix       string
}

func Load() (Config, error) {
	cfg := Config{
		HTTPAddr:        getEnv("HTTP_ADDR", ":8080"),
		DatabaseURL:     os.Getenv("DATABASE_URL"),
		LogLevel:        getEnv("LOG_LEVEL", "info"),
		CambioURL:       getEnv("CAMBIO_URL", ""),
		CambioTimeoutMS: getEnvInt("CAMBIO_TIMEOUT_MS", 2000),
		ShutdownTimeout: time.Duration(getEnvInt("SHUTDOWN_TIMEOUT_S", 15)) * time.Second,
		MigrationPath:   getEnv("MIGRATION_PATH", "migrations"),
		APIPrefix:       getEnv("API_PREFIX", "/api/v1"),
	}

	if err := cfg.validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (c Config) validate() error {
	if c.DatabaseURL == "" {
		return errors.New("variável obrigatória ausente: DATABASE_URL")
	}
	if c.CambioTimeoutMS <= 0 {
		return fmt.Errorf("CAMBIO_TIMEOUT_MS deve ser positivo, recebido %d", c.CambioTimeoutMS)
	}
	return nil
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}
