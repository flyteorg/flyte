package google

import (
	"context"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type defaultTokenSource struct{}

func (m *defaultTokenSource) GetTokenSource(
	ctx context.Context,
	identity Identity,
) (oauth2.TokenSource, error) {
	return google.DefaultTokenSource(ctx)
}

func NewDefaultTokenSourceFactory() (TokenSourceFactory, error) {
	return &defaultTokenSource{}, nil
}
