package models

import (
	"time"
)

type ExecutionEvent struct {
	BaseModel
	ExecutionKey
	RequestID  string
	OccurredAt time.Time
	Phase      string `gorm:"primary_key"`
}
