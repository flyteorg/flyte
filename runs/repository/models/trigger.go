package models

import (
	"database/sql"
	"time"

	"k8s.io/apimachinery/pkg/util/sets"
)

// TriggerColumns are the allowed columns for trigger list/sort queries.
var TriggerColumns = sets.New(
	"project", "domain", "task_name", "name",
	"active", "deployed_at", "updated_at",
	"task_version", "automation_type",
)

// TriggerRevisionColumns are the allowed columns for trigger revision list/sort queries.
var TriggerRevisionColumns = sets.New(
	"revision", "deployed_at", "updated_at", "action",
)

// Trigger is the mutable latest-state row for a trigger.
// One row per (project, domain, task_name, name); updated in-place on each deploy/activate/delete.
// Mirrors the pattern used by Action (latest state) + ActionEvent (history).
type Trigger struct {
	ID uint `db:"id"`

	// Trigger identity — unique constraint drives upsert
	Project  string `db:"project"`
	Domain   string `db:"domain"`
	TaskName string `db:"task_name"`
	Name     string `db:"name"`

	// Monotonically increasing counter; incremented on every write
	LatestRevision uint64 `db:"latest_revision"`

	// Serialized protos
	Spec           []byte `db:"spec"`
	AutomationSpec []byte `db:"automation_spec"`

	// Denormalized fields for cheap queries without deserializing protos
	TaskVersion    string `db:"task_version"`
	Active         bool   `db:"active"`
	AutomationType string `db:"automation_type"`

	// Identity
	DeployedBy sql.NullString `db:"deployed_by"`
	UpdatedBy  sql.NullString `db:"updated_by"`

	// Timestamps
	DeployedAt  time.Time    `db:"deployed_at"`
	UpdatedAt   time.Time    `db:"updated_at"`
	TriggeredAt sql.NullTime `db:"triggered_at"`
	DeletedAt   sql.NullTime `db:"deleted_at"`

	Description sql.NullString `db:"description"`
}

// ToTaskKey returns the TaskKey this trigger is attached to.
func (t Trigger) ToTaskKey() TaskKey {
	return TaskKey{
		Project: t.Project,
		Domain:  t.Domain,
		Name:    t.TaskName,
		Version: t.TaskVersion,
	}
}

// TriggerRevision is an immutable snapshot appended on every state change.
// Mirrors the ActionEvent (append-only history) pattern.
type TriggerRevision struct {
	// Composite PK: trigger identity + revision number
	Project  string `db:"project"`
	Domain   string `db:"domain"`
	TaskName string `db:"task_name"`
	Name     string `db:"name"`
	Revision uint64 `db:"revision"`

	// Snapshot of state at this revision
	Spec           []byte `db:"spec"`
	AutomationSpec []byte `db:"automation_spec"`

	TaskVersion    string `db:"task_version"`
	Active         bool   `db:"active"`
	AutomationType string `db:"automation_type"`

	DeployedBy sql.NullString `db:"deployed_by"`
	UpdatedBy  sql.NullString `db:"updated_by"`

	DeployedAt  time.Time    `db:"deployed_at"`
	UpdatedAt   time.Time    `db:"updated_at"`
	TriggeredAt sql.NullTime `db:"triggered_at"`
	DeletedAt   sql.NullTime `db:"deleted_at"`

	// What action produced this revision row
	// Values: TRIGGER_REVISION_ACTION_DEPLOY, _ACTIVATE, _DEACTIVATE, _DELETE
	Action string `db:"action"`

	CreatedAt time.Time `db:"created_at"`
}
