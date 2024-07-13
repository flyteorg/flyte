package impl

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc/codes"

	"github.com/flyteorg/flyte/flyteadmin/pkg/common"
	"github.com/flyteorg/flyte/flyteadmin/pkg/config"
	flyteAdminErrors "github.com/flyteorg/flyte/flyteadmin/pkg/errors"
	repoInterfaces "github.com/flyteorg/flyte/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/service"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/watch"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
)

type watchMetrics struct {
	alreadySentBufferSize *prometheus.GaugeVec
	statusUpdatesSent     *prometheus.CounterVec
	sendErrors            *prometheus.CounterVec
	readAndSendDuration   *promutils.HistogramStopWatchVec
}

type WatchManager struct {
	metrics               watchMetrics
	db                    repoInterfaces.Repository
	statusUpdatesPageSize int
	pollInterval          time.Duration

	mtx                         sync.Mutex
	activeClusterConnections    map[string]uint
	maxActiveClusterConnections uint
	nonTerminalUpdatesInterval  time.Duration
}

type sentUpdate struct {
	terminal bool
	sentAt   time.Time
}

type alreadySentMap map[uint]sentUpdate

func (e *WatchManager) WatchExecutionStatusUpdates(req *watch.WatchExecutionStatusUpdatesRequest, srv service.WatchService_WatchExecutionStatusUpdatesServer) error {
	ctx := srv.Context()

	cluster := req.GetCluster()
	if cluster == "" {
		logger.Debugf(ctx, "missing 'cluster' in watch execution status updates request")
		return flyteAdminErrors.NewFlyteAdminError(codes.InvalidArgument, "missing 'cluster'")
	}

	e.mtx.Lock()
	if e.activeClusterConnections[cluster] >= e.maxActiveClusterConnections {
		e.mtx.Unlock()
		return flyteAdminErrors.NewFlyteAdminError(codes.ResourceExhausted, "too many active watch status updates stream connections")
	}
	e.activeClusterConnections[cluster]++
	e.mtx.Unlock()
	defer func() {
		e.mtx.Lock()
		if e.activeClusterConnections[cluster] > 0 {
			e.activeClusterConnections[cluster]--
		}
		e.mtx.Unlock()
	}()

	logger.Debugf(ctx, "received request to watch execution status updates for cluster %q", cluster)

	// propeller will tolerate duplicates, this is to reduce network traffic
	alreadySent := alreadySentMap{}

	ticker := time.NewTicker(e.pollInterval)
	defer ticker.Stop()

	// find 1st terminal execution checkpoint
	checkpoint, err := e.db.ExecutionRepo().FindNextStatusUpdatesCheckpoint(ctx, cluster, 0)
	if err != nil {
		return err
	}

	for {
		checkpoint, err = e.readAndSendExecutions(ctx, cluster, checkpoint, alreadySent, srv)
		if err != nil {
			return err
		}

		select {
		case <-srv.Context().Done():
			err = srv.Context().Err()
			if errors.Is(err, context.Canceled) {
				logger.Debugf(ctx, "client closed watch execution status updates stream")
				return nil
			}
			logger.Errorf(ctx, "watch execution status updates stream connection closed with err: %v", err)
			return err

		case <-ticker.C:
		}
	}
}

func (e *WatchManager) readAndSendExecutions(
	ctx context.Context,
	cluster string,
	checkpoint uint,
	alreadySent alreadySentMap,
	srv service.WatchService_WatchExecutionStatusUpdatesServer,
) (uint, error) {
	defer e.metrics.readAndSendDuration.WithLabelValues(cluster).Start().Stop()

	newCheckpoint, err := e.db.ExecutionRepo().FindNextStatusUpdatesCheckpoint(ctx, cluster, checkpoint)
	if err != nil {
		return 0, err
	}

	offset := 0
	for {
		executions, err := e.db.ExecutionRepo().FindStatusUpdates(ctx, cluster, checkpoint, e.statusUpdatesPageSize, offset)
		if err != nil {
			return 0, err
		}

		err = e.sendStatusUpdates(ctx, executions, alreadySent, srv, cluster, newCheckpoint)
		if err != nil {
			return 0, err
		}

		if len(executions) < e.statusUpdatesPageSize {
			break
		}
		offset += e.statusUpdatesPageSize
	}
	return newCheckpoint, nil
}

