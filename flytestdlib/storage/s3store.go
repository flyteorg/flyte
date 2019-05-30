package storage

import (
	"context"
	"fmt"

	"github.com/lyft/flytestdlib/promutils"

	"github.com/aws/aws-sdk-go/aws/awserr"
	awsS3 "github.com/aws/aws-sdk-go/service/s3"
	"github.com/lyft/flytestdlib/logger"
	"github.com/pkg/errors"

	"github.com/graymeta/stow"
	"github.com/graymeta/stow/s3"
)

func getStowConfigMap(cfg *Config) stow.ConfigMap {
	// Non-nullable fields
	stowConfig := stow.ConfigMap{
		s3.ConfigAuthType: cfg.Connection.AuthType,
		s3.ConfigRegion:   cfg.Connection.Region,
	}

	// Fields that differ between minio and real S3
	if endpoint := cfg.Connection.Endpoint.String(); endpoint != "" {
		stowConfig[s3.ConfigEndpoint] = endpoint
	}

	if accessKey := cfg.Connection.AccessKey; accessKey != "" {
		stowConfig[s3.ConfigAccessKeyID] = accessKey
	}

	if secretKey := cfg.Connection.SecretKey; secretKey != "" {
		stowConfig[s3.ConfigSecretKey] = secretKey
	}

	if disableSsl := cfg.Connection.DisableSSL; disableSsl {
		stowConfig[s3.ConfigDisableSSL] = "True"
	}

	return stowConfig

}

func s3FQN(bucket string) DataReference {
	return DataReference(fmt.Sprintf("s3://%s", bucket))
}

func newS3RawStore(cfg *Config, metricsScope promutils.Scope) (RawStore, error) {
	if cfg.InitContainer == "" {
		return nil, fmt.Errorf("initContainer is required")
	}

	loc, err := stow.Dial(s3.Kind, getStowConfigMap(cfg))

	if err != nil {
		return emptyStore, fmt.Errorf("unable to configure the storage for s3. Error: %v", err)
	}

	c, err := loc.Container(cfg.InitContainer)
	if err != nil {
		if IsNotFound(err) || awsBucketIsNotFound(err) {
			c, err := loc.CreateContainer(cfg.InitContainer)
			if err != nil {
				// If the container's already created, move on. Otherwise, fail with error.
				if awsBucketAlreadyExists(err) {
					logger.Infof(context.TODO(), "Storage init-container already exists [%v].", cfg.InitContainer)
					return NewStowRawStore(s3FQN(c.Name()), c, metricsScope)
				}
				return emptyStore, fmt.Errorf("unable to initialize container [%v]. Error: %v", cfg.InitContainer, err)
			}
			return NewStowRawStore(s3FQN(c.Name()), c, metricsScope)
		}
		return emptyStore, err
	}

	return NewStowRawStore(s3FQN(c.Name()), c, metricsScope)
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
