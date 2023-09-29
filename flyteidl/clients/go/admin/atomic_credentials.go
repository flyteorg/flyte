package admin

import (
	"context"
	"sync/atomic"

	"google.golang.org/grpc/credentials"

	stdlibAtomic "github.com/flyteorg/flytestdlib/atomic"
)

// atomicPerRPCCredentials provides a convenience on top of atomic.Value and credentials.PerRPCCredentials to be thread-safe.
type atomicPerRPCCredentials struct {
	atomic.Value
}

func (t *atomicPerRPCCredentials) Store(properties credentials.PerRPCCredentials) {
	t.Value.Store(properties)
}

func (t *atomicPerRPCCredentials) Load() credentials.PerRPCCredentials {
	val := t.Value.Load()
	if val == nil {
		return CustomHeaderTokenSource{}
	}

	return val.(credentials.PerRPCCredentials)
}

func newAtomicPerPRCCredentials() *atomicPerRPCCredentials {
	return &atomicPerRPCCredentials{
		Value: atomic.Value{},
	}
}

// PerRPCCredentialsFuture is a future wrapper for credentials.PerRPCCredentials that can act as one and also be
// materialized later.
type PerRPCCredentialsFuture struct {
	perRPCCredentials *atomicPerRPCCredentials
	initialized       stdlibAtomic.Bool
}

// GetRequestMetadata gets the authorization metadata as a map using a TokenSource to generate a token
func (ts *PerRPCCredentialsFuture) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	if ts.initialized.Load() {
		tp := ts.perRPCCredentials.Load()
		return tp.GetRequestMetadata(ctx, uri...)
	}

	return map[string]string{}, nil
}

// RequireTransportSecurity returns whether this credentials class requires TLS/SSL. OAuth uses Bearer tokens that are
// susceptible to MITM (Man-In-The-Middle) attacks that are mitigated by TLS/SSL. We may return false here to make it
// easier to setup auth. However, in a production environment, TLS for OAuth2 is a requirement.
// see also: https://tools.ietf.org/html/rfc6749#section-3.1
func (ts *PerRPCCredentialsFuture) RequireTransportSecurity() bool {
	if ts.initialized.Load() {
		return ts.perRPCCredentials.Load().RequireTransportSecurity()
	}

	return false
}

func (ts *PerRPCCredentialsFuture) Store(tokenSource credentials.PerRPCCredentials) {
	ts.perRPCCredentials.Store(tokenSource)
	ts.initialized.Store(true)
}

func (ts *PerRPCCredentialsFuture) Get() credentials.PerRPCCredentials {
	return ts.perRPCCredentials.Load()
}

func (ts *PerRPCCredentialsFuture) IsInitialized() bool {
	return ts.initialized.Load()
}

// NewPerRPCCredentialsFuture initializes a new PerRPCCredentialsFuture that can act as a credentials.PerRPCCredentials
// and can also be resolved in the future. Users of the future can check if it has been initialized before by calling
// PerRPCCredentialsFuture.IsInitialized(). Calling PerRPCCredentialsFuture.Get() multiple times will return
// the same stored object (unless it changed in between calls). Calling PerRPCCredentialsFuture.Store() multiple
// times is supported and will result in overriding the old value atomically.
func NewPerRPCCredentialsFuture() *PerRPCCredentialsFuture {
	tokenSource := PerRPCCredentialsFuture{
		perRPCCredentials: newAtomicPerPRCCredentials(),
	}

	return &tokenSource
}
