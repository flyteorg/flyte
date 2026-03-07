package models

import (
	"database/sql"
	"encoding/json"
	"time"

	"gorm.io/datatypes"
)

// Action represents a workflow action in the database
// Root actions (where ParentActionName is NULL) represent runs
type Action struct {
	ID uint `gorm:"primaryKey" db:"id"`

	// Action Identifier (unique across org/project/domain/name)
	Org     string `gorm:"not null;uniqueIndex:idx_actions_identifier,priority:1;index:idx_actions_org" db:"org"`
	Project string `gorm:"not null;uniqueIndex:idx_actions_identifier,priority:2;index:idx_actions_project" db:"project"`
	Domain  string `gorm:"not null;uniqueIndex:idx_actions_identifier,priority:3;index:idx_actions_domain" db:"domain"`
	Name    string `gorm:"not null;uniqueIndex:idx_actions_identifier,priority:4" db:"name"`

	// Parent action (NULL for root actions/runs)
	ParentActionName *string `gorm:"index:idx_actions_parent" db:"parent_action_name"`

	// High-level status for quick queries/filtering.
	// Stores the proto ActionPhase enum integer value directly (e.g. 1 = QUEUED).
	Phase int32 `gorm:"not null;default:1;index:idx_actions_phase" db:"phase"`

	// Serialized protobuf messages
	// ActionSpec contains the full action specification proto
	ActionSpec datatypes.JSON `gorm:"type:jsonb" db:"action_spec"`

	// ActionDetails contains the full action details proto:
	// - ActionStatus (phase, timestamps, error, cache status, etc.)
	// - ActionAttempts array (all attempts with their phase transitions, cluster events, logs, etc.)
	// - Any other runtime state
	ActionDetails datatypes.JSON `gorm:"type:jsonb" db:"action_details"`

	// Timestamps
	// CreatedAt is set by the DB (NOW()) on insert — represents action start time.
	CreatedAt time.Time    `gorm:"not null;default:CURRENT_TIMESTAMP;index:idx_actions_created" db:"created_at"`
	UpdatedAt time.Time    `gorm:"not null;default:CURRENT_TIMESTAMP;index:idx_actions_updated" db:"updated_at"`
	// EndedAt is set when the action reaches a terminal phase.
	EndedAt   sql.NullTime `gorm:"index:idx_actions_ended" db:"ended_at"`
}

// TableName specifies the table name
func (Action) TableName() string { return "actions" }

// GetRunName extracts the run name from the action.
// For root actions (runs), returns the action's own name.
// For child actions, extracts from the ActionSpec JSON's action_id.run.name field.
func (a *Action) GetRunName() string {
	if a.ParentActionName == nil {
		return a.Name
	}

	if len(a.ActionSpec) == 0 {
		return ""
	}
	var spec struct {
		ActionID struct {
			Run struct {
				Name string `json:"name"`
			} `json:"run"`
		} `json:"action_id"`
	}
	if err := json.Unmarshal(a.ActionSpec, &spec); err != nil {
		return ""
	}
	return spec.ActionID.Run.Name
}

// Run is a type alias for Action (runs are just actions with ParentActionName == nil)
type Run = Action
