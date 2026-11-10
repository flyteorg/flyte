package service

import (
	"context"
	"fmt"
	"hash/fnv"
	"math/rand"
	"net/http"

	"connectrpc.com/connect"

	"github.com/flyteorg/flyte/v2/flytestdlib/random"
	"github.com/flyteorg/flyte/v2/flytestdlib/serviceclient"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/actions"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/actions/actionsconnect"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/task"
	"github.com/flyteorg/flyte/v2/runs/config"
	"github.com/flyteorg/flyte/v2/runs/repository/models"
)

// runGetter reads a run's stored record; interfaces.ActionRepo implements it.
type runGetter interface {
	GetRun(ctx context.Context, runID *common.RunIdentifier) (*models.Run, error)
}

// ActionsRouter is an actions service client that sends each run's calls to the actions service
// of the cluster the run was placed on. Child actions never pass through it: a parent's pod
// enqueues them on its own cluster's actions service, so a whole run stays in one cluster.
type ActionsRouter struct {
	home     string
	clusters map[string]actionsconnect.ActionsServiceClient
	queues   map[string]random.WeightedRandomList
	members  map[string][]string // every cluster of each queue, in config order
	runs     runGetter
}

var _ actionsconnect.ActionsServiceClient = (*ActionsRouter)(nil)

// NewActionsRouter builds a client for each configured cluster; home is the home cluster's client.
func NewActionsRouter(ctx context.Context, cfg *config.MultiClusterConfig, home actionsconnect.ActionsServiceClient,
	runs runGetter, opts ...connect.ClientOption) (*ActionsRouter, error) {
	clients := map[string]actionsconnect.ActionsServiceClient{cfg.HomeCluster: home}
	for name, svc := range cfg.Clusters {
		if name == cfg.HomeCluster {
			return nil, fmt.Errorf("cluster %q is the home cluster, which uses runs.actionsService", name)
		}
		httpClient, err := serviceclient.NewHTTPClient(ctx, http.DefaultClient, svc)
		if err != nil {
			return nil, fmt.Errorf("cluster %q: %w", name, err)
		}
		clients[name] = actionsconnect.NewActionsServiceClient(httpClient, svc.URL, opts...)
	}
	return newActionsRouter(ctx, cfg, clients, runs)
}

func newActionsRouter(ctx context.Context, cfg *config.MultiClusterConfig,
	clients map[string]actionsconnect.ActionsServiceClient, runs runGetter) (*ActionsRouter, error) {
	r := &ActionsRouter{
		home:     cfg.HomeCluster,
		clusters: clients,
		queues:   map[string]random.WeightedRandomList{},
		members:  map[string][]string{},
		runs:     runs,
	}
	for queue, entities := range cfg.QueueClusterMap {
		entries := make([]random.Entry, 0, len(entities))
		for _, e := range entities {
			if _, ok := clients[e.ID]; !ok {
				return nil, fmt.Errorf("queue %q: unknown cluster %q", queue, e.ID)
			}
			entries = append(entries, random.Entry{Item: clusterName(e.ID), Weight: e.Weight})
			r.members[queue] = append(r.members[queue], e.ID)
		}
		list, err := random.NewWeightedRandom(ctx, entries)
		if err != nil {
			return nil, fmt.Errorf("queue %q: %w", queue, err)
		}
		r.queues[queue] = list
	}
	return r, nil
}

// clusterName lets cluster names go in a random.WeightedRandomList, which needs a stable order.
type clusterName string

func (c clusterName) Compare(to random.Comparable) bool { return c < to.(clusterName) }

