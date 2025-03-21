package core

import "context"

//go:generate mockery --all --case=underscore --with-expecter
type SecretManager interface {
	Get(ctx context.Context, key string) (string, error)
}
