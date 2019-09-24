package data

import (
	"context"
	"time"

	"github.com/lyft/flytestdlib/logger"

	"github.com/lyft/flytestdlib/storage"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/lyft/flyteadmin/pkg/common"
	"github.com/lyft/flyteadmin/pkg/data/implementations"
	"github.com/lyft/flyteadmin/pkg/data/interfaces"
)

type RemoteDataHandlerConfig struct {
	CloudProvider            common.CloudProvider
	Retries                  int // Number of times to attempt to initialize a new config on failure.
	Region                   string
	SignedURLDurationMinutes int
	RemoteDataStoreClient    *storage.DataStore
}

type RemoteDataHandler interface {
	GetRemoteURLInterface() interfaces.RemoteURLInterface
}

type remoteDataHandler struct {
	remoteURL interfaces.RemoteURLInterface
}

func (r *remoteDataHandler) GetRemoteURLInterface() interfaces.RemoteURLInterface {
	return r.remoteURL
}

func GetRemoteDataHandler(cfg RemoteDataHandlerConfig) RemoteDataHandler {
	switch cfg.CloudProvider {
	case common.AWS:
		awsConfig := aws.NewConfig().WithRegion(cfg.Region).WithMaxRetries(cfg.Retries)
		presignedURLDuration := time.Minute * time.Duration(cfg.SignedURLDurationMinutes)
		return &remoteDataHandler{
			remoteURL: implementations.NewAWSRemoteURL(awsConfig, presignedURLDuration),
		}
	case common.Local:
		fallthrough
	default:
		logger.Infof(context.Background(),
			"Using default noop remote url implementation for cloud provider type [%s]", cfg.CloudProvider)
		return &remoteDataHandler{
			remoteURL: implementations.NewNoopRemoteURL(*cfg.RemoteDataStoreClient),
		}
	}
}
