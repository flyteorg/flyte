package storage

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws/awserr"
	awsS3 "github.com/aws/aws-sdk-go/service/s3"

	"github.com/lyft/flytestdlib/logger"
	"github.com/lyft/flytestdlib/promutils"

	"github.com/pkg/errors"

	"github.com/graymeta/stow"
	"github.com/graymeta/stow/azure"
	"github.com/graymeta/stow/google"
	"github.com/graymeta/stow/oracle"
	"github.com/graymeta/stow/s3"
	"github.com/graymeta/stow/swift"
)

var fQNFn = map[string]func(string) DataReference{
	s3.Kind: func(bucket string) DataReference {
		return DataReference(fmt.Sprintf("s3://%s", bucket))
	},
	google.Kind: func(bucket string) DataReference {
		return DataReference(fmt.Sprintf("gs://%s", bucket))
	},
	oracle.Kind: func(bucket string) DataReference {
		return DataReference(fmt.Sprintf("os://%s", bucket))
	},
	swift.Kind: func(bucket string) DataReference {
		return DataReference(fmt.Sprintf("sw://%s", bucket))
	},
	azure.Kind: func(bucket string) DataReference {
		return DataReference(fmt.Sprintf("afs://%s", bucket))
	},
}

func awsBucketIsNotFound(err error) bool {
	if IsNotFound(err) {
		return true
	}

	if awsErr, errOk := errors.Cause(err).(awserr.Error); errOk {
		return awsErr.Code() == awsS3.ErrCodeNoSuchBucket
	}

	return false
}

func awsBucketAlreadyExists(err error) bool {
	if IsExists(err) {
		return true
	}

	if awsErr, errOk := errors.Cause(err).(awserr.Error); errOk {
		return awsErr.Code() == awsS3.ErrCodeBucketAlreadyOwnedByYou
	}

	return false
}

func newStowRawStore(cfg *Config, metricsScope promutils.Scope) (RawStore, error) {
	if cfg.InitContainer == "" {
		return nil, fmt.Errorf("initContainer is required")
	}

	var cfgMap stow.ConfigMap
	var kind string
	if cfg.Stow != nil {
		kind = cfg.Stow.Kind
		cfgMap = stow.ConfigMap(cfg.Stow.Config)
	} else {
		logger.Warnf(context.TODO(), "stow configuration section missing, defaulting to legacy s3/minio connection config")
		// This is for supporting legacy configurations which configure S3 via connection config
		kind = s3.Kind
		cfgMap = legacyS3ConfigMap(cfg.Connection)
	}

	fn, ok := fQNFn[kind]
	if !ok {
		return nil, errors.Errorf("unsupported stow.kind [%s], add support in flytestdlib?", kind)
	}

	loc, err := stow.Dial(kind, cfgMap)
	if err != nil {
		return emptyStore, fmt.Errorf("unable to configure the storage for %s. Error: %v", kind, err)
	}

	c, err := loc.Container(cfg.InitContainer)
	if err != nil {
		if IsNotFound(err) || awsBucketIsNotFound(err) {
			c, err := loc.CreateContainer(cfg.InitContainer)
			// If the container's already created, move on. Otherwise, fail with error.
			if err != nil && !awsBucketAlreadyExists(err) {
				return emptyStore, fmt.Errorf("unable to initialize container [%v]. Error: %v", cfg.InitContainer, err)
			}
			return NewStowRawStore(fn(c.Name()), c, metricsScope)
		}

		return emptyStore, err
	}

	return NewStowRawStore(fn(c.Name()), c, metricsScope)
}

func legacyS3ConfigMap(cfg ConnectionConfig) stow.ConfigMap {
	// Non-nullable fields
	stowConfig := stow.ConfigMap{
		s3.ConfigAuthType: cfg.AuthType,
		s3.ConfigRegion:   cfg.Region,
	}

	// Fields that differ between minio and real S3
	if endpoint := cfg.Endpoint.String(); endpoint != "" {
		stowConfig[s3.ConfigEndpoint] = endpoint
	}

	if accessKey := cfg.AccessKey; accessKey != "" {
		stowConfig[s3.ConfigAccessKeyID] = accessKey
	}

	if secretKey := cfg.SecretKey; secretKey != "" {
		stowConfig[s3.ConfigSecretKey] = secretKey
	}

	if disableSsl := cfg.DisableSSL; disableSsl {
		stowConfig[s3.ConfigDisableSSL] = "True"
	}

	return stowConfig
}
