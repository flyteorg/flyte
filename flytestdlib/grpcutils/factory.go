package grpcutils

import (
	"sync"

	grpcprometheus "github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	namespace = "v2"
)

var (
	defaultGrpcClientMetrics         *grpcprometheus.ClientMetrics
	initDefaultGrpcClientMetricsOnce sync.Once

	defaultGrpcServerMetrics         *grpcprometheus.ServerMetrics
	initDefaultGrpcServerMetricsOnce sync.Once
)

// GrpcClientMetrics returns the grpc client metrics singleton initialized on first call.
// opts are ignored if already initialized.
func GrpcClientMetrics(opts ...grpcprometheus.ClientMetricsOption) *grpcprometheus.ClientMetrics {
	initDefaultGrpcClientMetricsOnce.Do(func() {
		v2Opts := append([]grpcprometheus.ClientMetricsOption{
			grpcprometheus.WithClientCounterOptions(grpcprometheus.WithNamespace(namespace)),
		}, opts...)
		defaultGrpcClientMetrics = grpcprometheus.NewClientMetrics(v2Opts...)
		prometheus.MustRegister(defaultGrpcClientMetrics)
	})
	return defaultGrpcClientMetrics
}

// GrpcServerMetrics returns the grpc server metrics singleton initialized on first call.
// opts are ignored if already initialized.
func GrpcServerMetrics(opts ...grpcprometheus.ServerMetricsOption) *grpcprometheus.ServerMetrics {
	initDefaultGrpcServerMetricsOnce.Do(func() {
		v2Opts := append([]grpcprometheus.ServerMetricsOption{
			grpcprometheus.WithServerCounterOptions(grpcprometheus.WithNamespace(namespace)),
		}, opts...)
		defaultGrpcServerMetrics = grpcprometheus.NewServerMetrics(v2Opts...)
		prometheus.MustRegister(defaultGrpcServerMetrics)
	})
	return defaultGrpcServerMetrics
}
