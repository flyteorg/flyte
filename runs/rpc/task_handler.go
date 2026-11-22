package rpc

import (
	"context"

	"connectrpc.com/connect"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/task"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/task/taskconnect"
	"github.com/flyteorg/flyte/v2/runs/service"
)

type taskHandler struct{ tasks service.Task }

// NewTaskHandler exposes the task service through the Connect API.
func NewTaskHandler(tasks service.Task) taskconnect.TaskServiceHandler {
	return &taskHandler{tasks: tasks}
}

func (h *taskHandler) DeployTask(ctx context.Context, req *connect.Request[task.DeployTaskRequest]) (*connect.Response[task.DeployTaskResponse], error) {
	return rpcResponse(h.tasks.DeployTask(ctx, req.Msg))
}

func (h *taskHandler) GetTaskDetails(ctx context.Context, req *connect.Request[task.GetTaskDetailsRequest]) (*connect.Response[task.GetTaskDetailsResponse], error) {
	return rpcResponse(h.tasks.GetTaskDetails(ctx, req.Msg))
}

func (h *taskHandler) ListTasks(ctx context.Context, req *connect.Request[task.ListTasksRequest]) (*connect.Response[task.ListTasksResponse], error) {
	return rpcResponse(h.tasks.ListTasks(ctx, req.Msg))
}

func (h *taskHandler) ListVersions(ctx context.Context, req *connect.Request[task.ListVersionsRequest]) (*connect.Response[task.ListVersionsResponse], error) {
	return rpcResponse(h.tasks.ListVersions(ctx, req.Msg))
}
