package models

import (
	"time"
)

// IMPORTANT: If you update the model below, be sure to double check model definitions in
// pkg/repositories/config/migration_models.go

type NodeExecutionKey struct {
	ExecutionKey
	NodeID string `gorm:"primary_key;index" valid:"length(0|255)"`
}

// By convention, gorm foreign key references are of the form {ModelName}ID
type NodeExecution struct {
	BaseModel
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
	// Metadata about the node execution.
	NodeExecutionMetadata []byte
	// Parent that spawned this node execution - value is empty for executions at level 0
	ParentID *uint `sql:"default:null" gorm:"index"`
	// List of child node executions - for cases like Dynamic task, sub workflow, etc
	ChildNodeExecutions []NodeExecution `gorm:"foreignKey:ParentID;references:ID"`
	// The task execution (if any) which launched this node execution.
	// TO BE DEPRECATED - as we have now introduced ParentID
	ParentTaskExecutionID *uint `sql:"default:null" gorm:"index"`
	// The workflow execution (if any) which this node execution launched
	// NOTE: LaunchedExecution[foreignkey:ParentNodeExecutionID] refers to Workflow execution launched and is different from ParentID
	LaunchedExecution Execution `gorm:"foreignKey:ParentNodeExecutionID;references:ID"`
	// Execution Error Kind. nullable, can be one of core.ExecutionError_ErrorKind
	ErrorKind *string `gorm:"index"`
	// Execution Error Code nullable. string value, but finite set determined by the execution engine and plugins
	ErrorCode *string
	// If the node is of Type Task, this should always exist for a successful execution, indicating the cache status for the execution
	CacheStatus *string
	// In the case of dynamic workflow nodes, the remote closure is uploaded to the path specified here.
	DynamicWorkflowRemoteClosureReference string
	// Metadata that is only relevant to the flyteadmin service that is used to parse the model and track additional attributes.
	InternalData []byte
}

var NodeExecutionColumns = modelColumns(NodeExecution{})
