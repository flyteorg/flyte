package models

import (
	"time"

	"gorm.io/datatypes"
)

// Action represents a workflow action in the database
// Root actions (where ParentActionName is NULL) represent runs
type Action struct {
	ID uint `gorm:"primaryKey"`

	// Action Identifier (unique across org/project/domain/name)
	Org     string `gorm:"not null;uniqueIndex:idx_actions_identifier,priority:1;index:idx_actions_org"`
	Project string `gorm:"not null;uniqueIndex:idx_actions_identifier,priority:2;index:idx_actions_project"`
	Domain  string `gorm:"not null;uniqueIndex:idx_actions_identifier,priority:3;index:idx_actions_domain"`
	Name    string `gorm:"not null;uniqueIndex:idx_actions_identifier,priority:4"`

	// Parent action (NULL for root actions/runs)
	ParentActionName *string `gorm:"index:idx_actions_parent"`

	// High-level status for quick queries/filtering
	Phase string `gorm:"not null;default:'PHASE_QUEUED';index:idx_actions_phase"`

	// Serialized protobuf messages
	// ActionSpec contains the full action specification proto
	ActionSpec datatypes.JSON `gorm:"type:jsonb"`

	// ActionDetails contains the full action details proto:
	// - ActionStatus (phase, timestamps, error, cache status, etc.)
	// - ActionAttempts array (all attempts with their phase transitions, cluster events, logs, etc.)
	// - Any other runtime state
	ActionDetails datatypes.JSON `gorm:"type:jsonb"`

	// Timestamps
	CreatedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP;index:idx_actions_created"`
	UpdatedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP;index:idx_actions_updated"`
}

// TableName specifies the table name
func (Action) TableName() string { return "actions" }

// GetRunName extracts the run name from the action
// For root actions (runs), returns the action's own name
// For child actions, extracts from ActionSpec JSON
func (a *Action) GetRunName() string {
	if a.ParentActionName == nil {
		// Root action - the run name is the action name
		return a.Name
	}

	// TODO: Extract run name from ActionSpec JSON
	// For now, return empty string as placeholder
	return ""
}

// Run is a type alias for Action (runs are just actions with ParentActionName == nil)
type Run = Action
