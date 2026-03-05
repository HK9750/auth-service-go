package domain

import (
	"time"

	"github.com/google/uuid"
)

type Role struct {
	ID          uuid.UUID `db:"id"`
	Name        string    `db:"name"`
	Description *string   `db:"description"`
	CreatedAt   time.Time `db:"created_at"`
}

type UserRole struct {
	UserID     uuid.UUID `db:"user_id"`
	RoleID     uuid.UUID `db:"role_id"`
	AssignedAt time.Time `db:"assigned_at"`
}
