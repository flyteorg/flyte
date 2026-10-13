package secret

import (
	"context"
	"fmt"
	"net/http"

	"connectrpc.com/connect/v2"
	"connectrpc.com/connect/v2/connecthttp"
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

	handlerServer := connect.NewServer(otelInterceptor)
	secretconnect.RegisterSecretServiceHandler(handlerServer, svc)
	path := "/" + secretconnect.SecretServiceName + "/"
	handler := http.NewServeMux()
	connecthttp.Mount(handler, handlerServer, connecthttp.WithReadMaxBytes(0))
	sc.Mux.Handle(path, handler)
	logger.Infof(ctx, "Mounted SecretService at %s", path)

	return nil
}
