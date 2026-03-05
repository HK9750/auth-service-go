package config

import (
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

func Load() Config {
	_ = godotenv.Load()

	cfg := Config{
		Address:         getEnv("SERVER_ADDR", ":8080"),
		LogLevel:        getEnv("LOG_LEVEL", "info"),
		LogFormat:       getEnv("LOG_FORMAT", "json"),
		DatabaseDSN:     getEnv("DATABASE_DSN", ""),
		DBMaxOpenConns:  getEnvInt("DB_MAX_OPEN_CONNS", 10),
		DBMaxIdleConns:  getEnvInt("DB_MAX_IDLE_CONNS", 10),
		DBMaxIdleTime:   getEnvDuration("DB_MAX_IDLE_TIME", 15*time.Minute),
		DBMaxLifetime:   getEnvDuration("DB_MAX_LIFETIME", 30*time.Minute),
		DBPingTimeout:   getEnvDuration("DB_PING_TIMEOUT", 5*time.Second),
		ReadTimeout:     getEnvDuration("SERVER_READ_TIMEOUT", 10*time.Second),
		WriteTimeout:    getEnvDuration("SERVER_WRITE_TIMEOUT", 10*time.Second),
		IdleTimeout:     getEnvDuration("SERVER_IDLE_TIMEOUT", 60*time.Second),
		ShutdownTimeout: getEnvDuration("SHUTDOWN_TIMEOUT", 10*time.Second),
		GinMode:         getEnv("GIN_MODE", "release"),
	}

	if cfg.DatabaseDSN == "" {
		cfg.DatabaseDSN = os.Getenv("DATABASE_URL")
	}

	return cfg
}

func getEnv(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func getEnvInt(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}
	return parsed
}
