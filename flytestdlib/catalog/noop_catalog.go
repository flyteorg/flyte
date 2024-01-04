package catalog

import (
	"context"
	"time"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/datacatalog"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/catalog"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/io"
)

var (
	_ catalog.Client = &NOOPCatalog{}
)

var disabledStatus = catalog.NewStatus(core.CatalogCacheStatus_CACHE_DISABLED, nil)

type NOOPCatalog struct {
}

func (n NOOPCatalog) Get(_ context.Context, _ catalog.Key) (catalog.Entry, error) {
	return catalog.NewCatalogEntry(nil, disabledStatus), nil
}

func (n NOOPCatalog) Put(_ context.Context, _ catalog.Key, _ io.OutputReader, _ catalog.Metadata) (catalog.Status, error) {
	return disabledStatus, nil
}

func (n NOOPCatalog) Update(_ context.Context, _ catalog.Key, _ io.OutputReader, _ catalog.Metadata) (catalog.Status, error) {
	return disabledStatus, nil
}

func (n NOOPCatalog) GetOrExtendReservation(_ context.Context, _ catalog.Key, _ string, _ time.Duration) (*datacatalog.Reservation, error) {
	return nil, nil
}

func (n NOOPCatalog) GetOrExtendReservationByArtifactTag(_ context.Context, _ *datacatalog.DatasetID, _ string, _ string, _ time.Duration) (*datacatalog.Reservation, error) {
	return nil, nil
}

func (n NOOPCatalog) ReleaseReservation(_ context.Context, _ catalog.Key, _ string) error {
	return nil
}

func (n NOOPCatalog) ReleaseReservationByArtifactTag(_ context.Context, _ *datacatalog.DatasetID, _ string, _ string) error {
	return nil
}

func (n NOOPCatalog) Delete(_ context.Context, _ catalog.Key) error {
	return nil
}

func (n NOOPCatalog) DeleteByArtifactTag(_ context.Context, _ *datacatalog.DatasetID, _ string) error {
	return nil
}

func (n NOOPCatalog) DeleteByArtifactID(_ context.Context, _ *datacatalog.DatasetID, _ string) error {
	return nil
}
