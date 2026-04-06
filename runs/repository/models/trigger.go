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
	ID uint `gorm:"primaryKey;autoIncrement"`

	// Trigger identity — unique constraint drives upsert
	Project  string `gorm:"not null;uniqueIndex:idx_triggers_name,priority:1"`
	Domain   string `gorm:"not null;uniqueIndex:idx_triggers_name,priority:2"`
	TaskName string `gorm:"not null;uniqueIndex:idx_triggers_name,priority:3"`
	Name     string `gorm:"not null;uniqueIndex:idx_triggers_name,priority:4"`

	// Monotonically increasing counter; incremented on every write
	LatestRevision uint64 `gorm:"not null;default:1"`

	// Serialized protos
	Spec           []byte `gorm:"type:bytea;not null"`
	AutomationSpec []byte `gorm:"type:bytea"`

	// Denormalized fields for cheap queries without deserializing protos
	TaskVersion    string `gorm:"not null"`
	Active         bool   `gorm:"not null;index:idx_triggers_schedule_lookup,priority:1"`
	AutomationType string `gorm:"not null;default:'TYPE_NONE';index:idx_triggers_schedule_lookup,priority:2"`

	// Identity
	DeployedBy sql.NullString `gorm:"default:null"`
	UpdatedBy  sql.NullString `gorm:"default:null"`

	// Timestamps
	DeployedAt  time.Time    `gorm:"not null"`
	UpdatedAt   time.Time    `gorm:"not null"`
	TriggeredAt sql.NullTime `gorm:"default:null"`
	DeletedAt   sql.NullTime `gorm:"default:null"`

	Description sql.NullString `gorm:"default:null"`
}

func (Trigger) TableName() string { return "triggers" }

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
	Project  string `gorm:"not null;primaryKey"`
	Domain   string `gorm:"not null;primaryKey"`
	TaskName string `gorm:"not null;primaryKey"`
	Name     string `gorm:"not null;primaryKey"`
	Revision uint64 `gorm:"not null;primaryKey"`

	// Snapshot of state at this revision
	Spec           []byte `gorm:"type:bytea;not null"`
	AutomationSpec []byte `gorm:"type:bytea"`

	TaskVersion    string `gorm:"not null"`
	Active         bool   `gorm:"not null"`
	AutomationType string `gorm:"not null;default:'TYPE_NONE'"`

	DeployedBy sql.NullString `gorm:"default:null"`
	UpdatedBy  sql.NullString `gorm:"default:null"`

	DeployedAt  time.Time    `gorm:"not null"`
	UpdatedAt   time.Time    `gorm:"not null"`
	TriggeredAt sql.NullTime `gorm:"default:null"`
	DeletedAt   sql.NullTime `gorm:"default:null"`

	// What action produced this revision row
	// Values: TRIGGER_REVISION_ACTION_DEPLOY, _ACTIVATE, _DEACTIVATE, _DELETE
	Action string `gorm:"not null"`

	CreatedAt time.Time `gorm:"not null;autoCreateTime"`
}

func (TriggerRevision) TableName() string { return "trigger_revisions" }
