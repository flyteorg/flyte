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
	if config.Type == TokenSourceTypeDefault {
		return NewDefaultTokenSourceFactory()
	}

	return nil, errors.Errorf("unknown token source type [%v], possible values are: 'default'", config.Type)
}
