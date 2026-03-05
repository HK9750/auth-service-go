package repository

import (
	"auth-service/internal/domain"
	"context"
	"database/sql"
	"errors"
	"net"

	"github.com/google/uuid"
)

type AuditLogRepository struct {
	DB *sql.DB
}

func NewAuditLogRepository(db *sql.DB) *AuditLogRepository {
	return &AuditLogRepository{DB: db}
}

func (r *AuditLogRepository) CreateAuditLog(ctx context.Context, userID *uuid.UUID, action string, resource *string, metadata []byte, ipAddress net.IP) error {
	if action == "" {
		return errors.New("action cannot be empty")
	}

	query := `INSERT INTO audit_logs (user_id, action, resource, metadata, ip_address)
	          VALUES ($1, $2, $3, $4, $5)`
	_, err := r.DB.ExecContext(ctx, query, userID, action, resource, metadata, ipAddress)
	return err
}

func (r *AuditLogRepository) GetAuditLogsByUserID(ctx context.Context, userID uuid.UUID) ([]domain.AuditLog, error) {
	query := `SELECT id, user_id, action, resource, metadata, ip_address, created_at
	          FROM audit_logs
	          WHERE user_id = $1
	          ORDER BY created_at DESC`

	rows, err := r.DB.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []domain.AuditLog

	for rows.Next() {
		var log domain.AuditLog
		err := rows.Scan(
			&log.ID,
			&log.UserID,
			&log.Action,
			&log.Resource,
			&log.Metadata,
			&log.IPAddress,
			&log.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}

	return logs, nil
}
