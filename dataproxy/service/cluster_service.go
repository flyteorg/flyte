package service

import (
	"context"

	"connectrpc.com/connect"
	"github.com/flyteorg/flyte/v2/flytestdlib/logger"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/cluster"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/cluster/clusterconnect"
)

type ClusterService struct {
	clusterconnect.UnimplementedClusterServiceHandler
}

func NewClusterService() *ClusterService {
	return &ClusterService{}
}

var _ clusterconnect.ClusterServiceHandler = (*ClusterService)(nil)

func (s *ClusterService) SelectCluster(
	ctx context.Context,
	req *connect.Request[cluster.SelectClusterRequest],
) (*connect.Response[cluster.SelectClusterResponse], error) {
	requestHost := req.Header().Get("Host")
	logger.Debugf(ctx, "Request Host: %s", requestHost)
	return connect.NewResponse(&cluster.SelectClusterResponse{
		ClusterEndpoint: requestHost,
	}), nil
}
