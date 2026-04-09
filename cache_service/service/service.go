package service

import (
	"context"
	"fmt"
	"time"

	"connectrpc.com/connect"
	cacheconfig "github.com/flyteorg/flyte/v2/cache_service/config"
	"github.com/flyteorg/flyte/v2/cache_service/manager"
	"github.com/flyteorg/flyte/v2/cache_service/repository"
	cacheservicepb "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/cacheservice"
	cacheservicev2 "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/cacheservice/v2"
	"github.com/jmoiron/sqlx"
)

type CacheService struct {
	manager *manager.Manager
}

func NewCacheService(cfg *cacheconfig.Config, db *sqlx.DB) *CacheService {
	repos := repository.NewRepository(db)
	return &CacheService{
		manager: manager.New(
			cfg,
			repos.CachedOutputRepo(),
			repos.ReservationRepo(),
		),
	}
}

func (s *CacheService) Get(ctx context.Context, req *connect.Request[cacheservicev2.GetCacheRequest]) (*connect.Response[cacheservicepb.GetCacheResponse], error) {
	if err := validateRequest(req.Msg); err != nil {
		return nil, err
	}
	base := req.Msg.GetBaseRequest()
	if base == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("base_request is required"))
	}
	if err := validateBaseKey(base.GetKey()); err != nil {
		return nil, err
	}
	entry, err := s.manager.Get(ctx, &cacheservicepb.GetCacheRequest{Key: scopedKey(base.GetKey(), req.Msg.GetIdentifier())})
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&cacheservicepb.GetCacheResponse{
		Output: &cacheservicepb.CachedOutput{
			Output:   &cacheservicepb.CachedOutput_OutputUri{OutputUri: entry.OutputURI},
			Metadata: entry.Metadata,
		},
	}), nil
}

func (s *CacheService) Put(ctx context.Context, req *connect.Request[cacheservicev2.PutCacheRequest]) (*connect.Response[cacheservicepb.PutCacheResponse], error) {
	if err := validateRequest(req.Msg); err != nil {
		return nil, err
	}
	base := req.Msg.GetBaseRequest()
	if base == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("base_request is required"))
	}
	if err := validateBaseKey(base.GetKey()); err != nil {
		return nil, err
	}
	if base.GetOutput() == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("base_request.output is required"))
	}
	managerReq := &cacheservicepb.PutCacheRequest{
		Key:       scopedKey(base.GetKey(), req.Msg.GetIdentifier()),
		Output:    base.GetOutput(),
		Overwrite: base.GetOverwrite(),
	}
	if err := s.manager.Put(ctx, managerReq); err != nil {
		return nil, err
	}
	return connect.NewResponse(&cacheservicepb.PutCacheResponse{}), nil
}

func (s *CacheService) Delete(ctx context.Context, req *connect.Request[cacheservicev2.DeleteCacheRequest]) (*connect.Response[cacheservicepb.DeleteCacheResponse], error) {
	if err := validateRequest(req.Msg); err != nil {
		return nil, err
	}
	base := req.Msg.GetBaseRequest()
	if base == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("base_request is required"))
	}
	if err := validateBaseKey(base.GetKey()); err != nil {
		return nil, err
	}
	if err := s.manager.Delete(ctx, &cacheservicepb.DeleteCacheRequest{Key: scopedKey(base.GetKey(), req.Msg.GetIdentifier())}); err != nil {
		return nil, err
	}
	return connect.NewResponse(&cacheservicepb.DeleteCacheResponse{}), nil
}

func (s *CacheService) GetOrExtendReservation(ctx context.Context, req *connect.Request[cacheservicev2.GetOrExtendReservationRequest]) (*connect.Response[cacheservicepb.GetOrExtendReservationResponse], error) {
	if err := validateRequest(req.Msg); err != nil {
		return nil, err
	}
	base := req.Msg.GetBaseRequest()
	if base == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("base_request is required"))
	}
	if err := validateBaseKey(base.GetKey()); err != nil {
		return nil, err
	}
	if err := validateOwnerID(base.GetOwnerId()); err != nil {
		return nil, err
	}
	managerReq := &cacheservicepb.GetOrExtendReservationRequest{
		Key:               scopedKey(base.GetKey(), req.Msg.GetIdentifier()),
		OwnerId:           base.GetOwnerId(),
		HeartbeatInterval: base.GetHeartbeatInterval(),
	}
	reservation, err := s.manager.GetOrExtendReservation(ctx, managerReq, time.Now().UTC())
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&cacheservicepb.GetOrExtendReservationResponse{
		Reservation: reservation,
	}), nil
}

func (s *CacheService) ReleaseReservation(ctx context.Context, req *connect.Request[cacheservicev2.ReleaseReservationRequest]) (*connect.Response[cacheservicepb.ReleaseReservationResponse], error) {
	if err := validateRequest(req.Msg); err != nil {
		return nil, err
	}
	base := req.Msg.GetBaseRequest()
	if base == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("base_request is required"))
	}
	if err := validateBaseKey(base.GetKey()); err != nil {
		return nil, err
	}
	if err := validateOwnerID(base.GetOwnerId()); err != nil {
		return nil, err
	}
	managerReq := &cacheservicepb.ReleaseReservationRequest{
		Key:     scopedKey(base.GetKey(), req.Msg.GetIdentifier()),
		OwnerId: base.GetOwnerId(),
	}
	if err := s.manager.ReleaseReservation(ctx, managerReq); err != nil {
		return nil, err
	}
	return connect.NewResponse(&cacheservicepb.ReleaseReservationResponse{}), nil
}

type validatableRequest interface {
	Validate() error
}

func validateRequest(msg validatableRequest) error {
	if err := msg.Validate(); err != nil {
		return connect.NewError(connect.CodeInvalidArgument, err)
	}
	return nil
}

func validateBaseKey(key string) error {
	if key == "" {
		return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("base_request.key is required"))
	}
	return nil
}

func validateOwnerID(ownerID string) error {
	if ownerID == "" {
		return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("base_request.owner_id is required"))
	}
	return nil
}

func scopedKey(key string, id *cacheservicev2.Identifier) string {
	return fmt.Sprintf("%s-%s-%s", id.GetProject(), id.GetDomain(), key)
}
