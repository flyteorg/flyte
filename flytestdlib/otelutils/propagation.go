package otelutils

import (
	"connectrpc.com/otelconnect"
	"go.opentelemetry.io/otel/propagation"
)

// RPCPropagator carries W3C trace context and baggage across service calls.
func RPCPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{})
}

// NewConnectInterceptor continues incoming traces and propagates context on outgoing calls.
func NewConnectInterceptor(options ...otelconnect.Option) (*otelconnect.Interceptor, error) {
	return otelconnect.NewInterceptor(append(options,
		otelconnect.WithPropagator(RPCPropagator()),
		otelconnect.WithTrustRemote(),
	)...)
}
