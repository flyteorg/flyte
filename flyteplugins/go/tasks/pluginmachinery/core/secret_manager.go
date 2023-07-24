package core

import "context"

type SecretManager interface {
	Get(ctx context.Context, key string) (string, error)
}
