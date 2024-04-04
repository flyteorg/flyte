package models

import (
	"time"
)

type CacheReservation struct {
	Key       string `gorm:"primary_key"`
	OwnerID   string
	ExpiresAt time.Time
}
