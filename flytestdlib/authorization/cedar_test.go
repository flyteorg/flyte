package authorization

import (
	"context"
	"encoding/base64"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"connectrpc.com/connect"
	cedar "github.com/cedar-policy/cedar-go"
	"github.com/stretchr/testify/require"
)

func testConfig(t *testing.T, policy string) Config {
	t.Helper()
	path := filepath.Join(t.TempDir(), "policy.cedar")
	require.NoError(t, os.WriteFile(path, []byte(policy), 0600))
	return Config{Enabled: true, TrustForwardedClaims: true, ClaimsHeader: "Authorization", Issuer: "https://issuer", PolicyFile: path}
}

func testHeader(claims string) http.Header {
	h := http.Header{}
	h.Set("Authorization", "Bearer e30."+base64.RawURLEncoding.EncodeToString([]byte(claims))+".signature")
	return h
}

func testResolver(string, any) ([]Requirement, error) {
	id := cedar.NewEntityUID("Flyte::Scope", "payments/prod")
	return []Requirement{{Action: "RunRead", Resource: id, Entities: cedar.EntityMap{id: {UID: id}}}}, nil
}

func TestCedarDecisions(t *testing.T) {
	for _, tc := range []struct {
		name, policy string
		allowed      bool
	}{
		{"default deny", "", false},
		{"group permit", `permit(principal in Flyte::Group::"engineers", action == Flyte::Action::"RunRead", resource == Flyte::Scope::"payments/prod");`, true},
		{"forbid wins", `permit(principal, action, resource); forbid(principal, action, resource);`, false},
		{"evaluation error fails closed", `permit(principal, action, resource); forbid(principal, action, resource) when { principal.missing };`, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			i, err := New(testConfig(t, tc.policy), testResolver)
			require.NoError(t, err)
			c, err := i.authenticate(testHeader(`{"sub":"alice","iss":"https://issuer","groups":["engineers"]}`))
			require.NoError(t, err)
			err = i.authorize(context.Background(), c, "test", nil)
			if tc.allowed {
				require.NoError(t, err)
			} else {
				require.Equal(t, connect.CodePermissionDenied, connect.CodeOf(err))
			}
		})
	}
}

func TestTrustedIdentity(t *testing.T) {
	i, err := New(testConfig(t, ""), testResolver)
	require.NoError(t, err)
	for _, claims := range []string{
		`{}`, `{"sub":" ","iss":"https://issuer"}`, `{"sub":"alice","iss":"attacker"}`,
		`{"sub":"alice","iss":"https://issuer","groups":"admins"}`, `{"sub":"alice","iss":"https://issuer","groups":[""]}`, `invalid`,
	} {
		_, err := i.authenticate(testHeader(claims))
		require.Equal(t, connect.CodeUnauthenticated, connect.CodeOf(err), claims)
	}
	for _, header := range []http.Header{nil, {"Authorization": {"garbage"}}, {"Authorization": {"one", "two"}}} {
		_, err := i.authenticate(header)
		require.Equal(t, connect.CodeUnauthenticated, connect.CodeOf(err))
	}
	i.cfg.ClaimsHeader = "X-Verified-Claims"
	h := testHeader(`{"sub":"alice","iss":"https://issuer"}`)
	_, err = i.authenticate(h) // no fallback to another identity source
	require.Equal(t, connect.CodeUnauthenticated, connect.CodeOf(err))
}

func TestConfiguration(t *testing.T) {
	_, err := New(testConfig(t, "not cedar"), testResolver)
	require.Error(t, err)
	for _, change := range []func(*Config){
		func(c *Config) { c.TrustForwardedClaims = false },
		func(c *Config) { c.Issuer = "" },
		func(c *Config) { c.PolicyFile = "missing" },
		func(c *Config) { c.ClaimsHeader = "bad\nheader" },
	} {
		cfg := testConfig(t, "")
		change(&cfg)
		_, err := New(cfg, testResolver)
		require.Error(t, err)
	}
}

func TestConcurrentDecisions(t *testing.T) {
	i, err := New(testConfig(t, `permit(principal == Flyte::Principal::"alice", action, resource);`), testResolver)
	require.NoError(t, err)
	c, err := i.authenticate(testHeader(`{"sub":"alice","iss":"https://issuer"}`))
	require.NoError(t, err)
	var wg sync.WaitGroup
	errors := make(chan error, 32)
	for n := 0; n < cap(errors); n++ {
		wg.Add(1)
		go func() { defer wg.Done(); errors <- i.authorize(context.Background(), c, "test", nil) }()
	}
	wg.Wait()
	close(errors)
	for err := range errors {
		require.NoError(t, err)
	}
}
