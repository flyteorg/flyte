package task

import (
	"bytes"
	"context"
	"strconv"
	"strings"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	pluginCatalog "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/catalog"
	pluginCore "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/encoding"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/io"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/ioutils"
	"github.com/flyteorg/flyte/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	controllerconfig "github.com/flyteorg/flyte/flytepropeller/pkg/controller/config"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/common"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/errors"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/interfaces"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/task/resourcemanager"
	"github.com/flyteorg/flyte/flytepropeller/pkg/utils"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/storage"
)

var (
	_ pluginCore.TaskExecutionContext = &taskExecutionContext{}
)

const IDMaxLength = 50

type taskExecutionID struct {
	execName     string
	id           *core.TaskExecutionIdentifier
	uniqueNodeID string
}

func (te taskExecutionID) GetID() core.TaskExecutionIdentifier {
	return *te.id
}

func (te taskExecutionID) GetUniqueNodeID() string {
	return te.uniqueNodeID
}

func (te taskExecutionID) GetGeneratedName() string {
	return te.execName
}

func (te taskExecutionID) GetGeneratedNameWith(minLength, maxLength int) (string, error) {
	origLength := len(te.execName)
	if origLength < minLength {
		return te.execName + strings.Repeat("0", minLength-origLength), nil
	}

	if origLength > maxLength {
		return encoding.FixedLengthUniqueID(te.execName, maxLength)
	}

	return te.execName, nil
}

type taskExecutionMetadata struct {
	interfaces.NodeExecutionMetadata
	taskExecID           taskExecutionID
	o                    pluginCore.TaskOverrides
	maxAttempts          uint32
	platformResources    *v1.ResourceRequirements
	environmentVariables map[string]string
}

func (t taskExecutionMetadata) GetTaskExecutionID() pluginCore.TaskExecutionID {
	return t.taskExecID
}

func (t taskExecutionMetadata) GetOverrides() pluginCore.TaskOverrides {
	return t.o
}

func (t taskExecutionMetadata) GetMaxAttempts() uint32 {
	return t.maxAttempts
}

func (t taskExecutionMetadata) GetPlatformResources() *v1.ResourceRequirements {
	return t.platformResources
}

func (t taskExecutionMetadata) GetEnvironmentVariables() map[string]string {
	return t.environmentVariables
}

type taskExecutionContext struct {
	interfaces.NodeExecutionContext
	tm  taskExecutionMetadata
	rm  resourcemanager.TaskResourceManager
	psm *pluginStateManager
	tr  pluginCore.TaskReader
	ow  *ioutils.BufferedOutputWriter
	ber *bufferedEventRecorder
	sm  pluginCore.SecretManager
	c   pluginCatalog.AsyncClient
}

func (t *taskExecutionContext) TaskRefreshIndicator() pluginCore.SignalAsync {
	return func(ctx context.Context) {
		err := t.NodeExecutionContext.EnqueueOwnerFunc()
		if err != nil {
			logger.Errorf(ctx, "Failed to enqueue owner for Task [%v] and Owner [%v]. Error: %v",
				t.TaskExecutionMetadata().GetTaskExecutionID(),
				t.TaskExecutionMetadata().GetOwnerID(),
				err)
		}
	}
}

func (t *taskExecutionContext) Catalog() pluginCatalog.AsyncClient {
	return t.c
}

func (t taskExecutionContext) EventsRecorder() pluginCore.EventsRecorder {
	return t.ber
}

func (t taskExecutionContext) ResourceManager() pluginCore.ResourceManager {
	return t.rm
}

func (t taskExecutionContext) PluginStateReader() pluginCore.PluginStateReader {
	return t.psm
}

func (t *taskExecutionContext) TaskReader() pluginCore.TaskReader {
	return t.tr
}

func (t *taskExecutionContext) TaskExecutionMetadata() pluginCore.TaskExecutionMetadata {
	return t.tm
}

func (t *taskExecutionContext) OutputWriter() io.OutputWriter {
	return t.ow
}

func (t *taskExecutionContext) PluginStateWriter() pluginCore.PluginStateWriter {
	return t.psm
}

