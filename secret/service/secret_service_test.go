package service

import (
	"context"
	"testing"

	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync/atomic"

	"connectrpc.com/connect"
	"github.com/flyteorg/flyte/v2/flytestdlib/serviceclient"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	k8sRuntime "k8s.io/apimachinery/pkg/runtime"
	k8sFake "k8s.io/client-go/kubernetes/fake"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/flyteorg/flyte/v2/executor/pkg/webhook"
	flytesecret "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/secret"
	secretpb "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/secret"
)

func newTestClient(t *testing.T) client.Client {
	t.Helper()
	scheme := k8sRuntime.NewScheme()
	require.NoError(t, corev1.AddToScheme(scheme))
	return fake.NewClientBuilder().WithScheme(scheme).Build()
}

func TestValidateScope(t *testing.T) {
	cases := []struct {
		name    string
		domain  string
		project string
		wantErr bool
	}{
		{name: "global", wantErr: false},
		{name: "domain only", domain: "development", wantErr: false},
		{name: "project+domain", domain: "development", project: "flytesnacks", wantErr: false},
		{name: "project without domain", project: "flytesnacks", wantErr: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateScope(tc.domain, tc.project)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetK8sSecretName_EncodingMatchesFetcher(t *testing.T) {
	// The k8s resource name written by the service must be identical to what
	// K8sSecretFetcher computes from pod labels, otherwise the fetcher cannot
	// locate the Secret at pod admission.
	id := &secretpb.SecretIdentifier{
		Project: "flytesnacks",
		Domain:  "development",
		Name:    "my-secret",
	}
	encoded, k8sName, err := getK8sSecretName(context.Background(), id)
	require.NoError(t, err)

	expectedEncoded := flytesecret.EncodeSecretName(defaultOrganization, id.GetDomain(), id.GetProject(), id.GetName())
	assert.Equal(t, expectedEncoded, encoded)
	assert.Equal(t, flytesecret.EncodeK8sSecretName(expectedEncoded), k8sName)
}

func TestGetK8sSecretName_ScopeHashesDiffer(t *testing.T) {
	// Same secret name at different scopes must hash to different k8s resource
	// names so they can coexist in a single namespace without collision.
	names := map[string]bool{}
	for _, id := range []*secretpb.SecretIdentifier{
		{Name: "s"},
		{Domain: "d", Name: "s"},
		{Domain: "d", Project: "p", Name: "s"},
	} {
		_, k8sName, err := getK8sSecretName(context.Background(), id)
		require.NoError(t, err)
		assert.False(t, names[k8sName], "scope hashes collided for %+v", id)
		names[k8sName] = true
	}
}

func TestSecretService_CreateThenGet(t *testing.T) {
	k := newTestClient(t)
	s := NewSecretService(k, "")
	ctx := context.Background()
	id := &secretpb.SecretIdentifier{
		Project: "flytesnacks",
		Domain:  "development",
		Name:    "my-secret",
	}

	_, err := s.CreateSecret(ctx, connect.NewRequest(&secretpb.CreateSecretRequest{
		Id: id,
		SecretSpec: &secretpb.SecretSpec{
			Value: &secretpb.SecretSpec_BinaryValue{BinaryValue: []byte("shhh")},
		},
	}))
	require.NoError(t, err)

	resp, err := s.GetSecret(ctx, connect.NewRequest(&secretpb.GetSecretRequest{Id: id}))
	require.NoError(t, err)
	assert.Equal(t, id.Name, resp.Msg.GetSecret().GetId().GetName())
}

func TestSecretService_GetNotFound_UsesOriginalName(t *testing.T) {
	k := newTestClient(t)
	s := NewSecretService(k, "")
	ctx := context.Background()
	_, err := s.GetSecret(ctx, connect.NewRequest(&secretpb.GetSecretRequest{
		Id: &secretpb.SecretIdentifier{Name: "my-secret1"},
	}))
	require.Error(t, err)
	// Error surfaces the user-facing name, not the internal encoded form.
	assert.Contains(t, err.Error(), "my-secret1")
	assert.NotContains(t, err.Error(), "u__org__")
}

func TestSecretService_List_FiltersByScope(t *testing.T) {
	k := newTestClient(t)
	s := NewSecretService(k, "")
	ctx := context.Background()

	// Write one secret at each scope.
	secrets := []*secretpb.SecretIdentifier{
		{Name: "g"},                             // global
		{Domain: "d", Name: "dd"},               // domain
		{Domain: "d", Project: "p", Name: "pp"}, // project+domain
	}
	for _, id := range secrets {
		_, err := s.CreateSecret(ctx, connect.NewRequest(&secretpb.CreateSecretRequest{
			Id:         id,
			SecretSpec: &secretpb.SecretSpec{Value: &secretpb.SecretSpec_BinaryValue{BinaryValue: []byte("v")}},
		}))
		require.NoError(t, err)
	}

	// Empty request returns everything.
	resp, err := s.ListSecrets(ctx, connect.NewRequest(&secretpb.ListSecretsRequest{}))
	require.NoError(t, err)
	assert.Len(t, resp.Msg.GetSecrets(), 3)

	// Domain-only request returns every secret that carries that domain
	// label (domain-scoped and project+domain-scoped both qualify).
	resp, err = s.ListSecrets(ctx, connect.NewRequest(&secretpb.ListSecretsRequest{Domain: "d"}))
	require.NoError(t, err)
	names := namesOf(resp.Msg.GetSecrets())
	assert.ElementsMatch(t, []string{"dd", "pp"}, names)

	// Project+domain request returns only that scope.
	resp, err = s.ListSecrets(ctx, connect.NewRequest(&secretpb.ListSecretsRequest{Domain: "d", Project: "p"}))
	require.NoError(t, err)
	require.Len(t, resp.Msg.GetSecrets(), 1)
	assert.Equal(t, "pp", resp.Msg.GetSecrets()[0].GetId().GetName())
}

func namesOf(secrets []*secretpb.Secret) []string {
	out := make([]string, 0, len(secrets))
	for _, s := range secrets {
		out = append(out, s.GetId().GetName())
	}
	return out
}

func TestSecretService_List_RejectsProjectWithoutDomain(t *testing.T) {
	k := newTestClient(t)
	s := NewSecretService(k, "")
	_, err := s.ListSecrets(context.Background(), connect.NewRequest(&secretpb.ListSecretsRequest{Project: "flytesnacks"}))
	require.Error(t, err)
}

func TestK8sSecretWrittenByServiceIsReadableByWebhookFetcher(t *testing.T) {
	// End-to-end naming compatibility: a Secret written by CreateSecret must be
	// found by the K8sSecretFetcher the pod webhook's embedded secret manager
	// uses, via the scope-fallback lookup IDs derived from task pod labels.
	// This is the invariant that, when broken, silently delivers no secrets.
	podLabels := map[string]string{
		flytesecret.OrganizationLabel: defaultOrganization,
		flytesecret.DomainLabel:       "development",
		flytesecret.ProjectLabel:      "flytesnacks",
	}

	for _, tc := range []struct {
		scope string
		id    *secretpb.SecretIdentifier
	}{
		{"project", &secretpb.SecretIdentifier{Domain: "development", Project: "flytesnacks", Name: "sec"}},
		{"domain", &secretpb.SecretIdentifier{Domain: "development", Name: "sec"}},
		{"org", &secretpb.SecretIdentifier{Name: "sec"}},
	} {
		t.Run(tc.scope, func(t *testing.T) {
			encoded, k8sName, err := getK8sSecretName(context.Background(), tc.id)
			require.NoError(t, err)
			ns := secretNamespace()
			k8sSecret, err := buildK8sSecret(encoded, k8sName, ns, tc.id,
				&secretpb.SecretSpec{Value: &secretpb.SecretSpec_StringValue{StringValue: "v"}})
			require.NoError(t, err)
			// Fake typed clientset does not merge StringData into Data; mimic apiserver behavior.
			k8sSecret.Data = map[string][]byte{encoded: []byte("v")}
			k8sSecret.StringData = nil

			fetcher := flytesecret.NewK8sSecretFetcher(
				k8sFake.NewSimpleClientset(k8sSecret).CoreV1().Secrets(ns))
			// Walk scopes exactly like EmbeddedSecretManagerInjector.lookUpSecret.
			ids := []string{
				flytesecret.EncodeSecretName(podLabels["organization"], podLabels["domain"], podLabels["project"], "sec"),
				flytesecret.EncodeSecretName(podLabels["organization"], podLabels["domain"], "", "sec"),
				flytesecret.EncodeSecretName(podLabels["organization"], "", "", "sec"),
			}
			expectedID := flytesecret.EncodeSecretName(defaultOrganization, tc.id.GetDomain(), tc.id.GetProject(), tc.id.GetName())
			var foundID string
			for _, id := range ids {
				v, err := fetcher.GetSecretValue(context.Background(), id)
				if err == nil {
					assert.Equal(t, "v", string(v.BinaryValue))
					foundID = id
					break
				}
			}
			assert.Equal(t, expectedID, foundID, "secret created at %s scope was not found at the expected lookup scope", tc.scope)
		})
	}
}

// Every write path must invalidate the webhook's secret cache, otherwise a task admitted after
// the write keeps getting the old value until the cache TTL expires. Create counts as a write:
// a newly created narrow-scope secret can be shadowed by an already-cached broader-scope one.
func TestSecretService_WritesInvalidateCache(t *testing.T) {
	ctx := context.Background()
	var got []webhook.InvalidateRequest

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, webhook.InvalidateSecretPath, r.URL.Path)
		var req webhook.InvalidateRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
		got = append(got, req)
	}))
	defer srv.Close()

	id := &secretpb.SecretIdentifier{Project: "flytesnacks", Domain: "development", Name: "my-secret"}
	spec := &secretpb.SecretSpec{Value: &secretpb.SecretSpec_BinaryValue{BinaryValue: []byte("v")}}
	s := NewSecretService(newTestClient(t), srv.URL)

	_, err := s.CreateSecret(ctx, connect.NewRequest(&secretpb.CreateSecretRequest{Id: id, SecretSpec: spec}))
	require.NoError(t, err)
	_, err = s.UpdateSecret(ctx, connect.NewRequest(&secretpb.UpdateSecretRequest{Id: id, SecretSpec: spec}))
	require.NoError(t, err)
	_, err = s.DeleteSecret(ctx, connect.NewRequest(&secretpb.DeleteSecretRequest{Id: id}))
	require.NoError(t, err)

	want := webhook.InvalidateRequest{Org: defaultOrganization, Domain: "development", Project: "flytesnacks", Name: "my-secret"}
	assert.Equal(t, []webhook.InvalidateRequest{want, want, want}, got)
}

