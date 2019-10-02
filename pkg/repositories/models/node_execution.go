package models

import (
	"time"
)

// IMPORTANT: If you update the model below, be sure to double check model definitions in
// pkg/repositories/config/migration_models.go

type NodeExecutionKey struct {
	ExecutionKey
	NodeID string `gorm:"primary_key;index"`
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
	NodeExecutionEvents    []NodeExecutionEvent
	// The task execution (if any) which launched this node execution.
	ParentTaskExecutionID uint `sql:"default:null" gorm:"index"`
	// The workflow execution (if any) which this node execution launched
	LaunchedExecution Execution `gorm:"foreignkey:ParentNodeExecutionID"`
}
