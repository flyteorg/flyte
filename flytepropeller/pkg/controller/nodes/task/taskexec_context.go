package task

import (
	"bytes"
	"context"
	"strconv"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/flyteorg/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"

	"github.com/flyteorg/flytepropeller/pkg/controller/nodes/common"

	"github.com/flyteorg/flytepropeller/pkg/controller/nodes/task/resourcemanager"

	"github.com/flyteorg/flytestdlib/logger"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"

	pluginCatalog "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/catalog"
	pluginCore "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/io"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/ioutils"

	"github.com/flyteorg/flytepropeller/pkg/controller/nodes/errors"
	"github.com/flyteorg/flytepropeller/pkg/controller/nodes/handler"
	"github.com/flyteorg/flytepropeller/pkg/utils"
)

var (
	_ pluginCore.TaskExecutionContext = &taskExecutionContext{}
)

const IDMaxLength = 50
const DefaultMaxAttempts = 1

type taskExecutionID struct {
	execName string
	id       *core.TaskExecutionIdentifier
}

func (te taskExecutionID) GetID() core.TaskExecutionIdentifier {
	return *te.id
}

func (te taskExecutionID) GetGeneratedName() string {
	return te.execName
}

type taskOverrides struct {
	pluginCore.TaskOverrides
	resourceRequirements *v1.ResourceRequirements
}

func (t taskOverrides) GetResources() *v1.ResourceRequirements {
	return t.resourceRequirements
}

func newTaskOverrides(overrides pluginCore.TaskOverrides, resourceRequirements *v1.ResourceRequirements) pluginCore.TaskOverrides {
	return &taskOverrides{
		TaskOverrides:        overrides,
		resourceRequirements: resourceRequirements,
	}
}

type taskExecutionMetadata struct {
	handler.NodeExecutionMetadata
	taskExecID  taskExecutionID
	o           pluginCore.TaskOverrides
	maxAttempts uint32
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

type taskExecutionContext struct {
	handler.NodeExecutionContext
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

// Reconciles platform-specific resource defaults requests and max limits with the static resource values
// defined by this task and node execution context.
func determineResourceRequirements(nCtx handler.NodeExecutionContext, taskResources v1alpha1.TaskResources) *v1.ResourceRequirements {
	var requests = make(v1.ResourceList)
	var limits = make(v1.ResourceList)
	if nCtx.Node().GetResources() != nil {
		if nCtx.Node().GetResources().Requests != nil {
			requests = nCtx.Node().GetResources().Requests
		}
		if nCtx.Node().GetResources().Limits != nil {
			limits = nCtx.Node().GetResources().Limits
		}
	}

	assignResource(v1.ResourceCPU, taskResources.Requests.CPU, taskResources.Limits.CPU, requests, limits)
	assignResource(v1.ResourceMemory, taskResources.Requests.Memory, taskResources.Limits.Memory, requests, limits)
	assignResource(v1.ResourceEphemeralStorage, taskResources.Requests.EphemeralStorage, taskResources.Limits.EphemeralStorage, requests, limits)
	assignResource(v1.ResourceStorage, taskResources.Requests.Storage, taskResources.Limits.Storage, requests, limits)
	return &v1.ResourceRequirements{
		Requests: requests,
		Limits:   limits,
	}
}

func (t *Handler) newTaskExecutionContext(ctx context.Context, nCtx handler.NodeExecutionContext, plugin pluginCore.Plugin) (*taskExecutionContext, error) {
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

	uniqueID, err := utils.FixedLengthUniqueIDForParts(length, nCtx.NodeExecutionMetadata().GetOwnerID().Name, currentNodeUniqueID, strconv.Itoa(int(id.RetryAttempt)))
	if err != nil {
		// SHOULD never really happen
		return nil, err
	}

	outputSandbox, err := ioutils.NewShardedRawOutputPath(ctx, nCtx.OutputShardSelector(), nCtx.RawOutputPrefix(), uniqueID, nCtx.DataStore())
	if err != nil {
		return nil, errors.Wrapf(errors.StorageError, nCtx.NodeID(), err, "failed to create output sandbox for node execution")
	}
	ow := ioutils.NewBufferedOutputWriter(ctx, ioutils.NewRemoteFileOutputPaths(ctx, nCtx.DataStore(), nCtx.NodeStatus().GetOutputDir(), outputSandbox))
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
	maxAttempts := uint32(DefaultMaxAttempts)
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
			taskExecID:            taskExecutionID{execName: uniqueID, id: id},
			o:                     newTaskOverrides(nCtx.Node(), determineResourceRequirements(nCtx, nCtx.ExecutionContext().GetExecutionConfig().TaskResources)),
			maxAttempts:           maxAttempts,
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
