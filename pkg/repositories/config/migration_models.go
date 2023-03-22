package config

import (
	"time"

	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
)

/*
	IMPORTANT: You'll observe several models are redefined below with named index tags *omitted*. This is because
	postgres requires that index names be unique across *all* tables. If you modify Task, Execution, NodeExecution or
	TaskExecution models in code be sure to update the appropriate duplicate definitions here.
	That is, in the actual code, it makes more sense to re-use structs, like how NodeExecutionKey is in both NodeExecution
	and in TaskExecution. But simply re-using in migrations would result in indices with the same name.
	In the new model where all models are replicated in each function, this is not an issue.
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
	// The task execution (if any) which launched this node execution.
	ParentTaskExecutionID uint `sql:"default:null" gorm:"index"`
	// The workflow execution (if any) which this node execution launched
	LaunchedExecution models.Execution `gorm:"foreignKey:ParentNodeExecutionID;references:ID"`
	// In the case of dynamic workflow nodes, the remote closure is uploaded to the path specified here.
	DynamicWorkflowRemoteClosureReference string
	// Metadata that is only relevant to the flyteadmin service that is used to parse the model and track additional attributes.
	InternalData []byte
}

type TaskExecutionKey struct {
	TaskKey
	Project string `gorm:"primary_key;column:execution_project;index:idx_task_executions_exec"`
	Domain  string `gorm:"primary_key;column:execution_domain;index:idx_task_executions_exec"`
	Name    string `gorm:"primary_key;column:execution_name;index:idx_task_executions_exec"`
	NodeID  string `gorm:"primary_key;index:idx_task_executions_exec;index"`
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
	ChildNodeExecution []NodeExecution `gorm:"foreignkey:ParentTaskExecutionID;references:ID"`
}
