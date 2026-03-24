package controller

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"

	flyteorgv1 "github.com/flyteorg/flyte/v2/executor/api/v1"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/catalog"
	pluginsCore "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/ioutils"
	corepb "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
)

const cacheReservationHeartbeatInterval = TaskActionDefaultRequeueDuration

type taskCacheConfig struct {
	key          catalog.Key
	ownerID      string
	serializable bool
}

func (r *TaskActionReconciler) evaluateCacheBeforeExecution(
	ctx context.Context,
	taskAction *flyteorgv1.TaskAction,
	tCtx pluginsCore.TaskExecutionContext,
) (pluginsCore.Transition, bool, error) {
	cacheCfg, ok, err := buildTaskCacheConfig(ctx, taskAction, tCtx)
	if err != nil || !ok || r.Catalog == nil {
		return pluginsCore.UnknownTransition, false, err
	}

	entry, err := r.Catalog.Get(ctx, cacheCfg.key)
	if err == nil {
		if err := tCtx.OutputWriter().Put(ctx, entry.GetOutputs()); err != nil {
			return pluginsCore.UnknownTransition, false, fmt.Errorf("persisting cached outputs: %w", err)
		}

		info := cacheTaskInfo(corepb.CatalogCacheStatus_CACHE_HIT, "cache hit")
		return pluginsCore.DoTransition(pluginsCore.PhaseInfoSuccess(info)), true, nil
	}

	if !catalog.IsNotFound(err) {
		log.FromContext(ctx).Error(err, "cache lookup failed, continuing with task execution")
		return pluginsCore.UnknownTransition, false, nil
	}

	if cacheCfg.serializable {
		reservation, err := r.Catalog.GetOrExtendReservation(ctx, cacheCfg.key, cacheCfg.ownerID, cacheReservationHeartbeatInterval)
		if err != nil {
			return pluginsCore.UnknownTransition, false, fmt.Errorf("acquiring cache reservation: %w", err)
		}
		if reservation.GetOwnerId() == cacheCfg.ownerID {
			return pluginsCore.UnknownTransition, false, nil
		}

		info := cacheTaskInfo(corepb.CatalogCacheStatus_CACHE_MISS, "waiting for serialized cache owner")
		phaseInfo := pluginsCore.PhaseInfoWaitingForCache(taskAction.Status.PluginPhaseVersion, info)
		phaseInfo.WithReason(fmt.Sprintf("waiting for cache to be populated by reservation owner %q", reservation.GetOwnerId()))
		return pluginsCore.DoTransition(phaseInfo), true, nil
	}

	return pluginsCore.UnknownTransition, false, nil
}

func (r *TaskActionReconciler) finalizeCacheAfterExecution(
	ctx context.Context,
	taskAction *flyteorgv1.TaskAction,
	tCtx pluginsCore.TaskExecutionContext,
	transition pluginsCore.Transition,
	cacheShortCircuited bool,
) (pluginsCore.Transition, error) {
	cacheCfg, ok, err := buildTaskCacheConfig(ctx, taskAction, tCtx)
	if err != nil || !ok || r.Catalog == nil {
		return transition, err
	}

	if cacheShortCircuited {
		return transition, nil
	}

	phase := transition.Info().Phase()
	if phase.IsSuccess() {
		status := corepb.CatalogCacheStatus_CACHE_POPULATED
		if err := r.writeTaskOutputsToCache(ctx, tCtx, cacheCfg.key); err != nil {
			if grpcstatus.Code(err) == codes.AlreadyExists {
				status = corepb.CatalogCacheStatus_CACHE_POPULATED
			} else {
				log.FromContext(ctx).Error(err, "failed to write task outputs to cache")
				status = corepb.CatalogCacheStatus_CACHE_PUT_FAILURE
			}
		}
		appendCacheStatus(transition.Info().Info(), status)
		if cacheCfg.serializable {
			if err := r.releaseCacheReservation(ctx, cacheCfg); err != nil {
				log.FromContext(ctx).Error(err, "failed to release cache reservation after success")
			}
		}
		return transition, nil
	}

	if cacheCfg.serializable && (phase.IsFailure() || phase.IsAborted()) {
		if err := r.releaseCacheReservation(ctx, cacheCfg); err != nil {
			log.FromContext(ctx).Error(err, "failed to release cache reservation after terminal failure")
		}
	}

	return transition, nil
}

