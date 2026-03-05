package domain

import (
	"net"
	"time"

	"github.com/google/uuid"
)

type AuditLog struct {
	ID        uuid.UUID  `db:"id"`
	UserID    *uuid.UUID `db:"user_id"`
	Action    string     `db:"action"`
	Resource  *string    `db:"resource"`
	Metadata  []byte     `db:"metadata"`
	IPAddress net.IP     `db:"ip_address"`
	CreatedAt time.Time  `db:"created_at"`
}