func (t taskExecutionContext) SecretManager() pluginCore.SecretManager {
	return t.sm
}

// Validates and assigns a single resource by examining the default requests and max limit with the static resource value
// defined by this task and node execution context.
func assignResource(resourceName v1.ResourceName, execConfigRequest, execConfigLimit resource.Quantity, requests, limits v1.ResourceList) {
	maxLimit := execConfigLimit
	request, ok := requests[resourceName]
	if !ok {
		// Requests aren't required so we glean it from the execution config value (when possible)
		if !execConfigRequest.IsZero() {
			request = execConfigRequest
		}
	} else {
		if request.Cmp(maxLimit) == 1 && !maxLimit.IsZero() {
			// Adjust the request downwards to not exceed the max limit if it's set.
			request = maxLimit
		}
	}

	limit, ok := limits[resourceName]
	if !ok {
		limit = request
	} else {
		if limit.Cmp(maxLimit) == 1 && !maxLimit.IsZero() {
			// Adjust the limit downwards to not exceed the max limit if it's set.
			limit = maxLimit
		}
	}
	if request.Cmp(limit) == 1 {
		// The limit should always be greater than or equal to the request
		request = limit
	}

	if !request.IsZero() {
		requests[resourceName] = request
	}
	if !limit.IsZero() {
		limits[resourceName] = limit
	}
}

func convertTaskResourcesToRequirements(taskResources v1alpha1.TaskResources) *v1.ResourceRequirements {
	return &v1.ResourceRequirements{
		Requests: v1.ResourceList{
			v1.ResourceCPU:              taskResources.Requests.CPU,
			v1.ResourceMemory:           taskResources.Requests.Memory,
			v1.ResourceEphemeralStorage: taskResources.Requests.EphemeralStorage,
			utils.ResourceNvidiaGPU:     taskResources.Requests.GPU,
		},
		Limits: v1.ResourceList{
			v1.ResourceCPU:              taskResources.Limits.CPU,
			v1.ResourceMemory:           taskResources.Limits.Memory,
			v1.ResourceEphemeralStorage: taskResources.Limits.EphemeralStorage,
			utils.ResourceNvidiaGPU:     taskResources.Limits.GPU,
		},
	}

}

// ComputeRawOutputPrefix constructs the output directory, where raw outputs of a task can be stored by the task. FlytePropeller may not have
// access to this location and can be passed in per execution.
// the function also returns the uniqueID generated
func ComputeRawOutputPrefix(ctx context.Context, length int, nCtx interfaces.NodeExecutionContext, currentNodeUniqueID v1alpha1.NodeID, currentAttempt uint32) (io.RawOutputPaths, string, error) {
	uniqueID, err := encoding.FixedLengthUniqueIDForParts(length, []string{nCtx.NodeExecutionMetadata().GetOwnerID().Name, currentNodeUniqueID, strconv.Itoa(int(currentAttempt))})
	if err != nil {
		// SHOULD never really happen
		return nil, uniqueID, err
	}

	rawOutputPrefix, err := ioutils.NewShardedRawOutputPath(ctx, nCtx.OutputShardSelector(), nCtx.RawOutputPrefix(), uniqueID, nCtx.DataStore())
	if err != nil {
		return nil, uniqueID, errors.Wrapf(errors.StorageError, nCtx.NodeID(), err, "failed to create output sandbox for node execution")
	}
	return rawOutputPrefix, uniqueID, nil
}

// ComputePreviousCheckpointPath returns the checkpoint path for the previous attempt, if this is the first attempt then returns an empty path
func ComputePreviousCheckpointPath(ctx context.Context, length int, nCtx interfaces.NodeExecutionContext, currentNodeUniqueID v1alpha1.NodeID, currentAttempt uint32) (storage.DataReference, error) {
	// If first attempt for this node execution, look for a checkpoint path in a prior execution
	if currentAttempt == 0 {
		return nCtx.NodeStateReader().GetTaskNodeState().PreviousNodeExecutionCheckpointURI, nil
	}
	// Otherwise derive previous checkpoint path from the prior attempt
	prevAttempt := currentAttempt - 1
	prevRawOutputPrefix, _, err := ComputeRawOutputPrefix(ctx, length, nCtx, currentNodeUniqueID, prevAttempt)
	if err != nil {
		return "", err
	}
	return ioutils.ConstructCheckpointPath(nCtx.DataStore(), prevRawOutputPrefix.GetRawOutputPrefix()), nil
}

