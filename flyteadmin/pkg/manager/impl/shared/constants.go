// Shared constants for the manager implementation.
package shared

// Field names for reference
const (
	Project               = "project"
	Domain                = "domain"
	Name                  = "name"
	ID                    = "id"
	Version               = "version"
	ResourceType          = "resource_type"
	Spec                  = "spec"
	Type                  = "type"
	RuntimeVersion        = "runtime version"
	Metadata              = "metadata"
	TypedInterface        = "typed interface"
	Image                 = "image"
	Limit                 = "limit"
	Filters               = "filters"
	ExpectedInputs        = "expected_inputs"
	FixedInputs           = "fixed_inputs"
	DefaultInputs         = "default_inputs"
	Inputs                = "inputs"
	State                 = "state"
	ExecutionID           = "execution_id"
	NodeID                = "node_id"
	NodeExecutionID       = "node_execution_id"
	TaskID                = "task_id"
	OccurredAt            = "occurred_at"
	Event                 = "event"
	ParentTaskExecutionID = "parent_task_execution_id"
	UserInputs            = "user_inputs"
	Attributes            = "attributes"
	MatchingAttributes    = "matching_attributes"
	// Parent of a node execution in the node executions table
	ParentID        = "parent_id"
	WorkflowClosure = "workflow_closure"
)
