package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"auth-service/internal/server"
	"auth-service/pkg/config"
	"auth-service/pkg/database"
	"auth-service/pkg/logger"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	log := logger.New(logger.Config{
		Level:  cfg.LogLevel,
		Format: cfg.LogFormat,
	})
	slog.SetDefault(log)

	if cfg.GinMode != "" {
		setGinMode(cfg.GinMode)
	}

	var db *sql.DB
	fmt.Println("Database DSN:", cfg.DatabaseDSN) // Debugging line
	if cfg.DatabaseDSN != "" {
		connection, err := database.New(database.Config{
			DSN:          cfg.DatabaseDSN,
			MaxOpenConns: cfg.DBMaxOpenConns,
			MaxIdleConns: cfg.DBMaxIdleConns,
			MaxIdleTime:  cfg.DBMaxIdleTime,
			MaxLifetime:  cfg.DBMaxLifetime,
			PingTimeout:  cfg.DBPingTimeout,
		})
		if err != nil {
			log.Error("database connection failed", "error", err)
			os.Exit(1)
		}
		db = connection
		defer func() {
			if err := db.Close(); err != nil {
				log.Error("database close failed", "error", err)
			}
		}()
	} else {
		log.Warn("DATABASE_DSN not set, running without database")
	}

	router := server.NewRouter(server.RouterConfig{
		Logger:        log,
		DB:            db,
		HealthTimeout: cfg.DBPingTimeout,
	})

	srv := server.New(router, server.Config{
		Address:      cfg.Address,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	})

	go func() {
		log.Info("http server starting", "addr", cfg.Address)
		if err := srv.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("http server failed", "error", err)
			os.Exit(1)
		}
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)
	<-shutdown

	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	log.Info("http server shutting down")
	if err := srv.Shutdown(ctx); err != nil {
		log.Error("http server shutdown failed", "error", err)
	}
}

func setGinMode(mode string) {
	switch mode {
	case gin.DebugMode, gin.ReleaseMode, gin.TestMode:
		gin.SetMode(mode)
	default:
		gin.SetMode(gin.ReleaseMode)
	}
}
