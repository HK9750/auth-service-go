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
)

func main() {
	result := config.Load()
	if err := config.Require(result.Config); err != nil {
		panic(err)
	}

	log := logger.New(logger.Config{Level: result.Config.LogLevel, Format: result.Config.LogFormat})
	slog.SetDefault(log)

	for _, warning := range result.Warnings {
		log.Warn("config warning", "warning", warning)
	}

	if result.Config.GinMode != "" {
		server.SetGinMode(result.Config.GinMode)
	}
	fmt.Print(result.Config)

	var db *sql.DB
	if result.Config.DatabaseDSN != "" {
		connection, err := database.New(database.Config{
			DSN:          result.Config.DatabaseDSN,
			MaxOpenConns: result.Config.DBMaxOpenConns,
			MaxIdleConns: result.Config.DBMaxIdleConns,
			MaxIdleTime:  result.Config.DBMaxIdleTime,
			MaxLifetime:  result.Config.DBMaxLifetime,
			PingTimeout:  result.Config.DBPingTimeout,
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
		log.Warn("database disabled; DATABASE_DSN or DATABASE_URL is empty")
	}

	router := server.NewRouter(server.RouterConfig{
		Logger:        log,
		DB:            db,
		HealthTimeout: result.Config.DBPingTimeout,
	})

	srv := server.New(router, server.Config{
		Address:      result.Config.Address,
		ReadTimeout:  result.Config.ReadTimeout,
		WriteTimeout: result.Config.WriteTimeout,
		IdleTimeout:  result.Config.IdleTimeout,
	})

	go func() {
		log.Info("http server starting", "addr", result.Config.Address)
		if err := srv.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("http server failed", "error", err)
			os.Exit(1)
		}
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)
	<-shutdown

	ctx, cancel := context.WithTimeout(context.Background(), result.Config.ShutdownTimeout)
	defer cancel()

	log.Info("http server shutting down")
	if err := srv.Shutdown(ctx); err != nil {
		log.Error("http server shutdown failed", "error", err)
	}
}
