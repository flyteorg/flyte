package repository

import (
	"time"

	"gorm.io/datatypes"
)

// Run represents a workflow run in the database
type Run struct {
	ID uint `gorm:"primaryKey"`

	// Run Identifier
	Org     string `gorm:"not null;uniqueIndex:idx_runs_identifier,priority:1"`
	Project string `gorm:"not null;uniqueIndex:idx_runs_identifier,priority:2"`
	Domain  string `gorm:"not null;uniqueIndex:idx_runs_identifier,priority:3"`
	Name    string `gorm:"not null;uniqueIndex:idx_runs_identifier,priority:4"`

	// Root action reference
	RootActionName string `gorm:"not null"`

	// Trigger reference (if applicable)
	TriggerOrg     *string
	TriggerProject *string
	TriggerDomain  *string
	TriggerName    *string

	// Run spec (serialized)
	RunSpec datatypes.JSON `gorm:"not null"`

	// Metadata
	CreatedByPrincipal          *string
	CreatedByK8sServiceAccount  *string

	// Timestamps
	CreatedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
}

// Action represents an action within a run
type Action struct {
	ID uint `gorm:"primaryKey"`

	// Action Identifier
	Org     string `gorm:"not null;uniqueIndex:idx_actions_identifier,priority:1"`
	Project string `gorm:"not null;uniqueIndex:idx_actions_identifier,priority:2"`
	Domain  string `gorm:"not null;uniqueIndex:idx_actions_identifier,priority:3"`
	RunName string `gorm:"not null;uniqueIndex:idx_actions_identifier,priority:4;index:idx_actions_run"`
	Name    string `gorm:"not null;uniqueIndex:idx_actions_identifier,priority:5"`

	// Foreign key to run
	RunID uint `gorm:"not null;index:idx_actions_run_id"`
	Run   *Run `gorm:"foreignKey:RunID;constraint:OnDelete:CASCADE"`

	// Parent action (for nested actions)
	ParentActionName *string `gorm:"index:idx_actions_parent"`

	// Action type and metadata
	ActionType   string  `gorm:"not null;index:idx_actions_type"` // task, trace, condition
	ActionGroup  *string

	// Task action metadata (nullable)
	TaskOrg       *string
	TaskProject   *string
	TaskDomain    *string
	TaskName      *string
	TaskVersion   *string
	TaskType      *string
	TaskShortName *string

	// Trace action metadata (nullable)
	TraceName *string

	// Condition action metadata (nullable)
	ConditionName       *string
	ConditionScopeType  *string // run_id, action_id, global
	ConditionScopeValue *string

	// Execution metadata
	ExecutedByPrincipal         *string
	ExecutedByK8sServiceAccount *string

	// Status
	Phase      string     `gorm:"not null;default:'PHASE_QUEUED';index:idx_actions_phase"`
	StartTime  *time.Time
	EndTime    *time.Time
	AttemptsCount uint    `gorm:"not null;default:0"`

	// Cache status
	CacheStatus *string

	// Error info (if failed)
	ErrorKind    *string
	ErrorMessage *string

	// Abort info (if aborted)
	AbortReason                  *string
	AbortedByPrincipal           *string
	AbortedByK8sServiceAccount   *string

	// Serialized specs
	TaskSpec      datatypes.JSON
	TraceSpec     datatypes.JSON
	ConditionSpec datatypes.JSON

	// Input/Output references
	InputURI      *string
	RunOutputBase *string

	// Timestamps
	CreatedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
}

// ActionAttempt represents a single attempt of an action
type ActionAttempt struct {
	ID uint `gorm:"primaryKey"`

	// Foreign key to action
	ActionID uint    `gorm:"not null;uniqueIndex:idx_attempts_action_attempt,priority:1"`
	Action   *Action `gorm:"foreignKey:ActionID;constraint:OnDelete:CASCADE"`

	// Attempt number (1-indexed)
	AttemptNumber uint `gorm:"not null;uniqueIndex:idx_attempts_action_attempt,priority:2"`

	// Phase tracking
	Phase     string     `gorm:"not null"`
	StartTime time.Time  `gorm:"not null"`
	EndTime   *time.Time

	// Error info (if failed)
	ErrorKind    *string
	ErrorMessage *string

	// Output references (JSON)
	Outputs datatypes.JSON

	// Logs (JSON array)
	LogInfo datatypes.JSON

	// Logs available flag
	LogsAvailable bool `gorm:"default:false"`

	// Cache status
	CacheStatus *string

	// Cluster assignment
	Cluster *string

	// Log context (k8s pod/container info) (JSON)
	LogContext datatypes.JSON

	// Timestamps
	CreatedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
}

// ClusterEvent represents a cluster event (e.g., k8s event)
type ClusterEvent struct {
	ID uint `gorm:"primaryKey"`

	// Foreign key to attempt
	AttemptID uint            `gorm:"not null;index:idx_cluster_events_attempt"`
	Attempt   *ActionAttempt  `gorm:"foreignKey:AttemptID;constraint:OnDelete:CASCADE"`

	// Event details
	OccurredAt time.Time `gorm:"not null"`
	Message    string    `gorm:"not null"`

	// Timestamps
	CreatedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
}

// PhaseTransition represents a phase change in an attempt
type PhaseTransition struct {
	ID uint `gorm:"primaryKey"`

	// Foreign key to attempt
	AttemptID uint           `gorm:"not null;index:idx_phase_transitions_attempt"`
	Attempt   *ActionAttempt `gorm:"foreignKey:AttemptID;constraint:OnDelete:CASCADE"`

	// Phase transition
	Phase     string     `gorm:"not null"`
	StartTime time.Time  `gorm:"not null"`
	EndTime   *time.Time

	// Timestamps
	CreatedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
}

// TableName specifies the table names
func (Run) TableName() string             { return "runs" }
func (Action) TableName() string          { return "actions" }
func (ActionAttempt) TableName() string   { return "action_attempts" }
func (ClusterEvent) TableName() string    { return "cluster_events" }
func (PhaseTransition) TableName() string { return "phase_transitions" }
