package interfaces

import (
	"context"

	datacatalog "github.com/lyft/datacatalog/protos/gen"
)

type TagManager interface {
	AddTag(ctx context.Context, request datacatalog.AddTagRequest) (*datacatalog.AddTagResponse, error)
}
