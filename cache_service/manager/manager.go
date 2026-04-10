package manager

import (
	"context"
	"fmt"
	"time"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/proto"
	durationpb "google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	cacheconfig "github.com/flyteorg/flyte/v2/cache_service/config"
	repositoryerrors "github.com/flyteorg/flyte/v2/cache_service/repository/errors"
	"github.com/flyteorg/flyte/v2/cache_service/repository/interfaces"
	"github.com/flyteorg/flyte/v2/cache_service/repository/models"
	cacheservicepb "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/cacheservice"
)

const reservationPrefix = "reservation"

// Manager owns the cache service's core behavior:
//   - read/write cached outputs
//   - maintain cache metadata timestamps
//   - coordinate serialized cache population with reservations
//
// The service layer stays thin and only translates V2 transport requests into
// manager calls. Repository implementations only persist rows. The policy of
// "who gets to populate cache and when" lives here.
type Manager struct {
	outputs                         interfaces.CachedOutputRepo
	reservations                    interfaces.ReservationRepo
	heartbeatGracePeriodMultiplier  int
	maxReservationHeartbeatInterval time.Duration
}

type CacheEntry struct {
	OutputURI string
	Metadata  *cacheservicepb.Metadata
}

func New(cfg *cacheconfig.Config, outputs interfaces.CachedOutputRepo, reservations interfaces.ReservationRepo) *Manager {
	maxHeartbeat := cfg.MaxReservationHeartbeat.Duration
	if maxHeartbeat <= 0 {
		maxHeartbeat = 10 * time.Second
	}

	graceMultiplier := cfg.HeartbeatGracePeriodMultiplier
	if graceMultiplier <= 0 {
		graceMultiplier = 3
	}

	return &Manager{
		outputs:                         outputs,
		reservations:                    reservations,
		heartbeatGracePeriodMultiplier:  graceMultiplier,
		maxReservationHeartbeatInterval: maxHeartbeat,
	}
}

// Get returns the materialized cache entry for the given key, if one exists.
func (m *Manager) Get(ctx context.Context, request *cacheservicepb.GetCacheRequest) (*CacheEntry, error) {
	if request.GetKey() == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("key is required"))
	}

	output, err := m.outputs.Get(ctx, request.GetKey())
	if err != nil {
		if repositoryerrors.IsNotFound(err) {
			return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("cache entry %q not found", request.GetKey()))
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	metadata, err := unmarshalMetadata(output.Metadata)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return &CacheEntry{
		OutputURI: output.OutputURI,
		Metadata:  metadata,
	}, nil
}

// Put stores or overwrites the materialized cache entry for a key.
//
// In OSS V2 we only persist output URIs. The actual output payload continues to
// live in object storage; cache service stores the lookup record plus metadata.
func (m *Manager) Put(ctx context.Context, request *cacheservicepb.PutCacheRequest) error {
	if request.GetKey() == "" {
		return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("key is required"))
	}
	if request.GetOutput() == nil || request.GetOutput().GetOutputUri() == "" {
		return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("output_uri is required"))
	}

	now := time.Now().UTC()
	existing, err := m.outputs.Get(ctx, request.GetKey())
	if err != nil && !repositoryerrors.IsNotFound(err) {
		return connect.NewError(connect.CodeInternal, err)
	}

	if err == nil {
		expired := false
		if maxAge := request.GetOverwrite().GetMaxAge(); maxAge != nil && !existing.LastUpdated.IsZero() {
			expired = time.Since(existing.LastUpdated) > maxAge.AsDuration()
		}
		if !request.GetOverwrite().GetOverwrite() && !expired {
			return connect.NewError(connect.CodeAlreadyExists, fmt.Errorf("cache entry %q already exists", request.GetKey()))
		}
	}

	metadata := mergeMetadata(existing, request.GetOutput().GetMetadata(), now)
	metadataBytes, err := proto.Marshal(metadata)
	if err != nil {
		return connect.NewError(connect.CodeInternal, err)
	}

	model := &models.CachedOutput{
		Key:         request.GetKey(),
		OutputURI:   request.GetOutput().GetOutputUri(),
		Metadata:    metadataBytes,
		LastUpdated: metadata.GetLastUpdatedAt().AsTime(),
	}
	if err := m.outputs.Put(ctx, model); err != nil {
		return connect.NewError(connect.CodeInternal, err)
	}

	return nil
}

// Delete removes the cache metadata row for a key.
//
// This does not delete the referenced object-storage blob. The caller may be
// pointing at a shared URI, and blob lifecycle is better handled separately.
func (m *Manager) Delete(ctx context.Context, request *cacheservicepb.DeleteCacheRequest) error {
	if request.GetKey() == "" {
		return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("key is required"))
	}
	if err := m.outputs.Delete(ctx, request.GetKey()); err != nil {
		if repositoryerrors.IsNotFound(err) {
			return connect.NewError(connect.CodeNotFound, fmt.Errorf("cache entry %q not found", request.GetKey()))
		}
		return connect.NewError(connect.CodeInternal, err)
	}
	return nil
}

