package models

import (
	"database/sql"
	"time"
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

	// Action type (task, trace, condition). Stores workflow.ActionType enum value.
	ActionType int32 `db:"action_type"`
	// Group this action belongs to, if applicable.
	ActionGroup sql.NullString `db:"action_group"`

	// Task reference columns
	TaskOrg     sql.NullString `db:"task_org"`
	TaskProject sql.NullString `db:"task_project"`
	TaskDomain  sql.NullString `db:"task_domain"`
	TaskName    sql.NullString `db:"task_name"`
	TaskVersion sql.NullString `db:"task_version"`

	// Task metadata columns
	TaskType        string         `db:"task_type"`
	TaskShortName   sql.NullString `db:"task_short_name"`
	FunctionName    string         `db:"function_name"`
	EnvironmentName sql.NullString `db:"environment_name"`

	// Serialized protobuf messages
	// ActionSpec contains the full action specification proto
	ActionSpec []byte `gorm:"type:bytea" db:"action_spec"`

	// ActionDetails contains the full action details proto:
	// - ActionStatus (phase, timestamps, error, cache status, etc.)
	// - ActionAttempts array (all attempts with their phase transitions, cluster events, logs, etc.)
	// - Any other runtime state
	ActionDetails []byte `gorm:"type:bytea" db:"action_details"`

	// DetailedInfo stores a serialized RunInfo proto containing metadata such as
	// the task spec digest (for looking up the resolved spec) and storage URIs.
	DetailedInfo []byte `gorm:"type:bytea" db:"detailed_info"`

	// RunSpec stores a serialized task.RunSpec proto (labels, annotations, envs,
	// interruptible, cluster, etc.) for this action's run.
	RunSpec []byte `gorm:"type:bytea" db:"run_spec"`

	// Timestamps
	// CreatedAt is set by the DB (NOW()) on insert — represents action start time.
	CreatedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP;index:idx_actions_created" db:"created_at"`
	UpdatedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP;index:idx_actions_updated" db:"updated_at"`
	// EndedAt is set when the action reaches a terminal phase.
	EndedAt    sql.NullTime  `gorm:"index:idx_actions_ended" db:"ended_at"`
	DurationMs sql.NullInt64 `db:"duration_ms"`
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
