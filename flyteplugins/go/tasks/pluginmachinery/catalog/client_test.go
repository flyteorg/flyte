package catalog

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/io/mocks"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
)

var (
	cacheStatus     = core.CatalogCacheStatus_CACHE_MISS
	catalogMetadata = core.CatalogMetadata{
		DatasetId: &core.Identifier{
			Project: "project",
			Domain:  "domain",
			Name:    "name",
		},
		ArtifactTag: &core.CatalogArtifactTag{
			ArtifactId: "artifactID",
			Name:       "artifactName",
		},
	}
	key = &Key{
		Identifier: core.Identifier{
			Project: "project",
			Domain:  "domain",
			Name:    "name",
			Version: "1.0.0",
		},
		CacheVersion: "1.0.0",
		TypedInterface: core.TypedInterface{
			Inputs:  nil,
			Outputs: nil,
		},
	}
)

func TestNewPutFailureStatus(t *testing.T) {
	status := NewPutFailureStatus(key)

	assert.Equal(t, status.GetCacheStatus(), core.CatalogCacheStatus_CACHE_PUT_FAILURE)
	assert.EqualValues(t, status.GetMetadata().GetDatasetId(), &key.Identifier)
}

func TestStatus(t *testing.T) {
	status := NewStatus(cacheStatus, &catalogMetadata)

	assert.Equal(t, status.GetCacheStatus(), cacheStatus)
	assert.Equal(t, status.GetMetadata().GetDatasetId().GetProject(), catalogMetadata.GetDatasetId().GetProject())
	assert.Equal(t, status.GetMetadata().GetDatasetId().GetDomain(), catalogMetadata.GetDatasetId().GetDomain())
	assert.Equal(t, status.GetMetadata().GetDatasetId().GetName(), catalogMetadata.GetDatasetId().GetName())
	assert.Equal(t, status.GetMetadata().GetArtifactTag().GetArtifactId(), catalogMetadata.GetArtifactTag().GetArtifactId())
	assert.Equal(t, status.GetMetadata().GetArtifactTag().GetName(), catalogMetadata.GetArtifactTag().GetName())
}

func TestEntry(t *testing.T) {
	tests := []struct {
		name  string
		entry Entry
	}{
		{
			"base",
			NewCatalogEntry(&mocks.OutputReader{}, NewStatus(cacheStatus, &catalogMetadata)),
		},
		{
			"failed",
			NewFailedCatalogEntry(NewStatus(cacheStatus, &catalogMetadata)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status := tt.entry.GetStatus()
			assert.Equal(t, status.GetCacheStatus(), cacheStatus)
			assert.Equal(t, status.GetMetadata().GetDatasetId().GetProject(), catalogMetadata.GetDatasetId().GetProject())
			assert.Equal(t, status.GetMetadata().GetDatasetId().GetDomain(), catalogMetadata.GetDatasetId().GetDomain())
			assert.Equal(t, status.GetMetadata().GetDatasetId().GetName(), catalogMetadata.GetDatasetId().GetName())
			assert.Equal(t, status.GetMetadata().GetArtifactTag().GetArtifactId(), catalogMetadata.GetArtifactTag().GetArtifactId())
			assert.Equal(t, status.GetMetadata().GetArtifactTag().GetName(), catalogMetadata.GetArtifactTag().GetName())
		})
	}
}

func TestReservationEntry(t *testing.T) {
	reservationStatus := core.CatalogReservation_RESERVATION_ACQUIRED
	tests := []struct {
		name              string
		reservationEntry  ReservationEntry
		expiresAt         time.Time
		heartbeatInterval time.Duration
		ownerID           string
		status            core.CatalogReservation_Status
	}{
		{
			"base",
			NewReservationEntry(time.Time{}, 5*time.Second, "owner", reservationStatus),
			time.Time{},
			5 * time.Second,
			"owner",
			reservationStatus,
		},
		{
			"status",
			NewReservationEntryStatus(reservationStatus),
			time.Time{},
			0 * time.Second,
			"",
			reservationStatus,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.reservationEntry.GetExpiresAt(), tt.expiresAt)
			assert.Equal(t, tt.reservationEntry.GetHeartbeatInterval(), tt.heartbeatInterval)
			assert.Equal(t, tt.reservationEntry.GetOwnerID(), tt.ownerID)
			assert.Equal(t, tt.reservationEntry.GetStatus(), tt.status)
		})
	}
}
