package models

import (
	"time"
)

type ExecutionEvent struct {
	BaseModel
	ExecutionKey
	RequestID  string `valid:"length(0|255)"`
	OccurredAt time.Time
	Phase      string `gorm:"primary_key"`
}