// GetOrExtendReservation returns the active reservation for a cache key,
// creating or refreshing it when the caller is allowed to own it.
//
// This is the coordination path for serialized cache population. On a cache
// miss, only the active owner should execute and publish the result; other
// callers observe the current reservation and wait for the cache entry to
// appear.
func (m *Manager) GetOrExtendReservation(ctx context.Context, request *cacheservicepb.GetOrExtendReservationRequest, now time.Time) (*cacheservicepb.Reservation, error) {
	if request.GetKey() == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("key is required"))
	}
	if request.GetOwnerId() == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("owner_id is required"))
	}

	reservationKey := fmt.Sprintf("%s:%s", reservationPrefix, request.GetKey())
	heartbeat := m.resolvedHeartbeat(request.GetHeartbeatInterval())
	reservation := &models.Reservation{
		Key:              reservationKey,
		OwnerID:          request.GetOwnerId(),
		HeartbeatSeconds: int64(heartbeat.Seconds()),
		ExpiresAt:        now.Add(heartbeat * time.Duration(m.heartbeatGracePeriodMultiplier)),
	}

	_, err := m.reservations.Get(ctx, reservationKey)
	if err != nil && !repositoryerrors.IsNotFound(err) {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Existing reservations follow one of two paths:
	//   - same owner or expired: current caller can refresh it
	//   - different active owner: keep returning the current holder
	if err == nil {
		if err := m.reservations.UpdateIfExpiredOrOwned(ctx, reservation, now); err != nil {
			if repositoryerrors.IsReservationNotClaimable(err) {
				// Another caller still owns the reservation or claimed it before we
				// could refresh an expired one. Re-read to return the current holder.
				current, getErr := m.reservations.Get(ctx, reservationKey)
				if getErr != nil {
					return nil, connect.NewError(connect.CodeInternal, getErr)
				}
				return reservationFromModel(current), nil
			}
			return nil, connect.NewError(connect.CodeInternal, err)
		}
		return reservationFromModel(reservation), nil
	}

	if err := m.reservations.Create(ctx, reservation); err != nil {
		if repositoryerrors.IsAlreadyExists(err) {
			// Another caller created the reservation after our initial read.
			current, getErr := m.reservations.Get(ctx, reservationKey)
			if getErr != nil {
				return nil, connect.NewError(connect.CodeInternal, getErr)
			}
			return reservationFromModel(current), nil
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return reservationFromModel(reservation), nil
}

// ReleaseReservation releases ownership for serialized cache population.
//
// Missing reservations are treated as already-released so callers can clean up
// idempotently.
func (m *Manager) ReleaseReservation(ctx context.Context, request *cacheservicepb.ReleaseReservationRequest) error {
	if request.GetKey() == "" {
		return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("key is required"))
	}
	if request.GetOwnerId() == "" {
		return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("owner_id is required"))
	}

	reservationKey := fmt.Sprintf("%s:%s", reservationPrefix, request.GetKey())
	if err := m.reservations.DeleteByKeyAndOwner(ctx, reservationKey, request.GetOwnerId()); err != nil {
		if repositoryerrors.IsNotFound(err) {
			return nil
		}
		return connect.NewError(connect.CodeInternal, err)
	}
	return nil
}

func (m *Manager) resolvedHeartbeat(requested *durationpb.Duration) time.Duration {
	heartbeat := m.maxReservationHeartbeatInterval
	if requested != nil && requested.AsDuration() > 0 && requested.AsDuration() < heartbeat {
		heartbeat = requested.AsDuration()
	}
	return heartbeat
}

func mergeMetadata(existing *models.CachedOutput, request *cacheservicepb.Metadata, now time.Time) *cacheservicepb.Metadata {
	var metadata *cacheservicepb.Metadata
	if request != nil {
		metadata = proto.Clone(request).(*cacheservicepb.Metadata)
	} else {
		metadata = &cacheservicepb.Metadata{}
	}

	if existing != nil && len(existing.Metadata) > 0 {
		if existingMetadata, err := unmarshalMetadata(existing.Metadata); err == nil && metadata.GetCreatedAt() == nil {
			// Keep the original CreatedAt
			metadata.CreatedAt = existingMetadata.GetCreatedAt()
		}
	}

	if metadata.GetCreatedAt() == nil {
		metadata.CreatedAt = timestamppb.New(now)
	}
	metadata.LastUpdatedAt = timestamppb.New(now)
	return metadata
}

func unmarshalMetadata(data []byte) (*cacheservicepb.Metadata, error) {
	metadata := &cacheservicepb.Metadata{}
	if len(data) == 0 {
		return metadata, nil
	}
	if err := proto.Unmarshal(data, metadata); err != nil {
		return nil, fmt.Errorf("unmarshal metadata: %w", err)
	}
	return metadata, nil
}

func reservationFromModel(model *models.Reservation) *cacheservicepb.Reservation {
	return &cacheservicepb.Reservation{
		Key:               model.Key,
		OwnerId:           model.OwnerID,
		HeartbeatInterval: durationpb.New(time.Duration(model.HeartbeatSeconds) * time.Second),
		ExpiresAt:         timestamppb.New(model.ExpiresAt),
	}
}
