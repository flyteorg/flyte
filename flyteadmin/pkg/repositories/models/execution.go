package models

import (
	"time"

	"github.com/flyteorg/flytestdlib/storage"
)

// IMPORTANT: If you update the model below, be sure to double check model definitions in
// pkg/repositories/config/migration_models.go

// Execution primary key
type ExecutionKey struct {
	Project string `gorm:"primary_key;column:execution_project" valid:"length(0|255)"`
	Domain  string `gorm:"primary_key;column:execution_domain" valid:"length(0|255)"`
	Name    string `gorm:"primary_key;column:execution_name" valid:"length(0|255)"`
}

// Database model to encapsulate a (workflow) execution.
type Execution struct {
	BaseModel
	ExecutionKey
	LaunchPlanID uint   `gorm:"index"`
	WorkflowID   uint   `gorm:"index"`
	TaskID       uint   `gorm:"index"`
	Phase        string `valid:"length(0|255)"`
	Closure      []byte
	Spec         []byte `gorm:"not null"`
	StartedAt    *time.Time
	// Corresponds to the CreatedAt field in the Execution closure.
	// Prefixed with Execution to avoid clashes with gorm.Model CreatedAt
	ExecutionCreatedAt *time.Time
	// Corresponds to the UpdatedAt field in the Execution closure
	// Prefixed with Execution to avoid clashes with gorm.Model UpdatedAt
	ExecutionUpdatedAt *time.Time
	Duration           time.Duration
	// In the case of an aborted execution this string may be non-empty.
	// It should be ignored for any other value of phase other than aborted.
	AbortCause string `valid:"length(0|255)"`
	// Corresponds to the execution mode used to trigger this execution
	Mode int32
	// The "parent" execution (if there is one) that is related to this execution.
	SourceExecutionID uint
	// The parent node execution if this was launched by a node
	ParentNodeExecutionID uint
	// Cluster where execution was triggered
	Cluster string `valid:"length(0|255)"`
	// Offloaded location of inputs LiteralMap. These are the inputs evaluated and contain applied defaults.
	InputsURI storage.DataReference
	// User specified inputs. This map might be incomplete and not include defaults applied
	UserInputsURI storage.DataReference
	// Execution Error Kind. nullable
	ErrorKind *string `gorm:"index"`
	// Execution Error Code nullable
	ErrorCode *string `valid:"length(0|255)"`
	// The user responsible for launching this execution.
	// This is also stored in the spec but promoted as a column for filtering.
	User string `gorm:"index" valid:"length(0|255)"`
	// GORM doesn't save the zero value for ints, so we use a pointer for the State field
	State *int32 `gorm:"default:0"`
}
