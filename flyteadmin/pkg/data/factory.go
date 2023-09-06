package data

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/stow"
	"github.com/flyteorg/stow/s3"

	"github.com/flyteorg/flytestdlib/storage"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/flyteorg/flyteadmin/pkg/common"
	"github.com/flyteorg/flyteadmin/pkg/data/implementations"
	"github.com/flyteorg/flyteadmin/pkg/data/interfaces"
)

type RemoteDataHandlerConfig struct {
	CloudProvider            common.CloudProvider
	Retries                  int // Number of times to attempt to initialize a new config on failure.
	Region                   string
	SignedURLDurationMinutes int
	SigningPrincipal         string
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
	case common.GCP:
		signedURLDuration := time.Minute * time.Duration(cfg.SignedURLDurationMinutes)
		return &remoteDataHandler{
			remoteURL: implementations.NewGCPRemoteURL(cfg.SigningPrincipal, signedURLDuration),
		}

	case common.Local:
		logger.Infof(context.TODO(), "setting up local signer ----- ")
		// Since minio = aws s3, we are creating the same client but using the config primitives from aws
		storageCfg := storage.GetConfig()
		accessKeyID := ""
		secret := ""
		endpoint := ""
		if len(storageCfg.Stow.Config) > 0 {
			stowCfg := stow.ConfigMap(storageCfg.Stow.Config)
			accessKeyID, _ = stowCfg.Config(s3.ConfigAccessKeyID)
			secret, _ = stowCfg.Config(s3.ConfigSecretKey)
			endpoint, _ = stowCfg.Config(s3.ConfigEndpoint)
		} else {
			accessKeyID = storageCfg.Connection.AccessKey
			secret = storageCfg.Connection.SecretKey
			endpoint = storageCfg.Connection.Endpoint.String()
		}
		logger.Infof(context.TODO(), "setting up local signer - %s, %s, %s", accessKeyID, secret, endpoint)
		creds := credentials.NewStaticCredentials(accessKeyID, secret, "")
		awsConfig := aws.NewConfig().
			WithRegion(cfg.Region).
			WithMaxRetries(cfg.Retries).
			WithCredentials(creds).
			WithEndpoint(endpoint).
			WithDisableSSL(true).
			WithS3ForcePathStyle(true)
		presignedURLDuration := time.Minute * time.Duration(cfg.SignedURLDurationMinutes)
		return &remoteDataHandler{
			remoteURL: implementations.NewAWSRemoteURL(awsConfig, presignedURLDuration),
		}
	case common.None:
		fallthrough
	default:
		logger.Infof(context.Background(),
			"Using default noop remote url implementation for cloud provider type [%s]", cfg.CloudProvider)
		return &remoteDataHandler{
			remoteURL: implementations.NewNoopRemoteURL(*cfg.RemoteDataStoreClient),
		}
	}
}
