package repository

import (
	"context"
	"database/sql"
	"errors"
	"net"

	"auth-service/internal/domain"

	"github.com/google/uuid"
)

type AuditLogRepository struct {
	db *sql.DB
}

func NewAuditLogRepository(db *sql.DB) *AuditLogRepository {
	return &AuditLogRepository{db: db}
}

func (r *AuditLogRepository) Create(ctx context.Context, userID *uuid.UUID, action string, resource *string, metadata []byte, ipAddress net.IP) error {
	if action == "" {
		return ErrInvalidArgument
	}

	const query = `INSERT INTO audit_logs (user_id, action, resource, metadata, ip_address)
        VALUES ($1, $2, $3, $4, $5)`

	_, err := r.db.ExecContext(ctx, query, userID, action, resource, metadata, ipAddress)
	return err
}

func (r *AuditLogRepository) ByUserID(ctx context.Context, userID uuid.UUID) ([]domain.AuditLog, error) {
	if userID == uuid.Nil {
		return nil, ErrInvalidArgument
	}

	const query = `SELECT id, user_id, action, resource, metadata, ip_address, created_at
        FROM audit_logs
        WHERE user_id = $1
        ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []domain.AuditLog
	for rows.Next() {
		var logEntry domain.AuditLog
		if err := rows.Scan(
			&logEntry.ID,
			&logEntry.UserID,
			&logEntry.Action,
			&logEntry.Resource,
			&logEntry.Metadata,
			&logEntry.IPAddress,
			&logEntry.CreatedAt,
		); err != nil {
			return nil, err
		}
		logs = append(logs, logEntry)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(logs) == 0 {
		return nil, ErrNotFound
	}
	return logs, nil
}

func (r *AuditLogRepository) ByAction(ctx context.Context, action string, limit int) ([]domain.AuditLog, error) {
	if action == "" || limit <= 0 {
		return nil, ErrInvalidArgument
	}

	const query = `SELECT id, user_id, action, resource, metadata, ip_address, created_at
        FROM audit_logs
        WHERE action = $1
        ORDER BY created_at DESC
        LIMIT $2`

	rows, err := r.db.QueryContext(ctx, query, action, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []domain.AuditLog
	for rows.Next() {
		var logEntry domain.AuditLog
		if err := rows.Scan(
			&logEntry.ID,
			&logEntry.UserID,
			&logEntry.Action,
			&logEntry.Resource,
			&logEntry.Metadata,
			&logEntry.IPAddress,
			&logEntry.CreatedAt,
		); err != nil {
			return nil, err
		}
		logs = append(logs, logEntry)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(logs) == 0 {
		return nil, ErrNotFound
	}
	return logs, nil
}

func (r *AuditLogRepository) ByResource(ctx context.Context, resource string, limit int) ([]domain.AuditLog, error) {
	if resource == "" || limit <= 0 {
		return nil, ErrInvalidArgument
	}

	const query = `SELECT id, user_id, action, resource, metadata, ip_address, created_at
        FROM audit_logs
        WHERE resource = $1
        ORDER BY created_at DESC
        LIMIT $2`

	rows, err := r.db.QueryContext(ctx, query, resource, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []domain.AuditLog
	for rows.Next() {
		var logEntry domain.AuditLog
		if err := rows.Scan(
			&logEntry.ID,
			&logEntry.UserID,
			&logEntry.Action,
			&logEntry.Resource,
			&logEntry.Metadata,
			&logEntry.IPAddress,
			&logEntry.CreatedAt,
		); err != nil {
			return nil, err
		}
		logs = append(logs, logEntry)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(logs) == 0 {
		return nil, ErrNotFound
	}
	return logs, nil
}

func (r *AuditLogRepository) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	if userID == uuid.Nil {
		return ErrInvalidArgument
	}

	const query = `DELETE FROM audit_logs WHERE user_id = $1`
	res, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *AuditLogRepository) DeleteOlderThan(ctx context.Context, cutoff string) error {
	if cutoff == "" {
		return ErrInvalidArgument
	}

	const query = `DELETE FROM audit_logs WHERE created_at < $1`
	_, err := r.db.ExecContext(ctx, query, cutoff)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	return err
}
