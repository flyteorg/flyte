package grpcutils

import (
	"sync"

	grpcprometheus "github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus"
	"github.com/prometheus/client_golang/prometheus"
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
		defaultGrpcClientMetrics = grpcprometheus.NewClientMetrics(opts...)
		prometheus.MustRegister(defaultGrpcClientMetrics)
	})
	return defaultGrpcClientMetrics
}

// GrpcServerMetrics returns the grpc server metrics singleton initialized on first call.
// opts are ignored if already initialized.
func GrpcServerMetrics(opts ...grpcprometheus.ServerMetricsOption) *grpcprometheus.ServerMetrics {
	initDefaultGrpcServerMetricsOnce.Do(func() {
		defaultGrpcServerMetrics = grpcprometheus.NewServerMetrics(opts...)
		prometheus.MustRegister(defaultGrpcServerMetrics)
	})
	return defaultGrpcServerMetrics
}
