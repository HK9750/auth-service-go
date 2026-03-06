package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Address         string
	LogLevel        string
	LogFormat       string
	DatabaseDSN     string
	DBMaxOpenConns  int
	DBMaxIdleConns  int
	DBMaxIdleTime   time.Duration
	DBMaxLifetime   time.Duration
	DBPingTimeout   time.Duration
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
	GinMode         string
}

type LoadResult struct {
	Config   Config
	Warnings []string
}

func Load() LoadResult {
	_ = godotenv.Load()

	cfg := Config{
		Address:         envString("SERVER_ADDR", ":8080"),
		LogLevel:        envString("LOG_LEVEL", "info"),
		LogFormat:       envString("LOG_FORMAT", "json"),
		DatabaseDSN:     envString("DATABASE_DSN", ""),
		DBMaxOpenConns:  envInt("DB_MAX_OPEN_CONNS", 10),
		DBMaxIdleConns:  envInt("DB_MAX_IDLE_CONNS", 10),
		DBMaxIdleTime:   envDuration("DB_MAX_IDLE_TIME", 15*time.Minute),
		DBMaxLifetime:   envDuration("DB_MAX_LIFETIME", 30*time.Minute),
		DBPingTimeout:   envDuration("DB_PING_TIMEOUT", 5*time.Second),
		ReadTimeout:     envDuration("SERVER_READ_TIMEOUT", 10*time.Second),
		WriteTimeout:    envDuration("SERVER_WRITE_TIMEOUT", 10*time.Second),
		IdleTimeout:     envDuration("SERVER_IDLE_TIMEOUT", 60*time.Second),
		ShutdownTimeout: envDuration("SHUTDOWN_TIMEOUT", 10*time.Second),
		GinMode:         envString("GIN_MODE", "release"),
	}

	if cfg.DatabaseDSN == "" {
		cfg.DatabaseDSN = envString("DATABASE_URL", "")
	}

	return LoadResult{Config: cfg, Warnings: validate(cfg)}
}

func validate(cfg Config) []string {
	var warnings []string
	if cfg.Address == "" {
		warnings = append(warnings, "SERVER_ADDR is empty")
	}
	if cfg.LogLevel == "" {
		warnings = append(warnings, "LOG_LEVEL is empty")
	}
	if cfg.LogFormat == "" {
		warnings = append(warnings, "LOG_FORMAT is empty")
	}
	if cfg.ReadTimeout <= 0 {
		warnings = append(warnings, "SERVER_READ_TIMEOUT must be positive")
	}
	if cfg.WriteTimeout <= 0 {
		warnings = append(warnings, "SERVER_WRITE_TIMEOUT must be positive")
	}
	if cfg.IdleTimeout <= 0 {
		warnings = append(warnings, "SERVER_IDLE_TIMEOUT must be positive")
	}
	if cfg.ShutdownTimeout <= 0 {
		warnings = append(warnings, "SHUTDOWN_TIMEOUT must be positive")
	}
	if cfg.DBMaxOpenConns < 0 {
		warnings = append(warnings, "DB_MAX_OPEN_CONNS must be >= 0")
	}
	if cfg.DBMaxIdleConns < 0 {
		warnings = append(warnings, "DB_MAX_IDLE_CONNS must be >= 0")
	}
	if cfg.DBMaxIdleTime < 0 {
		warnings = append(warnings, "DB_MAX_IDLE_TIME must be >= 0")
	}
	if cfg.DBMaxLifetime < 0 {
		warnings = append(warnings, "DB_MAX_LIFETIME must be >= 0")
	}
	if cfg.DBPingTimeout < 0 {
		warnings = append(warnings, "DB_PING_TIMEOUT must be >= 0")
	}
	return warnings
}

func envString(key, fallback string) string {
	if value, ok := lookup(key); ok {
		return value
	}
	return fallback
}

func envInt(key string, fallback int) int {
	value, ok := lookup(key)
	if !ok {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func envDuration(key string, fallback time.Duration) time.Duration {
	value, ok := lookup(key)
	if !ok {
		return fallback
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func lookup(key string) (string, bool) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return "", false
	}
	return value, true
}

func Require(cfg Config) error {
	if cfg.Address == "" {
		return errors.New("SERVER_ADDR is required")
	}
	if cfg.ReadTimeout <= 0 || cfg.WriteTimeout <= 0 || cfg.IdleTimeout <= 0 || cfg.ShutdownTimeout <= 0 {
		return errors.New("server timeouts must be positive")
	}
	if cfg.DBMaxOpenConns < 0 || cfg.DBMaxIdleConns < 0 || cfg.DBMaxIdleTime < 0 || cfg.DBMaxLifetime < 0 || cfg.DBPingTimeout < 0 {
		return errors.New("database pool settings must be >= 0")
	}
	if cfg.LogFormat != "" && !strings.EqualFold(cfg.LogFormat, "json") && !strings.EqualFold(cfg.LogFormat, "text") {
		return fmt.Errorf("LOG_FORMAT must be json or text: %s", cfg.LogFormat)
	}
	return nil
}
