package models

import (
	"time"
)

type NodeExecutionEvent struct {
	BaseModel
	NodeExecutionKey
	RequestID  string
	OccurredAt time.Time
	Phase      string `gorm:"primary_key"`
}

var NodeExecutionEventColumns = modelColumns(NodeExecutionEvent{})
