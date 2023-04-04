package example

import (
	"context"
	"time"

	idlCore "github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"

	"github.com/flyteorg/flytestdlib/errors"
	"github.com/flyteorg/flytestdlib/promutils"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/webapi"
)

const (
	ErrRemoteSystem errors.ErrorCode = "RemoteSystem"
	ErrRemoteUser   errors.ErrorCode = "RemoteUser"
	ErrSystem       errors.ErrorCode = "System"
)

type Plugin struct {
	metricScope promutils.Scope
	cfg         *Config
}

func (p Plugin) GetConfig() webapi.PluginConfig {
	return GetConfig().WebAPI
}

func (p Plugin) ResourceRequirements(_ context.Context, _ webapi.TaskExecutionContextReader) (
	namespace core.ResourceNamespace, constraints core.ResourceConstraintsSpec, err error) {

	// Resource requirements are assumed to be the same.
	return "default", p.cfg.ResourceConstraints, nil
}

func (p Plugin) Create(ctx context.Context, tCtx webapi.TaskExecutionContextReader) (resourceMeta webapi.ResourceMeta,
	resource webapi.Resource, err error) {

	// In the create method, your code should understand the request and translate the flyte task template data into a
	// format your webAPI expects then make the WebAPI call.

	// This snippet retrieves the TaskTemplate and unmarshals the custom field into a plugin-protobuf.
	//task, err := tCtx.TaskReader().Read(ctx)
	//if err != nil {
	//	return nil, nil, err
	//}

	//custom := task.GetCustom()
	//myPluginProtoStruct := &plugins.MyPluginProtoStruct{}
	//err = utils.UnmarshalStructToPb(custom, myPluginProtoStruct)
	//if err != nil {
	//	return nil, nil, err
	//}

	// The system will invoke this API at least once. In cases when a network partition/failure causes the system to
	// fail persisting this response, the system will call this API again. If it returns an error, it'll be called up to
	// the system-wide defined limit of retries with exponential backoff and jitter in between these trials

	return "my-request-id", nil, nil
}

// Get the resource that matches the keys. If the plugin hits any failure, it should stop and return the failure.
// This API will be called asynchronously and periodically to update the set of tasks currently in progress. It's
// acceptable if this API is blocking since it'll be called from a background go-routine.
// Best practices:
//  1. Instead of returning the entire response object retrieved from the WebAPI, construct a smaller object that
//     has enough information to construct the status/phase, error and/or output.
//  2. This object will NOT be serialized/marshaled. It's, therefore, not a requirement to make it so.
//  3. There is already client-side throttling in place. If the WebAPI returns a throttling error, you should return
//     it as is so that the appropriate metrics are updated and the system administrator can update throttling
//     params accordingly.
func (p Plugin) Get(ctx context.Context, tCtx webapi.GetContext) (latest webapi.Resource, err error) {
	return "my-resource", nil
}

// Delete the object in the remote service using the resource key. Flyte will call this API at least once. If the
// resource has already been deleted, the API should not fail.
func (p Plugin) Delete(ctx context.Context, tCtx webapi.DeleteContext) error {
	return nil
}

// Status checks the status of a given resource and translates it to a Flyte-understandable PhaseInfo. This API
// should avoid making any network calls and should run very efficiently.
func (p Plugin) Status(ctx context.Context, tCtx webapi.StatusContext) (phase core.PhaseInfo, err error) {
	tNow := time.Now()
	return core.PhaseInfoSuccess(&core.TaskInfo{
		Logs: []*idlCore.TaskLog{
			{
				Uri:  "https://my-service/abc",
				Name: "ServiceA Console",
			},
		},
		OccurredAt: &tNow,
		ExternalResources: []*core.ExternalResource{
			{
				ExternalID: "abc",
			},
		},
	}), nil
}

func NewPlugin(ctx context.Context, cfg *Config, metricScope promutils.Scope) (Plugin, error) {
	return Plugin{
		metricScope: metricScope,
		cfg:         cfg,
	}, nil
}

func init() {
	pluginmachinery.PluginRegistry().RegisterRemotePlugin(webapi.PluginEntry{
		ID:                 "service-a",
		SupportedTaskTypes: []core.TaskType{"my-task"},
		PluginLoader: func(ctx context.Context, iCtx webapi.PluginSetupContext) (webapi.AsyncPlugin, error) {
			return NewPlugin(ctx, GetConfig(), iCtx.MetricsScope())
		},
		IsDefault:           false,
		DefaultForTaskTypes: []core.TaskType{"my-task"},
	})
}
