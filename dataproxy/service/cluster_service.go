package service

import (
	"context"

	"connectrpc.com/connect"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/cluster"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/cluster/clusterconnect"
)

type ClusterService struct {
	clusterconnect.UnimplementedClusterServiceHandler
	dataplaneDomain string
}

func NewClusterService(dataplaneDomain string) *ClusterService {
	return &ClusterService{
		dataplaneDomain: dataplaneDomain,
	}
}

var _ clusterconnect.ClusterServiceHandler = (*ClusterService)(nil)

func (s *ClusterService) SelectCluster(
	ctx context.Context,
	req *connect.Request[cluster.SelectClusterRequest],
) (*connect.Response[cluster.SelectClusterResponse], error) {
	return connect.NewResponse(&cluster.SelectClusterResponse{
		ClusterEndpoint: s.dataplaneDomain,
	}), nil
}
