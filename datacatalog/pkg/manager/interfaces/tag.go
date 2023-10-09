package interfaces

import (
	"context"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/datacatalog"
)

type TagManager interface {
	AddTag(ctx context.Context, request *datacatalog.AddTagRequest) (*datacatalog.AddTagResponse, error)
}
