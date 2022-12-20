package interfaces

import (
	"context"

	idl_datacatalog "github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/datacatalog"
)

type DatasetManager interface {
	CreateDataset(ctx context.Context, request *idl_datacatalog.CreateDatasetRequest) (*idl_datacatalog.CreateDatasetResponse, error)
	GetDataset(ctx context.Context, request *idl_datacatalog.GetDatasetRequest) (*idl_datacatalog.GetDatasetResponse, error)
	ListDatasets(ctx context.Context, request *idl_datacatalog.ListDatasetsRequest) (*idl_datacatalog.ListDatasetsResponse, error)
}
