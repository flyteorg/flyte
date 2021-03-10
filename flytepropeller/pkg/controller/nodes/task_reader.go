package nodes

import (
	"context"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"

	"github.com/flyteorg/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
)

type taskReader struct {
	*core.TaskTemplate
}

func (t taskReader) GetTaskType() v1alpha1.TaskType {
	return t.TaskTemplate.Type
}

func (t taskReader) GetTaskID() *core.Identifier {
	return t.Id
}

func (t taskReader) Read(ctx context.Context) (*core.TaskTemplate, error) {
	return t.TaskTemplate, nil
}
