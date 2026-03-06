package repository

import (
	"context"
	"database/sql"
	"errors"

	"auth-service/internal/domain"

	"github.com/google/uuid"
)

type SessionRepository struct {
	db *sql.DB
}

func NewSessionRepository(db *sql.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

func (r *SessionRepository) Create(ctx context.Context, session *domain.Session) error {
	if session == nil || session.UserID == uuid.Nil || session.TokenHash == "" {
		return ErrInvalidArgument
	}

	const query = `INSERT INTO sessions (user_id, token_hash, ip_address, user_agent, expires_at, revoked)
        VALUES ($1, $2, $3, $4, $5, $6)
        RETURNING id, created_at`

	row := r.db.QueryRowContext(ctx, query,
		session.UserID,
		session.TokenHash,
		session.IPAddress,
		session.UserAgent,
		session.ExpiresAt,
		session.Revoked,
	)
	return row.Scan(&session.ID, &session.CreatedAt)
}

func (r *SessionRepository) ByTokenHash(ctx context.Context, tokenHash string) (domain.Session, error) {
	if tokenHash == "" {
		return domain.Session{}, ErrInvalidArgument
	}

	const query = `SELECT id, user_id, token_hash, ip_address, user_agent, expires_at, revoked, created_at
        FROM sessions WHERE token_hash = $1`

	var session domain.Session
	err := r.db.QueryRowContext(ctx, query, tokenHash).Scan(
		&session.ID,
		&session.UserID,
		&session.TokenHash,
		&session.IPAddress,
		&session.UserAgent,
		&session.ExpiresAt,
		&session.Revoked,
		&session.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Session{}, ErrNotFound
	}
	if err != nil {
		return domain.Session{}, err
	}
	return session, nil
}

func (r *SessionRepository) ActiveByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Session, error) {
	if userID == uuid.Nil {
		return nil, ErrInvalidArgument
	}

	const query = `SELECT id, user_id, token_hash, ip_address, user_agent, expires_at, revoked, created_at
        FROM sessions WHERE user_id = $1 AND revoked = false AND expires_at > NOW()`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []domain.Session
	for rows.Next() {
		var s domain.Session
		if err := rows.Scan(
			&s.ID,
			&s.UserID,
			&s.TokenHash,
			&s.IPAddress,
			&s.UserAgent,
			&s.ExpiresAt,
			&s.Revoked,
			&s.CreatedAt,
		); err != nil {
			return nil, err
		}
		sessions = append(sessions, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return sessions, nil
}

func (r *SessionRepository) Revoke(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return ErrInvalidArgument
	}

	const query = `UPDATE sessions SET revoked = true WHERE id = $1`
	res, err := r.db.ExecContext(ctx, query, id)
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

func (r *SessionRepository) RevokeAllForUser(ctx context.Context, userID uuid.UUID) error {
	if userID == uuid.Nil {
		return ErrInvalidArgument
	}

	const query = `UPDATE sessions SET revoked = true WHERE user_id = $1`
	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}

func (r *SessionRepository) DeleteExpired(ctx context.Context) error {
	const query = `DELETE FROM sessions WHERE expires_at < NOW()`
	_, err := r.db.ExecContext(ctx, query)
	return err
}
