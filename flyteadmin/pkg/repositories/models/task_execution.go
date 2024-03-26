package models

import (
	"time"
)

// IMPORTANT: If you update the model below, be sure to double check model definitions in
// pkg/repositories/config/migration_models.go

// Task execution primary key
type TaskExecutionKey struct {
	TaskKey
	NodeExecutionKey
	// *IMPORTANT* This is a pointer to an int in order to allow setting an empty ("0") value according to gorm convention.
	// Because RetryAttempt is part of the TaskExecution primary key is should *never* be null.
	RetryAttempt *uint32 `gorm:"primary_key"`
}

// By convention, gorm foreign key references are of the form {ModelName}ID
type TaskExecution struct {
	BaseModel
	TaskExecutionKey
	Phase        string `valid:"length(0|255)"`
	PhaseVersion uint32
	InputURI     string `valid:"length(0|255)"`
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

var TaskExecutionColumns = modelColumns(TaskExecution{})
