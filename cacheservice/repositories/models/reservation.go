package models

import (
	"time"
)

type Reservation struct {
	Key       string
	OwnerID   string
	ExpiresAt time.Time
}
