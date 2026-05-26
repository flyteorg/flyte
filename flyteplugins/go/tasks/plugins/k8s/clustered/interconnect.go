package clustered

import (
	"context"

	corev1 "k8s.io/api/core/v1"

	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	clusteredpb "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/plugins"
)

// applyInterconnect mutates the pod spec for the requested interconnect fabric.
// Phase 3 will implement EFA, InfiniBand, and RoCE mutations.
func applyInterconnect(ctx context.Context, interconnect clusteredpb.Interconnect, podSpec *corev1.PodSpec) *corev1.PodSpec {
	if interconnect != clusteredpb.Interconnect_TCP {
		logger.Warningf(ctx, "interconnect %v is not yet implemented; falling back to TCP", interconnect)
	}
	return podSpec
}
