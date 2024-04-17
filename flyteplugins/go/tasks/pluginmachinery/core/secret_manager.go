package core

import (
	"context"
)

//go:generate mockery -all -output=./mocks -case=underscore

type SecretManager interface {
	Get(ctx context.Context, key string) (string, error)
}
