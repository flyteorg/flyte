package models

import (
	"database/sql"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"k8s.io/apimachinery/pkg/util/sets"
)

// TaskColumns are the allowed columns for task queries
var TaskColumns = sets.New(
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
	Project string
	Domain  string
	Name    string
}

// TaskKey is a composite key for a task
type TaskKey struct {
	Project string `db:"project"`
	Domain  string `db:"domain"`
	Name    string `db:"name"`
	Version string `db:"version"`
}

// Tasks models the TaskDetails from the task_definition.proto
type Task struct {
	TaskKey

	// Extracted from Name
	Environment  string `db:"environment"`
	FunctionName string `db:"function_name"`

	// Base fields
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`

	// Metadata
	DeployedBy            string         `db:"deployed_by"`
	TriggerName           sql.NullString `db:"trigger_name"`
	TotalTriggers         uint32         `db:"total_triggers"`
	ActiveTriggers        uint32         `db:"active_triggers"`
	TriggerAutomationSpec []byte         `db:"trigger_automation_spec"`
	TriggerTypes          pgtype.Bits    `db:"trigger_types"`
	EnvDescription        sql.NullString `db:"env_description" json:"env_description,omitempty"`
	ShortDescription      sql.NullString `db:"short_description" json:"short_description,omitempty"`

	// Spec
	TaskSpec []byte `db:"task_spec"`
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
