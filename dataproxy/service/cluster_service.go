package service

import (
	"context"
	"strings"

	"connectrpc.com/connect/v2"

	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/cluster"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/cluster/clusterconnect"
)

const (
	// headerXForwardedProto is the header a TLS-terminating proxy/LB sets to the scheme
	// the client used.
	headerXForwardedProto = "X-Forwarded-Proto"
	schemeHTTP            = "http"
	schemeHTTPS           = "https"
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
	req *cluster.SelectClusterRequest,
) (*cluster.SelectClusterResponse, error) {
	info, _ := connect.CallInfoForServerContext(ctx)
	headers := &connect.Header{}
	if info != nil {
		headers = info.RequestHeader()
	}
	requestHost := headers.Get("Host")
	logger.Debugf(ctx, "Request Host: %s", requestHost)

	// Return a scheme-qualified base URL, not a bare host. Clients (the console) use
	// ClusterEndpoint directly as a Connect transport base URL; a value without a scheme
	// is treated by the browser as a relative path and resolved against the current page,
	// producing a 404. Prefer X-Forwarded-Proto, which any TLS-terminating proxy/LB sets
	// (e.g. the ALB on the dev cluster → https). When it's absent the request reached us
	// directly with no proxy (e.g. a local devbox over http), so default to http.
	var endpoint string
	host := strings.TrimSpace(requestHost)
	if host != "" {
		scheme := schemeHTTP
		if proto := headers.Get(headerXForwardedProto); proto != "" {
			// X-Forwarded-Proto may be a comma-separated list (proxy chain); the first value
			// is the scheme the client used.
			candidate := strings.ToLower(strings.TrimSpace(strings.Split(proto, ",")[0]))
			if candidate == schemeHTTP || candidate == schemeHTTPS {
				scheme = candidate
			}
		}
		endpoint = scheme + "://" + host
	}

	return &cluster.SelectClusterResponse{
		ClusterEndpoint: endpoint,
	}, nil
}
