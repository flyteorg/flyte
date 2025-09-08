package interfaces

import (
	"context"
	"time"

	v1 "k8s.io/api/core/v1"

	idlcore "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core"

	"github.com/unionai/flyte/fasttask/plugin/pb"
)

//go:generate mockery-v2 --all --case=underscore --with-expecter --output=./mocks

/**
 * `state` refers to the current status of either an environment or worker. This value dictates
 *  which operations are allowed to be performed on the entity.
 *
 *  `Environment`
 *  - HEALTHY: The environment has fully populated metadata (ex. executionEnvID,
 *    fastTaskEnvironmentSpec). It **can accept new task executions** and **can scale up / down**.
 *  - ORPHANED: The environment has not been fully populated (ex. missing executionEnvID and / or
 *    fastTaskEnviornmentSpec - or some subset). This state occurs when recovering from a failure,
 *    where the environment is created either from orphan detection or from a worker heartbeat. It
 *    **can accept new task executions** and **can scale down** but **cannot scale up**. If the
 *    environment is not active (ex. task assigned) within a TTL, it will be removed from the store.
 *  - TOMBSTONED: The environment has been marked as dead. This state occurs when the environment
 *    reaches an unrecoverable state, for example when the `podTemplateSpec` for workers is invalid
 *    and fails to successfully create any workers. It **cannot accept new task executions** and
 *    **can not scale up / down**. The environment will be removed from the store after a TTL.
 *  - INITIALIZING: The environment is being created. This state is used to handle multiple processes
 *    attempting to create the same environment at the same time and also prevent processes from assigning
 *    tasks to an environment that is not ready.
 *
 *  `Worker`
 *  - HEALTHY: The pod for this worker exists and the gRPC connection is active. This worker **can
 *    accept new task executions**.
 *  - ORPHANED: The pod for this worker exists and previously had an active gRPC connection that was lost.
 *    This worker **can not accept new task executions** but may reconnect within the grace period.
 *  - TOMBSTONED: Workers can not be tombstoned.
 *  - INITIALIZING: The pod for this worker is starting up (scheduling, pulling image, or launching).
 *    It has never established a gRPC connection and **can not accept new task executions**. Workers
 *    in this state are also created during orphan detection for pods that exist but haven't connected
 *    to the current deployment.
 */

type State int32

const (
	HEALTHY State = iota
	ORPHANED
	TOMBSTONED
	INITIALIZING
)

type Worker interface {
	Capacity() *pb.Capacity
	SetCapacity(*pb.Capacity)
	LastAccessedAt() int64
	SetLastAccessedAt(int64)
	State() State
	SetState(State)
	ID() string
	EnqueueHeartbeatResponse(*pb.HeartbeatResponse)
	Responses() <-chan *pb.HeartbeatResponse
}

type TaskStatus struct {
	Phase        core.Phase
	Reason       string
	TaskDuration time.Duration
}

type FastTaskService interface {
	AddPendingOwner(queueID, taskID string, enqueueLabels map[string]string) bool
	CheckStatus(ctx context.Context, taskID, queueID, workerID string) (TaskStatus, error)
	Cleanup(ctx context.Context, taskID, queueID, workerID string) error
	OfferTaskToEnvironment(ctx context.Context, execID *idlcore.WorkflowExecutionIdentifier, environmentID, taskID, namespace, workflowID string, cmd []string, envVars map[string]string, enqueueLabels map[string]string) (Worker, error)
}

type ExecutionEnvID struct {
	Org     string
	Project string
	Domain  string
	Name    string
	Version string
}

func (e ExecutionEnvID) String() string {
	if len(e.Org) == 0 {
		return e.Project + "_" + e.Domain + "_" + e.Name + "_" + e.Version
	}
	return e.Org + "_" + e.Project + "_" + e.Domain + "_" + e.Name + "_" + e.Version
}

type Environment interface {
	GetWorker(workerID string) Worker
	RangeWorkers(func(workerID string, worker Worker) bool)
	DeleteWorker(workerID string)
	GetOrCreateWorker(workerID string) Worker
	CreatedAt() int64
	EnvID() ExecutionEnvID
	FailureMessage() string
	SetFailureMessage(failureMessage string)
	LastScaledDownAt() int64
	SetLastScaledDownAt(lastScaledDownAt int64)
	State() State
	SetState(state State)
	Recover(envID ExecutionEnvID, fastTaskEnvironmentSpec *pb.FastTaskEnvironmentSpec)
	FastTaskEnvironmentSpec() *pb.FastTaskEnvironmentSpec
}

type EnvironmentBuilder interface {
	GetOrCreateEnvironment(ctx context.Context, tCtx core.TaskExecutionContext,
		envID ExecutionEnvID, executionEnv *idlcore.ExecutionEnv) (Environment, error)
	ScaleUp(ctx context.Context, executionEnvID string)

	GetWorkerPod(ctx context.Context, executionEnvID, workerID string) (*v1.Pod, error)
	ValidateWorkerPods(ctx context.Context, executionEnvID string, taskInfo *core.TaskInfo) (string, error)
}

type EnvironmentStore interface {
	Delete(executionEnvID string)
	Get(executionEnvID string) Environment
	List() []Environment
	GetOrCreate(executionEnvID string, env Environment) Environment
}
