package service

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	k8sRuntime "k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

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
	s := NewSecretService(k)
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
	s := NewSecretService(k)
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
	s := NewSecretService(k)
	ctx := context.Background()

	// Write one secret at each scope.
	secrets := []*secretpb.SecretIdentifier{
		{Name: "g"},                                           // global
		{Domain: "d", Name: "dd"},                             // domain
		{Domain: "d", Project: "p", Name: "pp"},               // project+domain
	}
	for _, id := range secrets {
		_, err := s.CreateSecret(ctx, connect.NewRequest(&secretpb.CreateSecretRequest{
			Id: id,
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
	s := NewSecretService(k)
	_, err := s.ListSecrets(context.Background(), connect.NewRequest(&secretpb.ListSecretsRequest{Project: "flytesnacks"}))
	require.Error(t, err)
}
