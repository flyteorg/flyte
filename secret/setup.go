package secret

import (
	"context"

	"github.com/flyteorg/flyte/v2/flytestdlib/app"
	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/secret/secretconnect"
	"github.com/flyteorg/flyte/v2/secret/service"
)

// Setup registers the SecretService handler on the SetupContext mux.
// Requires sc.K8sClient and sc.Namespace to be set.
func Setup(ctx context.Context, sc *app.SetupContext) error {
	svc := service.NewSecretService(sc.K8sClient)

	path, handler := secretconnect.NewSecretServiceHandler(svc)
	sc.Mux.Handle(path, handler)
	logger.Infof(ctx, "Mounted SecretService at %s", path)

	return nil
}