func (t *Handler) newTaskExecutionContext(ctx context.Context, nCtx interfaces.NodeExecutionContext, plugin pluginCore.Plugin) (*taskExecutionContext, error) {
	id := GetTaskExecutionIdentifier(nCtx)

	currentNodeUniqueID := nCtx.NodeID()
	if nCtx.ExecutionContext().GetEventVersion() != v1alpha1.EventVersion0 {
		var err error
		currentNodeUniqueID, err = common.GenerateUniqueID(nCtx.ExecutionContext().GetParentInfo(), nCtx.NodeID())
		if err != nil {
			return nil, err
		}
	}

	length := IDMaxLength
	if l := plugin.GetProperties().GeneratedNameMaxLength; l != nil {
		length = *l
	}

	rawOutputPrefix, uniqueID, err := ComputeRawOutputPrefix(ctx, length, nCtx, currentNodeUniqueID, id.RetryAttempt)
	if err != nil {
		return nil, err
	}

	prevCheckpointPath, err := ComputePreviousCheckpointPath(ctx, length, nCtx, currentNodeUniqueID, id.RetryAttempt)
	if err != nil {
		return nil, err
	}

	ow := ioutils.NewBufferedOutputWriter(ctx, ioutils.NewCheckpointRemoteFilePaths(ctx, nCtx.DataStore(), nCtx.NodeStatus().GetOutputDir(), rawOutputPrefix, prevCheckpointPath))
	ts := nCtx.NodeStateReader().GetTaskNodeState()
	var b *bytes.Buffer
	if ts.PluginState != nil {
		b = bytes.NewBuffer(ts.PluginState)
	}
	psm, err := newPluginStateManager(ctx, GobCodecVersion, ts.PluginStateVersion, b)
	if err != nil {
		return nil, errors.Wrapf(errors.RuntimeExecutionError, nCtx.NodeID(), err, "unable to initialize plugin state manager")
	}

	resourceNamespacePrefix := pluginCore.ResourceNamespace(t.resourceManager.GetID()).CreateSubNamespace(pluginCore.ResourceNamespace(plugin.GetID()))
	maxAttempts := uint32(controllerconfig.GetConfig().NodeConfig.DefaultMaxAttempts)
	if nCtx.Node().GetRetryStrategy() != nil && nCtx.Node().GetRetryStrategy().MinAttempts != nil {
		maxAttempts = uint32(*nCtx.Node().GetRetryStrategy().MinAttempts)
	}

	taskTemplatePath, err := ioutils.GetTaskTemplatePath(ctx, nCtx.DataStore(), nCtx.NodeStatus().GetDataDir())
	if err != nil {
		return nil, err
	}

	return &taskExecutionContext{
		NodeExecutionContext: nCtx,
		tm: taskExecutionMetadata{
			NodeExecutionMetadata: nCtx.NodeExecutionMetadata(),
			taskExecID: taskExecutionID{
				execName:     uniqueID,
				id:           id,
				uniqueNodeID: currentNodeUniqueID,
			},
			o:                    nCtx.Node(),
			maxAttempts:          maxAttempts,
			platformResources:    convertTaskResourcesToRequirements(nCtx.ExecutionContext().GetExecutionConfig().TaskResources),
			environmentVariables: nCtx.ExecutionContext().GetExecutionConfig().EnvironmentVariables,
		},
		rm: resourcemanager.GetTaskResourceManager(
			t.resourceManager, resourceNamespacePrefix, id),
		psm: psm,
		tr:  ioutils.NewLazyUploadingTaskReader(nCtx.TaskReader(), taskTemplatePath, nCtx.DataStore()),
		ow:  ow,
		ber: newBufferedEventRecorder(),
		c:   t.asyncCatalog,
		sm:  t.secretManager,
	}, nil
}
