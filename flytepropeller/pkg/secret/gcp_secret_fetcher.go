package secret

import (
	"context"
	"fmt"

	gcpsmpb "cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/flyteorg/flyte/flytepropeller/pkg/secret/config"
	stdlibErrors "github.com/flyteorg/flyte/flytestdlib/errors"
	"github.com/flyteorg/flyte/flytestdlib/logger"
)

type GCPSecretFetcher struct {
	client GCPSecretsIface
	cfg    config.GCPConfig
}

func (g GCPSecretFetcher) Get(ctx context.Context, key string) (string, error) {
	return g.GetSecretValue(ctx, key)
}

func (g GCPSecretFetcher) GetSecretValue(ctx context.Context, secretID string) (string, error) {
	logger.Infof(ctx, "Got fetch secret Request for %v!", secretID)
	resp, err := g.client.AccessSecretVersion(ctx, &gcpsmpb.AccessSecretVersionRequest{
		Name: fmt.Sprintf(GCPSecretNameFormat, g.cfg.Project, secretID),
	})
	if err != nil {
		if s, ok := status.FromError(err); ok && s.Code() == codes.NotFound {
			wrappedErr := stdlibErrors.Wrapf(ErrCodeSecretNotFound, err, fmt.Sprintf(SecretNotFoundErrorFormat, secretID))
			logger.Warn(ctx, wrappedErr)
			return "", wrappedErr
		}
		wrappedErr := stdlibErrors.Wrapf(ErrCodeSecretReadFailure, err, fmt.Sprintf(SecretReadFailureErrorFormat, secretID))
		logger.Error(ctx, wrappedErr)
		return "", wrappedErr
	}
	if resp.GetPayload() == nil {
		wrappedErr := stdlibErrors.Wrapf(ErrCodeSecretNil, err, fmt.Sprintf(SecretNilErrorFormat, secretID))
		logger.Error(ctx, wrappedErr)
		return "", wrappedErr
	}
	return string(resp.GetPayload().GetData()), nil
}

// NewGCPSecretFetcher creates a secret value fetcher for GCP
func NewGCPSecretFetcher(cfg config.GCPConfig, client GCPSecretsIface) SecretFetcher {
	return GCPSecretFetcher{cfg: cfg, client: client}
}
