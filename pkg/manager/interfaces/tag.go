package interfaces

import (
	"context"

	datacatalog "github.com/flyteorg/datacatalog/protos/gen"
)

type TagManager interface {
	AddTag(ctx context.Context, request *datacatalog.AddTagRequest) (*datacatalog.AddTagResponse, error)
}
