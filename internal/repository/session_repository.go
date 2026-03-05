package repository

import (
	"context"
	"database/sql"
	"errors"

	"auth-service/internal/domain"

	"github.com/google/uuid"
)

type sessionRepository struct {
	db *sql.DB
}

func NewSessionRepository(db *sql.DB) *sessionRepository {
	return &sessionRepository{db: db}
}

func (r *sessionRepository) Create(ctx context.Context, session *domain.Session) error {
	query := `
        INSERT INTO sessions (user_id, token_hash, ip_address, user_agent, expires_at, revoked)
        VALUES ($1, $2, $3, $4, $5, $6)
        RETURNING id, created_at
    `
	row := r.db.QueryRowContext(ctx, query,
		session.UserID, session.TokenHash, session.IPAddress, session.UserAgent,
		session.ExpiresAt, session.Revoked,
	)
	err := row.Scan(&session.ID, &session.CreatedAt)
	return err
}

func (r *sessionRepository) FindByTokenHash(ctx context.Context, tokenHash string) (*domain.Session, error) {
	var session domain.Session
	query := `SELECT id, user_id, token_hash, ip_address, user_agent, expires_at, revoked, created_at
              FROM sessions WHERE token_hash = $1`
	row := r.db.QueryRowContext(ctx, query, tokenHash)
	err := row.Scan(
		&session.ID, &session.UserID, &session.TokenHash, &session.IPAddress,
		&session.UserAgent, &session.ExpiresAt, &session.Revoked, &session.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *sessionRepository) FindActiveByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Session, error) {
	query := `SELECT id, user_id, token_hash, ip_address, user_agent, expires_at, revoked, created_at
              FROM sessions WHERE user_id = $1 AND revoked = false AND expires_at > NOW()`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []domain.Session
	for rows.Next() {
		var s domain.Session
		err := rows.Scan(
			&s.ID, &s.UserID, &s.TokenHash, &s.IPAddress,
			&s.UserAgent, &s.ExpiresAt, &s.Revoked, &s.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, s)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return sessions, nil
}

func (r *sessionRepository) Revoke(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE sessions SET revoked = true WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *sessionRepository) RevokeAllForUser(ctx context.Context, userID uuid.UUID) error {
	query := `UPDATE sessions SET revoked = true WHERE user_id = $1`
	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}

func (r *sessionRepository) DeleteExpired(ctx context.Context) error {
	query := `DELETE FROM sessions WHERE expires_at < NOW()`
	_, err := r.db.ExecContext(ctx, query)
	return err
}