func (e *WatchManager) sendStatusUpdates(
	ctx context.Context,
	executions []repoInterfaces.ExecutionStatus,
	alreadySent alreadySentMap,
	srv service.WatchService_WatchExecutionStatusUpdatesServer,
	clusterName string,
	checkpoint uint,
) error {
	defer func() {
		e.metrics.alreadySentBufferSize.WithLabelValues(clusterName).Set(float64(len(alreadySent)))
	}()

	if len(executions) == 0 {
		return nil
	}

	for _, exec := range executions {
		var closure admin.ExecutionClosure
		if err := proto.Unmarshal(exec.Closure, &closure); err != nil {
			logger.Errorf(ctx, "failed to unmarshal execution [%v] closure: %v", exec.ExecutionKey, err)
			return err
		}

		phase := closure.GetPhase()
		isTerminal := common.IsExecutionTerminal(phase)
		lastUpdate, wasAlreadyNotified := alreadySent[exec.ID]
		lastUpdateWasLongAgo := time.Since(lastUpdate.sentAt) >= e.nonTerminalUpdatesInterval
		if !wasAlreadyNotified || (!lastUpdate.terminal && (isTerminal || lastUpdateWasLongAgo)) {
			logger.Debugf(ctx, "sending execution status update for %v with %v", exec.ExecutionKey, phase.String())
			resp := &watch.WatchExecutionStatusUpdatesResponse{
				Id: &core.WorkflowExecutionIdentifier{
					Project: exec.Project,
					Domain:  exec.Domain,
					Name:    exec.Name,
					Org:     exec.Org,
				},
				Phase:     phase,
				OutputUri: closure.GetOutputs().GetUri(),
				Error:     closure.GetError(),
			}

			if err := srv.Send(resp); err != nil {
				e.metrics.sendErrors.WithLabelValues(clusterName).Inc()
				logger.Errorf(ctx, "failed to send execution status update: %v", err)
				return err
			}
			e.metrics.statusUpdatesSent.WithLabelValues(clusterName).Inc()

			alreadySent[exec.ID] = sentUpdate{terminal: isTerminal, sentAt: time.Now()}
		}

		if exec.ID <= checkpoint {
			delete(alreadySent, exec.ID)
		}
	}

	return nil
}

func NewWatchManager(cfg *config.ServerConfig, db repoInterfaces.Repository, scope promutils.Scope) *WatchManager {
	return &WatchManager{
		db:                          db,
		statusUpdatesPageSize:       cfg.WatchService.MaxPageSize,
		pollInterval:                cfg.WatchService.PollInterval.Duration,
		maxActiveClusterConnections: uint(cfg.WatchService.MaxActiveClusterConnections),
		nonTerminalUpdatesInterval:  cfg.WatchService.NonTerminalStatusUpdatesInterval.Duration,
		activeClusterConnections:    map[string]uint{},
		metrics: watchMetrics{
			alreadySentBufferSize: scope.MustNewGaugeVec("already_sent_buffer_size", "How many items does already_sent_buffer contain at the moment", "cluster_name"),
			statusUpdatesSent:     scope.MustNewCounterVec("status_updates_sent", "How many execution status updates were sent", "cluster_name"),
			sendErrors:            scope.MustNewCounterVec("send_errors", "How many times sending execution status update failed", "cluster_name"),
			readAndSendDuration:   scope.MustNewHistogramStopWatchVec("read_and_send_duration", "How long does a single fetching and sending of page of status updates take", "cluster_name"),
		},
	}
}
