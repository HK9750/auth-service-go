package handler

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type HealthHandler struct {
	db      *sql.DB
	timeout time.Duration
}

type DBStatus string

var (
	DBStatusOK          DBStatus = "ok"
	DBStatusUnreachable DBStatus = "unreachable"
	DBStatusSkipped     DBStatus = "skipped"
)

type HealthStatus string

var (
	HealthStatusOK       HealthStatus = "ok"
	HealthStatusDegraded HealthStatus = "degraded"
)

func NewHealthHandler(db *sql.DB, timeout time.Duration) *HealthHandler {
	if timeout <= 0 {
		timeout = 2 * time.Second
	}

	return &HealthHandler{
		db:      db,
		timeout: timeout,
	}
}

func (h *HealthHandler) Check(c *gin.Context) {
	if h.db == nil {
		c.JSON(http.StatusOK, gin.H{"status": HealthStatusOK, "db": DBStatusSkipped})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), h.timeout)
	defer cancel()

	if err := h.db.PingContext(ctx); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": HealthStatusDegraded, "db": DBStatusUnreachable})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": HealthStatusOK, "db": DBStatusOK})
}
