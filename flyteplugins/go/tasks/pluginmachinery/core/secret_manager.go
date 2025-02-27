package core

import "context"

//go:generate mockery-v2 --all --case=underscore --with-expecter
type SecretManager interface {
	Get(ctx context.Context, key string) (string, error)
}
