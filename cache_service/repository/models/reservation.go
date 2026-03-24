package models

import "time"

type Reservation struct {
	Key              string    `gorm:"primaryKey;size:512" db:"key"`
	OwnerID          string    `gorm:"not null;index" db:"owner_id"`
	HeartbeatSeconds int64     `gorm:"not null" db:"heartbeat_seconds"`
	ExpiresAt        time.Time `gorm:"not null;index" db:"expires_at"`
	CreatedAt        time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" db:"created_at"`
	UpdatedAt        time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" db:"updated_at"`
}

func (Reservation) TableName() string { return "cache_service_reservations" }
