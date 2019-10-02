package config

import (
	"time"

	"github.com/lyft/flyteadmin/pkg/repositories/models"
)

/*
	IMPORTANT: You'll observe several models are redefined below with named index tags *omitted*. This is because
	postgres requires that index names be unique across *all* tables. If you modify Task, Execution, NodeExecution or
	TaskExecution models in code be sure to update the appropriate duplicate definitions here.
*/

type TaskKey struct {
	Project string `gorm:"primary_key"`
	Domain  string `gorm:"primary_key"`
	Name    string `gorm:"primary_key"`
	Version string `gorm:"primary_key"`
}

type ExecutionKey struct {
	Project string `gorm:"primary_key;column:execution_project"`
	Domain  string `gorm:"primary_key;column:execution_domain"`
	Name    string `gorm:"primary_key;column:execution_name"`
}

type NodeExecutionKey struct {
	ExecutionKey
	NodeID string `gorm:"primary_key;index"`
}

type NodeExecution struct {
	models.BaseModel
	NodeExecutionKey
	// Also stored in the closure, but defined as a separate column because it's useful for filtering and sorting.
	Phase     string
	InputURI  string
	Closure   []byte
	StartedAt *time.Time
	// Corresponds to the CreatedAt field in the NodeExecution closure
	// Prefixed with NodeExecution to avoid clashes with gorm.Model CreatedAt
	NodeExecutionCreatedAt *time.Time
	// Corresponds to the UpdatedAt field in the NodeExecution closure
	// Prefixed with NodeExecution to avoid clashes with gorm.Model UpdatedAt
	NodeExecutionUpdatedAt *time.Time
	Duration               time.Duration
	NodeExecutionEvents    []models.NodeExecutionEvent
	// The task execution (if any) which launched this node execution.
	ParentTaskExecutionID uint `sql:"default:null" gorm:"index"`
	// The workflow execution (if any) which this node execution launched
	LaunchedExecution models.Execution `gorm:"foreignkey:ParentNodeExecutionID"`
}

type TaskExecutionKey struct {
	TaskKey
	NodeExecutionKey
	// *IMPORTANT* This is a pointer to an int in order to allow setting an empty ("0") value according to gorm convention.
	// Because RetryAttempt is part of the TaskExecution primary key is should *never* be null.
	RetryAttempt *uint32 `gorm:"primary_key;AUTO_INCREMENT:FALSE"`
}

type TaskExecution struct {
	models.BaseModel
	TaskExecutionKey
	Phase        string
	PhaseVersion uint32
	InputURI     string
	Closure      []byte
	StartedAt    *time.Time
	// Corresponds to the CreatedAt field in the TaskExecution closure
	// This field is prefixed with TaskExecution because it signifies when
	// the execution was createdAt, not to be confused with gorm.Model.CreatedAt
	TaskExecutionCreatedAt *time.Time
	// Corresponds to the UpdatedAt field in the TaskExecution closure
	// This field is prefixed with TaskExecution because it signifies when
	// the execution was UpdatedAt, not to be confused with gorm.Model.UpdatedAt
	TaskExecutionUpdatedAt *time.Time
	Duration               time.Duration
	// The child node executions (if any) launched by this task execution.
	ChildNodeExecution []NodeExecution `gorm:"foreignkey:ParentTaskExecutionID"`
}
