/*
 * Copyright (c) 2018 Lyft. All rights reserved.
 */

// Package aws contains AWS-specific logic to handle execution and monitoring of batch jobs.
package aws

import (
	"fmt"
	"os"

	"github.com/flyteorg/flytestdlib/errors"

	"context"

	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/flyteorg/flytestdlib/logger"
)

const (
	EnvSharedCredFilePath                  = "AWS_SHARED_CREDENTIALS_FILE" // #nosec
	EnvAwsProfile                          = "AWS_PROFILE"
	ErrEmptyCredentials   errors.ErrorCode = "EMPTY_CREDS"
	ErrUnknownHost        errors.ErrorCode = "UNKNOWN_HOST"
)

type singleton struct {
	client Client
	lock   sync.RWMutex
}

var single = singleton{
	lock: sync.RWMutex{},
}

// Client is a generic AWS Client that can be used for all AWS Client libraries.
type Client interface {
	GetSession() *session.Session
	GetSdkConfig() *aws.Config
	GetConfig() *Config
	GetHostName() string
}

type client struct {
	config    *Config
	Session   *session.Session
	SdkConfig *aws.Config
	HostName  string
}

// Gets the initialized session.
func (c client) GetSession() *session.Session {
	return c.Session
}

// Gets the final config that was used to initialize AWS Session.
func (c client) GetSdkConfig() *aws.Config {
	return c.SdkConfig
}

// Gets client's Hostname
func (c client) GetHostName() string {
	return c.HostName
}

func (c client) GetConfig() *Config {
	return c.config
}

func newClient(ctx context.Context, cfg *Config) (Client, error) {
	awsConfig := aws.NewConfig().WithRegion(cfg.Region).WithMaxRetries(cfg.Retries)
	if os.Getenv(EnvSharedCredFilePath) != "" {
		creds := credentials.NewSharedCredentials(os.Getenv(EnvSharedCredFilePath), os.Getenv(EnvAwsProfile))
		if creds == nil {
			return nil, fmt.Errorf("unable to Load AWS credentials")
		}

		_, e := creds.Get()
		if e != nil {
			return nil, errors.Wrapf(ErrEmptyCredentials, e, "Empty credentials")
		}

		awsConfig = awsConfig.WithCredentials(creds)
	}

	sess, err := session.NewSession(awsConfig)
	if err != nil {
		logger.Fatalf(ctx, "Error while creating session: %v", err)
	}

	hostname, err := os.Hostname()
	if err != nil {
		return nil, errors.Wrapf(ErrUnknownHost, err, "Unable to discover current hostname")
	}

	return &client{
		config:    cfg,
		SdkConfig: awsConfig,
		Session:   sess,
		HostName:  hostname,
	}, nil
}

// Initializes singleton AWS Client if one hasn't been initialized yet.
func Init(ctx context.Context, cfg *Config) (err error) {
	if single.client == nil {
		single.lock.Lock()
		defer single.lock.Unlock()

		if single.client == nil {
			single.client, err = newClient(ctx, cfg)
		}
	}

	return err
}

// Gets singleton AWS Client.
func GetClient() (c Client, err error) {
	single.lock.RLock()
	defer single.lock.RUnlock()
	if single.client == nil {
		single.client, err = newClient(context.TODO(), GetConfig())
	}

	return single.client, err
}

func SetClient(c Client) {
	single.lock.Lock()
	defer single.lock.Unlock()
	if single.client == nil {
		single.client = c
	}
}
