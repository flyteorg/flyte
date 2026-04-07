package core

import (
	"context"
	"fmt"
	"unicode/utf8"

	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/secret"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/secret/config"
)

type SecretManager interface {
	Get(ctx context.Context, key string) (string, error)
}

type EmbeddedSecretManager struct {
	secretFetcher secret.SecretFetcher
}

func (e *EmbeddedSecretManager) Get(ctx context.Context, key string) (string, error) {
	secretValue, err := e.secretFetcher.GetSecretValue(ctx, key)
	if err != nil {
		return "", err
	}

	if secretValue.StringValue != "" {
		return secretValue.StringValue, nil
	}

	// GCP secrets store values as binary only. We could fail this path for AWS, but for
	// consistent behaviour between AWS and GCP we will allow this path for AWS as well.
	if !utf8.Valid(secretValue.BinaryValue) {
		return "", fmt.Errorf("secret %q has a binary value that is not a valid UTF-8 string", key)
	}
	return string(secretValue.BinaryValue), nil
}

func NewEmbeddedSecretManager(ctx context.Context, cfg config.EmbeddedSecretManagerConfig) (SecretManager, error) {
	secretFetcher, err := secret.NewSecretFetcher(ctx, cfg)
	if err != nil {
		return nil, err
	}

	return &EmbeddedSecretManager{secretFetcher}, nil
}
