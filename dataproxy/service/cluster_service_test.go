package service

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/cluster"
)

func TestSelectCluster_ReturnsSchemeQualifiedEndpoint(t *testing.T) {
	svc := NewClusterService()

	tests := []struct {
		name     string
		host     string
		fwdProto string
		want     string
	}{
		{name: "ALB sets X-Forwarded-Proto https", host: "development.uniondemo.run", fwdProto: "https", want: "https://development.uniondemo.run"},
		{name: "devbox direct http (no forwarded proto)", host: "localhost:8090", want: "http://localhost:8090"},
		{name: "explicit http forwarded proto", host: "localhost:8090", fwdProto: "http", want: "http://localhost:8090"},
		{name: "takes first of proxy-chain list", host: "development.uniondemo.run", fwdProto: "https, http", want: "https://development.uniondemo.run"},
		{name: "normalizes uppercase forwarded proto", host: "development.uniondemo.run", fwdProto: "HTTPS", want: "https://development.uniondemo.run"},
		{name: "blank forwarded proto defaults to http", host: "localhost:8090", fwdProto: "  ", want: "http://localhost:8090"},
		{name: "empty host yields empty endpoint", host: "", want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := connect.NewRequest(&cluster.SelectClusterRequest{})
			if tt.host != "" {
				req.Header().Set("Host", tt.host)
			}
			if tt.fwdProto != "" {
				req.Header().Set("X-Forwarded-Proto", tt.fwdProto)
			}
			resp, err := svc.SelectCluster(context.Background(), req)
			require.NoError(t, err)
			assert.Equal(t, tt.want, resp.Msg.GetClusterEndpoint())
		})
	}
}