func buildTaskCacheConfig(
	ctx context.Context,
	taskAction *flyteorgv1.TaskAction,
	tCtx pluginsCore.TaskExecutionContext,
) (*taskCacheConfig, bool, error) {
	taskTemplate, err := tCtx.TaskReader().Read(ctx)
	if err != nil {
		return nil, false, fmt.Errorf("reading task template for cache handling: %w", err)
	}

	metadata := taskTemplate.GetMetadata()
	if metadata == nil || !metadata.GetDiscoverable() {
		return nil, false, nil
	}
	if taskTemplate.GetId() == nil {
		return nil, false, fmt.Errorf("discoverable task is missing identifier")
	}

	key := catalog.Key{
		Identifier:           proto.Clone(taskTemplate.GetId()).(*corepb.Identifier),
		CacheVersion:         taskAction.Spec.CacheKey,
		CacheIgnoreInputVars: metadata.GetCacheIgnoreInputVars(),
		TypedInterface:       taskTemplate.GetInterface(),
		InputReader:          tCtx.InputReader(),
	}

	return &taskCacheConfig{
		key:          key,
		ownerID:      cacheReservationOwnerID(taskAction),
		serializable: metadata.GetCacheSerializable(),
	}, true, nil
}

func cacheReservationOwnerID(taskAction *flyteorgv1.TaskAction) string {
	return types.NamespacedName{Name: taskAction.Name, Namespace: taskAction.Namespace}.String()
}

func (r *TaskActionReconciler) writeTaskOutputsToCache(ctx context.Context, tCtx pluginsCore.TaskExecutionContext, key catalog.Key) error {
	outputPaths := ioutils.NewReadOnlyOutputFilePaths(ctx, r.DataStore, tCtx.OutputWriter().GetOutputPrefixPath())
	outputReader := ioutils.NewRemoteFileOutputReader(ctx, r.DataStore, outputPaths, 0)
	_, err := r.Catalog.Put(ctx, key, outputReader, cacheMetadataForUpload(tCtx, key.Identifier))
	return err
}

func (r *TaskActionReconciler) releaseCacheReservation(ctx context.Context, cacheCfg *taskCacheConfig) error {
	if r.Catalog == nil || cacheCfg == nil || !cacheCfg.serializable {
		return nil
	}

	return r.Catalog.ReleaseReservation(ctx, cacheCfg.key, cacheCfg.ownerID)
}

func cacheMetadataForUpload(tCtx pluginsCore.TaskExecutionContext, taskID *corepb.Identifier) catalog.Metadata {
	taskExecID := proto.Clone(tCtx.TaskExecutionMetadata().GetTaskExecutionID().GetID()).(*corepb.TaskExecutionIdentifier)
	taskExecID.TaskId = proto.Clone(taskID).(*corepb.Identifier)

	return catalog.Metadata{
		WorkflowExecutionIdentifier: taskExecID.GetNodeExecutionId().GetExecutionId(),
		NodeExecutionIdentifier:     taskExecID.GetNodeExecutionId(),
		TaskExecutionIdentifier:     taskExecID,
		CreatedAt:                   timestamppb.Now(),
	}
}

func cacheTaskInfo(status corepb.CatalogCacheStatus, reason string) *pluginsCore.TaskInfo {
	now := time.Now()
	return &pluginsCore.TaskInfo{
		OccurredAt: &now,
		ExternalResources: []*pluginsCore.ExternalResource{
			{
				ExternalID:  "cache",
				CacheStatus: status,
			},
		},
		AdditionalReasons: []pluginsCore.ReasonInfo{
			{Reason: reason, OccurredAt: &now},
		},
	}
}

func appendCacheStatus(info *pluginsCore.TaskInfo, status corepb.CatalogCacheStatus) {
	if info == nil {
		return
	}
	info.ExternalResources = append(info.ExternalResources, &pluginsCore.ExternalResource{
		ExternalID:  "cache",
		CacheStatus: status,
	})
}
