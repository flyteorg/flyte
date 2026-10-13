package service

import (
	"context"
	"sync"

	"connectrpc.com/connect/v2"
	"golang.org/x/sync/semaphore"

	appk8s "github.com/flyteorg/flyte/v2/app/internal/k8s"
	flyteapp "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/app"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/app/appconnect"
)

const defaultMaxConcurrentLogStreams = 100

// AppLogStreamer streams pod logs for an app replica.
type AppLogStreamer interface {
	TailLogs(ctx context.Context, replicaID *flyteapp.ReplicaIdentifier, send func(*flyteapp.LogLines) error) error
}

// InternalAppLogsService is the data plane implementation of AppLogsService.
// It resolves an app_id or replica_id to the backing pods and streams logs.
type InternalAppLogsService struct {
	appconnect.UnimplementedAppLogsServiceHandler
	k8s      appk8s.AppK8sClientInterface
	streamer AppLogStreamer
	sem      *semaphore.Weighted
}

// NewInternalAppLogsService creates a new InternalAppLogsService.
func NewInternalAppLogsService(k8sClient appk8s.AppK8sClientInterface, streamer AppLogStreamer) *InternalAppLogsService {
	return &InternalAppLogsService{
		k8s:      k8sClient,
		streamer: streamer,
		sem:      semaphore.NewWeighted(defaultMaxConcurrentLogStreams),
	}
}

var _ appconnect.AppLogsServiceHandler = (*InternalAppLogsService)(nil)

// TailLogs streams log lines for the given app or replica.
func (s *InternalAppLogsService) TailLogs(
	ctx context.Context,
	req *flyteapp.TailLogsRequest,
	stream appconnect.AppLogsServiceTailLogsServerStream,
) error {
	if !s.sem.TryAcquire(1) {
		return connect.Errorf(connect.CodeResourceExhausted, "too many concurrent log streams")
	}
	defer s.sem.Release(1)

	replicas, err := s.resolveReplicas(ctx, req)
	if err != nil {
		return err
	}
	if len(replicas) == 0 {
		return connect.Errorf(connect.CodeNotFound, "no replicas found")
	}

	// Send initial replica list so the client knows which replicas will be streamed.
	if err := stream.Send(&flyteapp.TailLogsResponse{
		Resp: &flyteapp.TailLogsResponse_Replicas{
			Replicas: &flyteapp.ReplicaIdentifierList{Replicas: replicas},
		},
	}); err != nil {
		return err
	}

	// Serialize stream.Send across goroutines — connect server streams are not
	// safe for concurrent writes.
	var sendMu sync.Mutex
	send := func(replicaID *flyteapp.ReplicaIdentifier) func(*flyteapp.LogLines) error {
		return func(logs *flyteapp.LogLines) error {
			logs.ReplicaId = replicaID
			sendMu.Lock()
			defer sendMu.Unlock()
			return stream.Send(&flyteapp.TailLogsResponse{
				Resp: &flyteapp.TailLogsResponse_Batches{
					Batches: &flyteapp.LogLinesBatch{Logs: []*flyteapp.LogLines{logs}},
				},
			})
		}
	}

	if len(replicas) == 1 {
		return s.streamer.TailLogs(ctx, replicas[0], send(replicas[0]))
	}

	streamCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	var wg sync.WaitGroup
	errCh := make(chan error, len(replicas))
	for _, r := range replicas {
		wg.Add(1)
		go func(replicaID *flyteapp.ReplicaIdentifier) {
			defer wg.Done()
			if err := s.streamer.TailLogs(streamCtx, replicaID, send(replicaID)); err != nil {
				errCh <- err
				cancel()
			}
		}(r)
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		if err != nil && ctx.Err() == nil {
			return err
		}
	}
	return nil
}

// resolveReplicas expands a TailLogsRequest target into a list of ReplicaIdentifiers.
func (s *InternalAppLogsService) resolveReplicas(ctx context.Context, req *flyteapp.TailLogsRequest) ([]*flyteapp.ReplicaIdentifier, error) {
	switch t := req.GetTarget().(type) {
	case *flyteapp.TailLogsRequest_AppId:
		if t.AppId == nil {
			return nil, connect.Errorf(connect.CodeInvalidArgument, "app_id is required")
		}
		replicas, err := s.k8s.GetReplicas(ctx, t.AppId)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, err.Error()).WithCause(err)
		}
		ids := make([]*flyteapp.ReplicaIdentifier, 0, len(replicas))
		for _, r := range replicas {
			ids = append(ids, r.GetMetadata().GetId())
		}
		return ids, nil
	case *flyteapp.TailLogsRequest_ReplicaId:
		if t.ReplicaId == nil || t.ReplicaId.GetAppId() == nil || t.ReplicaId.GetName() == "" {
			return nil, connect.Errorf(connect.CodeInvalidArgument, "replica_id with app_id and name is required")
		}
		return []*flyteapp.ReplicaIdentifier{t.ReplicaId}, nil
	default:
		return nil, connect.Errorf(connect.CodeInvalidArgument, "target (app_id or replica_id) is required")
	}
}
