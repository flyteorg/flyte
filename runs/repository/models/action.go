package models

import (
	"database/sql"
	"time"

	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
)

// ActionColumnsSet is the allowlist of columns that can be used in filters/sort for the actions table.
// This prevents SQL injection via user-supplied field names.
var ActionColumnsSet = sets.New(
	"project",
	"domain",
	"run_name",
	"phase",
	"run_source",
	"trigger_task_name",
	"trigger_name",
	"trigger_revision",
	"trigger_type",
	"task_project",
	"task_domain",
	"task_name",
	"task_version",
	"function_name",
	"created_at",
	"updated_at",
	"ended_at",
	"duration_ms",
)

// Action represents a workflow action in the database
// Root actions (where ParentActionName is NULL) represent runs
type Action struct {
	// Action Identifier (unique across project/domain/run_name/name)
	Project string `db:"project"`
	Domain  string `db:"domain"`
	RunName string `db:"run_name"`
	Name    string `db:"name"`

	// Parent action (NULL for root actions/runs)
	ParentActionName sql.NullString `db:"parent_action_name"`

	// High-level status for quick queries/filtering.
	// Stores the proto ActionPhase enum integer value directly (e.g. 1 = QUEUED).
	Phase int32 `db:"phase"`

	// Who initiated this run(web, CLI, scheduler, etc.)
	RunSource string `db:"run_source" json:"run_source,omitempty"`

	// Trigger fields — only set for runs created via RUN_SOURCE_SCHEDULE_TRIGGER.
	TriggerTaskName sql.NullString `db:"trigger_task_name"`
	TriggerName     sql.NullString `db:"trigger_name"`
	TriggerRevision sql.NullInt64  `db:"trigger_revision"`
	// TriggerType stores the automation type string (e.g. "TYPE_SCHEDULE").
	TriggerType sql.NullString `db:"trigger_type"`

	// Action type (task, trace, condition). Stores workflow.ActionType enum value.
	ActionType int32 `db:"action_type"`
	// Group this action belongs to, if applicable.
	ActionGroup sql.NullString `db:"action_group"`

	// Task reference columns
	TaskProject sql.NullString `db:"task_project"`
	TaskDomain  sql.NullString `db:"task_domain"`
	TaskName    sql.NullString `db:"task_name"`
	TaskVersion sql.NullString `db:"task_version"`

	// Task metadata columns
	TaskType        string         `db:"task_type"`
	TaskShortName   sql.NullString `db:"task_short_name"`
	FunctionName    string         `db:"function_name"`
	EnvironmentName sql.NullString `db:"environment_name"`

	// Serialized protobuf messages
	// ActionSpec contains the full action specification proto
	ActionSpec []byte `db:"action_spec"`

	// ActionDetails contains the full action details proto:
	// - ActionStatus (phase, timestamps, error, cache status, etc.)
	// - ActionAttempts array (all attempts with their phase transitions, cluster events, logs, etc.)
	// - Any other runtime state
	ActionDetails []byte `db:"action_details"`

	// DetailedInfo stores a serialized RunInfo proto containing metadata such as
	// the task spec digest (for looking up the resolved spec) and storage URIs.
	DetailedInfo []byte `db:"detailed_info"`

	// RunSpec stores a serialized task.RunSpec proto (labels, annotations, envs,
	// interruptible, cluster, etc.) for this action's run.
	RunSpec []byte `db:"run_spec"`

	// Abort tracking — set when a user requests abort; cleared once the pod is confirmed terminated.
	AbortRequestedAt  *time.Time `db:"abort_requested_at"`
	AbortAttemptCount int        `db:"abort_attempt_count"`
	AbortReason       *string    `db:"abort_reason"`

	// Timestamps
	// CreatedAt is set by the DB (NOW()) on insert — represents when the action was queued.
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
	// EndedAt is set when the action reaches a terminal phase.
	EndedAt     sql.NullTime            `db:"ended_at"`
	DurationMs  sql.NullInt64           `db:"duration_ms"`
	Attempts    uint32                  `db:"attempts" json:"attempts,omitempty"`
	CacheStatus core.CatalogCacheStatus `db:"cache_status" json:"cache_status,omitempty"`
}

// Clone returns an independent copy of the action, including pointer and JSON fields.
func (a *Action) Clone() *Action {
	if a == nil {
		return nil
	}

	cloned := *a
	if a.ParentActionName.Valid {
		cloned.ParentActionName = a.ParentActionName
	}
	if a.ActionSpec != nil {
		actionSpec := make([]byte, len(a.ActionSpec))
		copy(actionSpec, a.ActionSpec)
		cloned.ActionSpec = actionSpec
	}
	if a.ActionDetails != nil {
		actionDetails := make([]byte, len(a.ActionDetails))
		copy(actionDetails, a.ActionDetails)
		cloned.ActionDetails = actionDetails
	}

	return &cloned
}

// GetRunName extracts the run name from the action
// For root actions (runs), returns the action's own name
// For child actions, extracts from ActionSpec JSON
func (a *Action) GetRunName() string {
	if a.RunName != "" {
		return a.RunName
	}

	if !a.ParentActionName.Valid {
		// Root action - the run name is the action name
		return a.Name
	}
	return ""
}

func NewActionModel(actionID *common.ActionIdentifier) *Action {
	return &Action{
		Project: actionID.GetRun().GetProject(),
		Domain:  actionID.GetRun().GetDomain(),
		RunName: actionID.GetRun().GetName(),
		Name:    actionID.GetName(),
		Phase:   int32(common.ActionPhase_ACTION_PHASE_QUEUED),
	}
}

// Run is a type alias for Action (runs are just actions with ParentActionName == nil)
type Run = Action