// placement returns the clusters a run may be on, the one it was placed on first. The others are
// the queue's remaining clusters, for runs started before the queue's clusters or weights changed.
func (r *ActionsRouter) placement(runName string, spec *task.RunSpec) ([]string, error) {
	queue := spec.GetQueue()
	if queue == "" {
		return []string{r.home}, nil
	}
	list, ok := r.queues[queue]
	if !ok {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("unknown queue %q", queue))
	}
	// Seeded by the run name, as v1 seeds by execution name, so every call for a run picks the same cluster.
	h := fnv.New64a()
	_, _ = h.Write([]byte(runName))
	picked, err := list.GetWithSeed(rand.NewSource(int64(h.Sum64()))) // #nosec G115
	if err != nil {
		return nil, err
	}
	clusters := []string{string(picked.(clusterName))}
	for _, c := range r.members[queue] {
		if c != clusters[0] {
			clusters = append(clusters, c)
		}
	}
	return clusters, nil
}

// runPlacement is the placement of a started run, read from its stored run spec.
func (r *ActionsRouter) runPlacement(ctx context.Context, runID *common.RunIdentifier) ([]string, error) {
	run, err := r.runs.GetRun(ctx, runID)
	if err != nil {
		return nil, fmt.Errorf("load run %s: %w", runID.GetName(), err)
	}
	return r.placement(runID.GetName(), extractRunSpecFromActionModel(ctx, run))
}

// tryEach calls the run's cluster first, then the queue's other clusters while the action is
// not found, which the actions service reports as isAlreadyTerminated recognizes.
func (r *ActionsRouter) tryEach(ctx context.Context, runID *common.RunIdentifier,
	call func(actionsconnect.ActionsServiceClient) error) error {
	clusters, err := r.runPlacement(ctx, runID)
	if err != nil {
		return err
	}
	for _, c := range clusters {
		if err = call(r.clusters[c]); err == nil || !isAlreadyTerminated(err) {
			return err
		}
	}
	return err
}

func (r *ActionsRouter) Enqueue(ctx context.Context, req *connect.Request[actions.EnqueueRequest]) (
	*connect.Response[actions.EnqueueResponse], error) {
	clusters, err := r.placement(req.Msg.GetAction().GetActionId().GetRun().GetName(), req.Msg.GetRunSpec())
	if err != nil {
		return nil, err
	}
	return r.clusters[clusters[0]].Enqueue(ctx, req)
}

func (r *ActionsRouter) Abort(ctx context.Context, req *connect.Request[actions.AbortRequest]) (
	resp *connect.Response[actions.AbortResponse], err error) {
	err = r.tryEach(ctx, req.Msg.GetActionId().GetRun(), func(c actionsconnect.ActionsServiceClient) (callErr error) {
		resp, callErr = c.Abort(ctx, req)
		return callErr
	})
	return resp, err
}

func (r *ActionsRouter) Signal(ctx context.Context, req *connect.Request[actions.SignalRequest]) (
	resp *connect.Response[actions.SignalResponse], err error) {
	err = r.tryEach(ctx, req.Msg.GetActionId().GetRun(), func(c actionsconnect.ActionsServiceClient) (callErr error) {
		resp, callErr = c.Signal(ctx, req)
		return callErr
	})
	return resp, err
}

// Update and WatchForUpdates come from task pods to their local actions service, never through
// runs; they follow the run's placement only to complete the interface.

func (r *ActionsRouter) Update(ctx context.Context, req *connect.Request[actions.UpdateRequest]) (
	*connect.Response[actions.UpdateResponse], error) {
	clusters, err := r.runPlacement(ctx, req.Msg.GetActionId().GetRun())
	if err != nil {
		return nil, err
	}
	return r.clusters[clusters[0]].Update(ctx, req)
}

func (r *ActionsRouter) WatchForUpdates(ctx context.Context, req *connect.Request[actions.WatchForUpdatesRequest]) (
	*connect.ServerStreamForClient[actions.WatchForUpdatesResponse], error) {
	clusters, err := r.runPlacement(ctx, req.Msg.GetParentActionId().GetRun())
	if err != nil {
		return nil, err
	}
	return r.clusters[clusters[0]].WatchForUpdates(ctx, req)
}
