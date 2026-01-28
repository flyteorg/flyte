package interfaces

import (
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
)


type NodeExecutionEventWriter interface {
	Run()
	Write(nodeExecutionEvent *admin.NodeExecutionEventRequest)
}
