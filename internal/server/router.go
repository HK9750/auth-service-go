package server

import (
	"database/sql"
	"time"

	"auth-service/internal/handler"
	"auth-service/pkg/config"

	"log/slog"

	"github.com/gin-gonic/gin"
)

type RouterConfig struct {
	Logger *slog.Logger
	DB     *sql.DB
	Config *config.Config
}

func NewRouter(cfg RouterConfig) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())

	if cfg.Logger != nil {
		router.Use(requestLogger(cfg.Logger))
	} else {
		router.Use(gin.Logger())
	}

	healthHandler := handler.NewHealthHandler(cfg.DB, cfg.Config.DBPingTimeout)
	router.GET("/healthz", healthHandler.Check)

	return router
}

func requestLogger(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		logger.Info("http request",
			"status", c.Writer.Status(),
			"size", c.Writer.Size(),
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"query", c.Request.URL.RawQuery,
			"ip", c.ClientIP(),
			"latency", time.Since(start),
			"user_agent", c.Request.UserAgent(),
		)
	}
}
