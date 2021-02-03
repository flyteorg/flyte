package admin

import (
	"context"

	"golang.org/x/oauth2"
)

// This class is here because we cannot use the normal "github.com/grpc/grpc-go/credentials/oauth" package to satisfy
// the credentials.PerRPCCredentials interface. This is because we want to be able to support a different 'header'
// when passing the token in the gRPC call's metadata. The default is filled in in the constructor if none is supplied.
type CustomHeaderTokenSource struct {
	oauth2.TokenSource
	customHeader string
}

const DefaultAuthorizationHeader = "authorization"

// GetRequestMetadata gets the authorization metadata as a map using a TokenSource to generate a token
func (ts CustomHeaderTokenSource) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	token, err := ts.Token()
	if err != nil {
		return nil, err
	}
	return map[string]string{
		ts.customHeader: token.Type() + " " + token.AccessToken,
	}, nil
}

// Even though Admin is capable of serving authentication without SSL, we're going to require it here. That is, this module's
// canonical Admin client will only do auth over SSL.
func (ts CustomHeaderTokenSource) RequireTransportSecurity() bool {
	return true
}

func NewCustomHeaderTokenSource(source oauth2.TokenSource, customHeader string) CustomHeaderTokenSource {
	header := DefaultAuthorizationHeader
	if customHeader != "" {
		header = customHeader
	}
	return CustomHeaderTokenSource{
		TokenSource:  source,
		customHeader: header,
	}
}
