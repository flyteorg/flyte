package interfaces

import (
	"context"
	"time"

	runtime "github.com/flyteorg/flyte/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flytestdlib/storage"
)

//go:generate mockery -name=WorkflowExecutor -output=../mocks/ -case=underscore

type TaskResources struct {
	Defaults runtime.TaskResourceSet
	Limits   runtime.TaskResourceSet
}

type AttributesSource int

// Connection store the secret and config information required to connect to an external service.
type Connection struct {
	// Defines the type of the task that the connection is associated with.
	TaskType string
	// Defines the source of the attributes.
	Source AttributesSource
	// Defines a map of secrets that are used to connect to the external services.
	// The key is the name of the secret, such as openai_api_key, databricks_access_token, etc.
	// The value is the reference of the k8s or aws secret.
	Secrets map[string]string
	// Defines a map of configs that are used to connect to the external services, such as organization_id, workspace_id, etc.
	Configs map[string]string
}

type ExternalResourceAttributes struct {
	// Defines a map of connections that are used to connect to the external services.
	// The key is the name of the connection, such as openai, databricks, etc.
	// The value is the connection object.
	connections map[string]Connection
}

func (e *ExternalResourceAttributes) AddConnection(name string, conn Connection) {
	if e.connections == nil {
		e.connections = make(map[string]Connection)
	}
	e.connections[name] = conn
}

func (e *ExternalResourceAttributes) GetConnections() map[string]Connection {
	return e.connections
}

type ExecutionParameters struct {
	Inputs                     *core.LiteralMap
	AcceptedAt                 time.Time
	Labels                     map[string]string
	Annotations                map[string]string
	TaskPluginOverrides        []*admin.PluginOverride
	ExecutionConfig            *admin.WorkflowExecutionConfig
	RecoveryExecution          *core.WorkflowExecutionIdentifier
	TaskResources              *TaskResources
	EventVersion               int
	RoleNameKey                string
	RawOutputDataConfig        *admin.RawOutputDataConfig
	ClusterAssignment          *admin.ClusterAssignment
	ExecutionClusterLabel      *admin.ExecutionClusterLabel
	ExternalResourceAttributes *ExternalResourceAttributes
}

// ExecutionData includes all parameters required to create an execution CRD object.
type ExecutionData struct {
	// Execution namespace.
	Namespace string
	// Execution identifier.
	ExecutionID *core.WorkflowExecutionIdentifier
	// Underlying workflow name for the execution.
	ReferenceWorkflowName string
	// Launch plan name used to trigger the execution.
	ReferenceLaunchPlanName string
	// Compiled workflow closure used to build the flyte workflow
	WorkflowClosure *core.CompiledWorkflowClosure
	// Storage Data Reference for the WorkflowClosure
	WorkflowClosureReference storage.DataReference
	// Additional parameters used to build a workflow execution
	ExecutionParameters ExecutionParameters
	// Storage data reference of the execution inputs
	OffloadedInputsReference storage.DataReference
}

// ExecutionResponse is returned when a Flyte workflow execution is successfully created.
type ExecutionResponse struct {
	// Cluster identifier where the execution was created
	Cluster string
}

// AbortData includes all parameters required to abort an execution CRD object.
type AbortData struct {
	// Execution namespace.
	Namespace string
	// Execution identifier.
	ExecutionID *core.WorkflowExecutionIdentifier
	// Cluster identifier where the execution was created
	Cluster string
}

// WorkflowExecutor is a client interface used to create and delete Flyte workflow CRD objects.
type WorkflowExecutor interface {
	// ID returns the unique name of this executor implementation.
	ID() string
	// Execute creates a Flyte workflow execution CRD object.
	Execute(ctx context.Context, data ExecutionData) (ExecutionResponse, error)
	// Abort aborts a running Flyte workflow execution CRD object.
	Abort(ctx context.Context, data AbortData) error
}
