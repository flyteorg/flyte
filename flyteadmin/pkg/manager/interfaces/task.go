package interfaces

import (
	"context"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
)

// Interface for managing Flyte Tasks
type TaskInterface interface {
	CreateTask(ctx context.Context, request admin.TaskCreateRequest) (*admin.TaskCreateResponse, error)
	GetTask(ctx context.Context, request admin.ObjectGetRequest) (*admin.Task, error)
	ListTasks(ctx context.Context, request admin.ResourceListRequest) (*admin.TaskList, error)
	ListUniqueTaskIdentifiers(ctx context.Context, request admin.NamedEntityIdentifierListRequest) (
		*admin.NamedEntityIdentifierList, error)
}
