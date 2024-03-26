package google

import (
	"context"

	"github.com/pkg/errors"
	"golang.org/x/oauth2"
)

type Identity struct {
	K8sNamespace      string
	K8sServiceAccount string
}

type TokenSourceFactory interface {
	GetTokenSource(ctx context.Context, identity Identity) (oauth2.TokenSource, error)
}

func NewTokenSourceFactory(config TokenSourceFactoryConfig) (TokenSourceFactory, error) {
	switch config.Type {
	case TokenSourceTypeDefault:
		return NewDefaultTokenSourceFactory()
	case TokenSourceTypeGkeTaskWorkloadIdentity:
		return NewGkeTaskWorkloadIdentityTokenSourceFactory(
			&config.GkeTaskWorkloadIdentityTokenSourceFactoryConfig,
		)
	}

	return nil, errors.Errorf(
		"unknown token source type [%v], possible values are: 'default' and 'gke-task-workload-identity'",
		config.Type,
	)
}
