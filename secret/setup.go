package secret

import (
	"context"
	"fmt"

	"connectrpc.com/connect/v2/connecthttp"

	"connectrpc.com/connect/v2"
	"connectrpc.com/otelconnect"
	"github.com/flyteorg/flyte/v2/flytestdlib/app"
	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	"github.com/flyteorg/flyte/v2/flytestdlib/otelutils"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/secret/secretconnect"
	"github.com/flyteorg/flyte/v2/secret/config"
	"github.com/flyteorg/flyte/v2/secret/service"
)

const otelServiceName = "secret-service"

// Setup registers the SecretService handler on the SetupContext mux.
// Requires sc.K8sClient and sc.Namespace to be set.
func Setup(ctx context.Context, sc *app.SetupContext) error {
	cfg := config.GetConfig()
	sc.AddProfilingServer("secret", cfg.PprofPort)

	svc, err := service.NewSecretServiceWithConfig(ctx, sc.K8sClient, cfg.Webhook)
	if err != nil {
		return fmt.Errorf("secret: configure webhook service client: %w", err)
	}

	otelCfg := otelutils.GetConfig()
	if err := otelutils.RegisterProvidersWithContext(ctx, otelServiceName, otelCfg); err != nil {
		return fmt.Errorf("registering otel providers: %w", err)
	}
	otelInterceptor, err := otelconnect.NewServerInterceptor(
		otelconnect.WithTracerProvider(otelutils.GetTracerProvider(otelServiceName)),
		otelconnect.WithMeterProvider(otelutils.GetMeterProvider(otelServiceName)),
	)
	if err != nil {
		return fmt.Errorf("creating otel interceptor: %w", err)
	}

	handler := connect.NewServer(otelInterceptor)
	secretconnect.RegisterSecretServiceHandler(handler, svc)
	connecthttp.Mount(sc.Mux, handler, connecthttp.WithReadMaxBytes(0))
	path := "/" + secretconnect.SecretServiceName + "/"
	logger.Infof(ctx, "Mounted SecretService at %s", path)

	return nil
}
