package admin

import (
	"context"

	"golang.org/x/oauth2"
)

// CustomHeaderTokenSource class is here because we cannot use the normal "github.com/grpc/grpc-go/credentials/oauth" package to satisfy
// the credentials.PerRPCCredentials interface. This is because we want to be able to support a different 'header'
// when passing the token in the gRPC call's metadata. The default is filled in in the constructor if none is supplied.
type CustomHeaderTokenSource struct {
	tokenSource  oauth2.TokenSource
	customHeader string
	insecure     bool
}

const DefaultAuthorizationHeader = "authorization"

// RequireTransportSecurity returns whether this credentials class requires TLS/SSL. OAuth uses Bearer tokens that are
// susceptible to MITM (Man-In-The-Middle) attacks that are mitigated by TLS/SSL. We may return false here to make it
// easier to setup auth. However, in a production environment, TLS for OAuth2 is a requirement.
// see also: https://tools.ietf.org/html/rfc6749#section-3.1
func (ts CustomHeaderTokenSource) RequireTransportSecurity() bool {
	return !ts.insecure
}

// GetRequestMetadata gets the authorization metadata as a map using a TokenSource to generate a token
func (ts CustomHeaderTokenSource) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	token, err := ts.tokenSource.Token()
	if err != nil {
		return nil, err
	}

	return map[string]string{
		ts.customHeader: token.Type() + " " + token.AccessToken,
	}, nil
}

func NewCustomHeaderTokenSource(source oauth2.TokenSource, insecure bool, customHeader string) CustomHeaderTokenSource {
	header := DefaultAuthorizationHeader
	if customHeader != "" {
		header = customHeader
	}

	return CustomHeaderTokenSource{
		tokenSource:  source,
		customHeader: header,
		insecure:     insecure,
	}
}
