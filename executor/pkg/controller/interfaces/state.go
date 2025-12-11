package interfaces

import (
	"github.com/flyteorg/flyte/v2/executor/pkg/controller/handler"
)

type NodeStateWriter interface {
	PutTaskNodeState(s handler.TaskNodeState) error
	ClearNodeStatus()
}

type NodeStateReader interface {
	HasTaskNodeState() bool
	GetTaskNodeState() handler.TaskNodeState
}
