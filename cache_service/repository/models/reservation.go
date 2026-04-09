package models

import "time"

type Reservation struct {
	Key              string    `db:"key"`
	OwnerID          string    `db:"owner_id"`
	HeartbeatSeconds int64     `db:"heartbeat_seconds"`
	ExpiresAt        time.Time `db:"expires_at"`
	CreatedAt        time.Time `db:"created_at"`
	UpdatedAt        time.Time `db:"updated_at"`
}
