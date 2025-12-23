package models

import (
	"database/sql"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"k8s.io/apimachinery/pkg/util/sets"
)

// TaskColumns are the allowed columns for task queries
var TaskColumns = sets.New(
	"org",
	"project",
	"domain",
	"name",
	"version",
	"environment",
	"function_name",
	"deployed_by",
	"created_at",
	"updated_at",
	"trigger_name",
	"total_triggers",
	"active_triggers",
	"env_description",
	"short_description",
)

// TaskVersionColumns are the allowed columns for task version queries
var TaskVersionColumns = sets.New(
	"version",
	"created_at",
)

// TaskName is a composite key representing a task
type TaskName struct {
	Org     string
	Project string
	Domain  string
	Name    string
}

// TaskKey is a composite key for a task
type TaskKey struct {
	Org     string
	Project string
	Domain  string
	Name    string
	Version string
}

// Tasks models the TaskDetails from the task_definition.proto
type Task struct {
	TaskKey `gorm:"-" db:"-"`

	// Extracted from Name
	Environment  string `gorm:"column:environment" db:"environment"`
	FunctionName string `gorm:"column:function_name" db:"function_name"`

	// Base fields
	CreatedAt time.Time `gorm:"column:created_at" db:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" db:"updated_at"`

	// Metadata
	DeployedBy            string         `gorm:"column:deployed_by" db:"deployed_by"`
	TriggerName           sql.NullString `gorm:"column:trigger_name" db:"trigger_name"`
	TotalTriggers         uint32         `gorm:"column:total_triggers" db:"total_triggers"`
	ActiveTriggers        uint32         `gorm:"column:active_triggers" db:"active_triggers"`
	TriggerAutomationSpec []byte         `gorm:"column:trigger_automation_spec" db:"trigger_automation_spec"`
	TriggerTypes          pgtype.Bits    `gorm:"column:trigger_types" db:"trigger_types"`
	EnvDescription        sql.NullString `gorm:"column:env_description" db:"env_description" json:"env_description,omitempty"`
	ShortDescription      sql.NullString `gorm:"column:short_description" db:"short_description" json:"short_description,omitempty"`

	// Spec
	TaskSpec []byte `gorm:"column:task_spec" db:"task_spec"`
}

type TaskCounts struct {
	// FilteredTotal is the number of tasks matching the applied filter
	FilteredTotal uint32 `db:"filtered_total"`
	// Total is the total number of tasks without any filters
	Total uint32 `db:"total"`
}

// TaskWithCounts extends Task with count information for pagination
type TaskWithCounts struct {
	Task
	TaskCounts
}

type TaskListResult struct {
	Tasks         []*Task
	FilteredTotal uint32
	Total         uint32
}

type TaskGroup struct {
	TaskName        string  `db:"task_name"`
	EnvironmentName *string `db:"environment_name"`
	TaskShortName   *string `db:"task_short_name"`

	ActionCount           int64      `db:"action_count"`
	LatestCreatedAt       time.Time  `db:"latest_created_at"`
	LastRunEndedAt        *time.Time `db:"last_run_ended_at"`
	AverageFailurePercent float64    `db:"average_failure_percent"`
	AverageDurationMs     *float64   `db:"average_duration_ms"`
	CreatedByList         []string   `db:"created_by_list"`
	RecentPhases          []int32    `db:"-"`

	// Error counts from action_events table by error kind
	UserErrorCount        int64 `db:"user_error_count"`
	SystemErrorCount      int64 `db:"system_error_count"`
	UnspecifiedErrorCount int64 `db:"unspecified_error_count"`
}

type RecentAction struct {
	TaskName string `db:"task_name"`
	Phase    int32  `db:"phase"`
}

type TaskGroupNotificationPayload struct {
	Org             string
	Project         string
	Domain          string
	TaskName        string
	EnvironmentName *string
}

// Models the TaskVersion response
type TaskVersion struct {
	Version   string    `db:"version"`
	CreatedAt time.Time `db:"created_at"`
}
