package manager

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	durationpb "google.golang.org/protobuf/types/known/durationpb"

	cacheconfig "github.com/flyteorg/flyte/v2/cache_service/config"
	"github.com/flyteorg/flyte/v2/cache_service/repository"
	cacheservicepb "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/cacheservice"
)

func newTestManager(t *testing.T) *Manager {
	t.Helper()
	t.Cleanup(func() {
		testDB.Exec("DELETE FROM cached_outputs")
		testDB.Exec("DELETE FROM reservations")
	})

	cfg := &cacheconfig.Config{
		HeartbeatGracePeriodMultiplier: 3,
	}
	cfg.MaxReservationHeartbeat.Duration = 10 * time.Second

	repos := repository.NewRepository(testDB)
	return New(cfg, repos.CachedOutputRepo(), repos.ReservationRepo())
}

func TestManagerPutGetURIOutput(t *testing.T) {
	m := newTestManager(t)
	ctx := context.Background()

	err := m.Put(ctx, &cacheservicepb.PutCacheRequest{
		Key: "k1",
		Output: &cacheservicepb.CachedOutput{
			Output: &cacheservicepb.CachedOutput_OutputUri{OutputUri: "s3://bucket/path"},
		},
	})
	require.NoError(t, err)

	entry, err := m.Get(ctx, &cacheservicepb.GetCacheRequest{Key: "k1"})
	require.NoError(t, err)
	require.Equal(t, "s3://bucket/path", entry.OutputURI)
}

func TestManagerReservationReturnsExistingOwnerWhenHeld(t *testing.T) {
	m := newTestManager(t)
	ctx := context.Background()
	now := time.Now().UTC()

	first, err := m.GetOrExtendReservation(ctx, &cacheservicepb.GetOrExtendReservationRequest{
		Key:               "k1",
		OwnerId:           "owner-a",
		HeartbeatInterval: durationpb.New(2 * time.Second),
	}, now)
	require.NoError(t, err)
	require.Equal(t, "owner-a", first.GetOwnerId())

	second, err := m.GetOrExtendReservation(ctx, &cacheservicepb.GetOrExtendReservationRequest{
		Key:               "k1",
		OwnerId:           "owner-b",
		HeartbeatInterval: durationpb.New(2 * time.Second),
	}, now.Add(1*time.Second))
	require.NoError(t, err)
	require.Equal(t, "owner-a", second.GetOwnerId())
}
