package repository

import (
	"time"

	"gorm.io/datatypes"
)

// QueuedAction represents a queued action in the database
type QueuedAction struct {
	ID uint `gorm:"primaryKey"`

	// Action Identifier
	Org        string `gorm:"not null;index:idx_queued_actions_identifier,priority:1"`
	Project    string `gorm:"not null;index:idx_queued_actions_identifier,priority:2"`
	Domain     string `gorm:"not null;index:idx_queued_actions_identifier,priority:3"`
	RunName    string `gorm:"not null;index:idx_queued_actions_identifier,priority:4;index:idx_queued_actions_run"`
	ActionName string `gorm:"not null;index:idx_queued_actions_identifier,priority:5"`

	// Parent reference
	ParentActionName *string `gorm:"index:idx_queued_actions_parent"`

	// Action group
	ActionGroup *string

	// Subject who created the action
	Subject *string

	// Serialized action spec (protobuf as JSON)
	ActionSpec datatypes.JSON `gorm:"not null"`

	// Input/Output paths
	InputURI       string `gorm:"not null"`
	RunOutputBase  string `gorm:"not null"`

	// Queue metadata
	EnqueuedAt  time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP;index:idx_queued_actions_enqueued"`
	ProcessedAt *time.Time

	// Status tracking
	Status       string `gorm:"not null;default:'queued';index:idx_queued_actions_status"` // queued, processing, completed, aborted, failed
	AbortReason  *string
	ErrorMessage *string

	// Timestamps
	CreatedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
}

// TableName specifies the table name for QueuedAction
func (QueuedAction) TableName() string {
	return "queued_actions"
}
