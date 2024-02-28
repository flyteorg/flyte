package models

import (
	"time"
)

type Reservation struct {
	Key       string    `dynamodbav:"id"`
	OwnerID   string    `dynamodbav:"ownerId"`
	ExpiresAt time.Time `dynamodbav:"expiresAt"`
}
