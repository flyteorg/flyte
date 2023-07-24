package interfaces

import (
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
)

//go:generate mockery -name=NodeExecutionEventWriter -output=../mocks -case=underscore

type NodeExecutionEventWriter interface {
	Run()
	Write(nodeExecutionEvent admin.NodeExecutionEventRequest)
}