func TestSecretService_AuthenticatedCacheInvalidation(t *testing.T) {
	ctx := context.Background()
	var tokenRequests atomic.Int32
	mux := http.NewServeMux()
	srv := httptest.NewServer(mux)
	defer srv.Close()

	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		tokenRequests.Add(1)
		clientID, clientSecret, ok := r.BasicAuth()
		if !ok {
			require.NoError(t, r.ParseForm())
			clientID, clientSecret = r.Form.Get("client_id"), r.Form.Get("client_secret")
		}
		assert.Equal(t, "secret-service", clientID)
		assert.Equal(t, "secret", clientSecret)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"access_token":"webhook-token","token_type":"Bearer","expires_in":3600}`)
	})
	mux.HandleFunc("/.well-known/oauth-authorization-server", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintf(w, `{"token_endpoint":%q}`, srv.URL+"/token")
	})
	mux.HandleFunc(webhook.InvalidateSecretPath, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer webhook-token", r.Header.Get("Authorization"))
		w.WriteHeader(http.StatusNoContent)
	})

	s, err := NewSecretServiceWithConfig(ctx, newTestClient(t), serviceclient.ServiceConfig{
		URL: srv.URL,
		Auth: serviceclient.AuthConfig{
			Type:         serviceclient.AuthTypeOAuth2ClientCredentials,
			IssuerURL:    srv.URL,
			ClientID:     "secret-service",
			ClientSecret: "secret",
		},
	})
	require.NoError(t, err)

	_, err = s.CreateSecret(ctx, connect.NewRequest(&secretpb.CreateSecretRequest{
		Id:         &secretpb.SecretIdentifier{Project: "flytesnacks", Domain: "development", Name: "my-secret"},
		SecretSpec: &secretpb.SecretSpec{Value: &secretpb.SecretSpec_BinaryValue{BinaryValue: []byte("v")}},
	}))
	require.NoError(t, err)
	assert.Equal(t, int32(1), tokenRequests.Load())
}

// Invalidation is best-effort: the secret is already durably written when it runs, so an
// unreachable or erroring webhook must not fail the RPC. It only means the old value is served
// until the webhook cache TTL expires.
func TestSecretService_InvalidationFailureDoesNotFailWrite(t *testing.T) {
	spec := &secretpb.SecretSpec{Value: &secretpb.SecretSpec_BinaryValue{BinaryValue: []byte("v")}}

	failing := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer failing.Close()

	for name, url := range map[string]string{
		"disabled":     "",
		"unreachable":  "http://127.0.0.1:1",
		"server error": failing.URL,
	} {
		t.Run(name, func(t *testing.T) {
			s := NewSecretService(newTestClient(t), url)
			_, err := s.CreateSecret(context.Background(), connect.NewRequest(&secretpb.CreateSecretRequest{
				Id:         &secretpb.SecretIdentifier{Domain: "development", Name: "s"},
				SecretSpec: spec,
			}))
			require.NoError(t, err)
		})
	}
}

// A delete whose Secret is already gone (deleted out-of-band) still has to invalidate: the
// webhook may be serving the old value from cache, so skipping it would keep a secret the user
// believes is deleted injected into pods until the TTL expires.
func TestSecretService_DeleteInvalidatesEvenWhenNotFound(t *testing.T) {
	var got []webhook.InvalidateRequest
	srv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		var req webhook.InvalidateRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
		got = append(got, req)
	}))
	defer srv.Close()

	s := NewSecretService(newTestClient(t), srv.URL)
	_, err := s.DeleteSecret(context.Background(), connect.NewRequest(&secretpb.DeleteSecretRequest{
		Id: &secretpb.SecretIdentifier{Domain: "development", Name: "never-created"},
	}))

	require.Error(t, err, "deleting a missing secret should still report NotFound")
	assert.Equal(t, connect.CodeNotFound, connect.CodeOf(err))
	assert.Equal(t, []webhook.InvalidateRequest{
		{Org: defaultOrganization, Domain: "development", Name: "never-created"},
	}, got)
}

// A headless Service resolves to one A record per webhook pod, and each replica caches
// independently — so invalidation has to reach every one of them. Posting to the name and
// letting the HTTP client pick an address would silently invalidate one pod and leave the rest
// serving stale secrets. localhost is the stand-in here: it resolves to 127.0.0.1 and, on a
// dual-stack host, ::1 as well.
func TestSecretService_InvalidationFansOutToEveryResolvedAddress(t *testing.T) {
	ips, err := net.DefaultResolver.LookupHost(context.Background(), "localhost")
	require.NoError(t, err)

	// One listener per resolved address, all on the same port, so a fan-out hits each exactly
	// once and a single-target implementation hits exactly one.
	port, listeners, hits := newPerAddressListeners(t, ips)
	defer func() {
		for _, l := range listeners {
			_ = l.Close()
		}
	}()

	s := NewSecretService(newTestClient(t), fmt.Sprintf("http://localhost:%s", port))
	_, err = s.CreateSecret(context.Background(), connect.NewRequest(&secretpb.CreateSecretRequest{
		Id:         &secretpb.SecretIdentifier{Domain: "development", Name: "fanout"},
		SecretSpec: &secretpb.SecretSpec{Value: &secretpb.SecretSpec_BinaryValue{BinaryValue: []byte("v")}},
	}))
	require.NoError(t, err)

	for _, ip := range ips {
		assert.Equal(t, int32(1), hits[ip].Load(), "address %s should have been invalidated exactly once", ip)
	}
}

// newPerAddressListeners binds one HTTP server per address on a shared free port and returns
// that port plus a per-address hit counter.
func newPerAddressListeners(t *testing.T, ips []string) (string, []net.Listener, map[string]*atomic.Int32) {
	t.Helper()

	probe, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	port := strconv.Itoa(probe.Addr().(*net.TCPAddr).Port)
	require.NoError(t, probe.Close())

	hits := make(map[string]*atomic.Int32, len(ips))
	listeners := make([]net.Listener, 0, len(ips))
	for _, ip := range ips {
		counter := &atomic.Int32{}
		hits[ip] = counter

		l, err := net.Listen("tcp", net.JoinHostPort(ip, port))
		require.NoError(t, err, "binding %s", ip)
		listeners = append(listeners, l)

		srv := &http.Server{Handler: http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
			assert.Equal(t, webhook.InvalidateSecretPath, r.URL.Path)
			counter.Add(1)
		})}
		go func() { _ = srv.Serve(l) }()
	}
	return port, listeners, hits
}

// Target construction is host-independent: an IP literal passes through the resolver untouched,
// so these lock in the port defaulting and scheme/path assembly without depending on DNS.
func TestSecretService_InvalidationTargets(t *testing.T) {
	for name, tc := range map[string]struct {
		url  string
		want []string
	}{
		"explicit port":  {"http://127.0.0.1:9999", []string{"http://127.0.0.1:9999/invalidate-secret"}},
		"default port":   {"http://127.0.0.1", []string{"http://127.0.0.1:9444/invalidate-secret"}},
		"trailing slash": {"http://127.0.0.1:9999/", []string{"http://127.0.0.1:9999/invalidate-secret"}},
		"ipv6 literal":   {"http://[::1]:9999", []string{"http://[::1]:9999/invalidate-secret"}},

		// Scheme-less forms: url.Parse reads "127.0.0.1:9999" as scheme "127.0.0.1" with an
		// empty Host, so these would be rejected and invalidation would degrade to TTL.
		"no scheme, with port": {"127.0.0.1:9999", []string{"http://127.0.0.1:9999/invalidate-secret"}},
		"no scheme, no port":   {"127.0.0.1", []string{"http://127.0.0.1:9444/invalidate-secret"}},
	} {
		t.Run(name, func(t *testing.T) {
			got, err := NewSecretService(nil, tc.url).invalidationTargets(context.Background())
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}

	t.Run("rejects a url with no host", func(t *testing.T) {
		_, err := NewSecretService(nil, "http://").invalidationTargets(context.Background())
		assert.Error(t, err)
	})
}

// A client that hangs up after the write is persisted must not leave the webhook serving the old
// value: by then the secret is already durably changed, and the cancelled request context would
// otherwise propagate into DNS resolution and the POST, failing invalidation for up to the cache
// TTL. Fails without context.WithoutCancel in invalidateWebhookSecretCache.
func TestSecretService_InvalidatesWithAlreadyCancelledContext(t *testing.T) {
	got := make(chan webhook.InvalidateRequest, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		var req webhook.InvalidateRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
		got <- req
	}))
	defer srv.Close()

	// Cancelled before the call, standing in for a caller that hung up once the write landed.
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	NewSecretService(newTestClient(t), srv.URL).invalidateWebhookSecretCache(ctx,
		&secretpb.SecretIdentifier{Domain: "development", Project: "flytesnacks", Name: "hungup"})

	select {
	case req := <-got:
		assert.Equal(t, webhook.InvalidateRequest{
			Org: defaultOrganization, Domain: "development", Project: "flytesnacks", Name: "hungup",
		}, req)
	default:
		t.Fatal("invalidation was not sent: a cancelled caller context suppressed it")
	}
}
