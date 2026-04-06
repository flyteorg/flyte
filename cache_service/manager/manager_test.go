package manager

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/stretchr/testify/require"
	durationpb "google.golang.org/protobuf/types/known/durationpb"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	cacheconfig "github.com/flyteorg/flyte/v2/cache_service/config"
	"github.com/flyteorg/flyte/v2/cache_service/migrations"
	"github.com/flyteorg/flyte/v2/cache_service/repository"
	cacheservicepb "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/cacheservice"
)

func newTestManager(t *testing.T) *Manager {
	t.Helper()

	host := os.Getenv("TEST_POSTGRES_HOST")
	if host == "" {
		host = "localhost"
	}
	port := os.Getenv("TEST_POSTGRES_PORT")
	if port == "" {
		port = "5433"
	}
	user := os.Getenv("TEST_POSTGRES_USER")
	if user == "" {
		user = "postgres"
	}
	password := os.Getenv("TEST_POSTGRES_PASSWORD")
	if password == "" {
		password = "postgres"
	}
	dbname := os.Getenv("TEST_POSTGRES_DB")
	if dbname == "" {
		dbname = "flyte_runs"
	}
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	m := gormigrate.New(db, gormigrate.DefaultOptions, migrations.CacheServiceMigrations)
	require.NoError(t, m.Migrate())

	cfg := &cacheconfig.Config{
		HeartbeatGracePeriodMultiplier: 3,
	}
	cfg.MaxReservationHeartbeat.Duration = 10 * time.Second

	repos := repository.NewRepository(db)
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
