package interfaces

import (
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
)

//go:generate mockery --name=NodeExecutionEventWriter --output=../mocks --case=underscore --with-expecter

type NodeExecutionEventWriter interface {
	Run()
	Write(nodeExecutionEvent *admin.NodeExecutionEventRequest)
}
