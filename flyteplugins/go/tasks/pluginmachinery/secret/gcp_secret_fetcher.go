package secret

import (
	"context"
	"fmt"

	gcpsmpb "cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/secret/config"
	stdlibErrors "github.com/flyteorg/flyte/v2/flytestdlib/errors"
	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
)

const (
	GCPSecretNameFormat = "projects/%s/secrets/%s/versions/latest" // #nosec G101
)

type GCPSecretFetcher struct {
	client GCPSecretManagerClient
	cfg    config.GCPConfig
}

func (g GCPSecretFetcher) GetSecretValue(ctx context.Context, secretID string) (*SecretValue, error) {
	logger.Infof(ctx, "Got fetch secret Request for %v!", secretID)
	resp, err := g.client.AccessSecretVersion(ctx, &gcpsmpb.AccessSecretVersionRequest{
		Name: fmt.Sprintf(GCPSecretNameFormat, g.cfg.Project, secretID),
	})
	if err != nil {
		if s, ok := status.FromError(err); ok && s.Code() == codes.NotFound {
			wrappedErr := stdlibErrors.Wrapf(ErrCodeSecretNotFound, err, fmt.Sprintf(SecretNotFoundErrorFormat, secretID))
			logger.Warn(ctx, wrappedErr)
			return nil, wrappedErr
		}
		wrappedErr := stdlibErrors.Wrapf(ErrCodeSecretReadFailure, err, fmt.Sprintf(SecretReadFailureErrorFormat, secretID))
		logger.Error(ctx, wrappedErr)
		return nil, wrappedErr
	}
	if resp.GetPayload() == nil {
		wrappedErr := stdlibErrors.Wrapf(ErrCodeSecretNil, err, fmt.Sprintf(SecretNilErrorFormat, secretID))
		logger.Error(ctx, wrappedErr)
		return nil, wrappedErr
	}

	return &SecretValue{
		BinaryValue: resp.GetPayload().GetData(),
	}, nil
}

// NewGCPSecretFetcher creates a secret value fetcher for GCP
func NewGCPSecretFetcher(cfg config.GCPConfig, client GCPSecretManagerClient) SecretFetcher {
	return GCPSecretFetcher{cfg: cfg, client: client}
}
