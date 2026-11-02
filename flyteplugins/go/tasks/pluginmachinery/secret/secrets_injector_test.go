package secret

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/secret/config"
	"github.com/flyteorg/flyte/v2/flytestdlib/promutils"
)

func TestSecretInformerNamespaces(t *testing.T) {
	const (
		podNS     = "flyte-system"
		limitNS   = "tasks"
		secretsNS = "flyte-secrets"
	)
	cfgWith := func(typ config.EmbeddedSecretManagerType, imagePull bool, k8sNS string) *config.Config {
		return &config.Config{EmbeddedSecretManagerConfig: config.EmbeddedSecretManagerConfig{
			Type:             typ,
			K8sConfig:        config.K8sConfig{Namespace: k8sNS},
			ImagePullSecrets: config.ImagePullSecretsConfig{Enabled: imagePull},
		}}
	}

	tests := []struct {
		name           string
		cfg            *config.Config
		limitNamespace string
		wantInformer   bool
		wantNamespaces []string
	}{
		{
			name:         "no image pull secrets, non-K8s backend: direct client, no informer",
			cfg:          cfgWith(config.EmbeddedSecretManagerTypeAWS, false, ""),
			wantInformer: false,
		},
		{
			name:           "no image pull secrets, non-K8s backend, limit set: still no informer",
			cfg:            cfgWith(config.EmbeddedSecretManagerTypeGCP, false, ""),
			limitNamespace: limitNS,
			wantInformer:   false,
		},
		{
			name:         "image pull secrets without limit: cluster-wide informer",
			cfg:          cfgWith(config.EmbeddedSecretManagerTypeAWS, true, ""),
			wantInformer: true,
		},
		{
			name:           "image pull secrets with limit 'all': cluster-wide informer",
			cfg:            cfgWith(config.EmbeddedSecretManagerTypeAWS, true, ""),
			limitNamespace: AllNamespaces,
			wantInformer:   true,
		},
		{
			name:           "image pull secrets with limit: scoped to limit + reference namespace",
			cfg:            cfgWith(config.EmbeddedSecretManagerTypeAWS, true, ""),
			limitNamespace: limitNS,
			wantInformer:   true,
			wantNamespaces: []string{limitNS, podNS},
		},
		{
			name:           "K8s backend only: scoped to the secrets namespace regardless of limit",
			cfg:            cfgWith(config.EmbeddedSecretManagerTypeK8s, false, secretsNS),
			wantInformer:   true,
			wantNamespaces: []string{secretsNS},
		},
		{
			name:           "K8s backend + image pull secrets with limit: union, deduplicated",
			cfg:            cfgWith(config.EmbeddedSecretManagerTypeK8s, true, limitNS),
			limitNamespace: limitNS,
			wantInformer:   true,
			wantNamespaces: []string{limitNS, podNS},
		},
		{
			name:         "K8s backend + image pull secrets without limit: cluster-wide",
			cfg:          cfgWith(config.EmbeddedSecretManagerTypeK8s, true, secretsNS),
			wantInformer: true,
		},
		{
			name:         "K8s backend without a namespace: direct client rather than cluster-wide watch",
			cfg:          cfgWith(config.EmbeddedSecretManagerTypeK8s, false, ""),
			wantInformer: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotInformer, gotNamespaces := secretInformerNamespaces(tt.cfg, podNS, tt.limitNamespace)
			assert.Equal(t, tt.wantInformer, gotInformer)
			assert.Equal(t, tt.wantNamespaces, gotNamespaces)
		})
	}
}

func TestNewSecretsMutatorWithOptions(t *testing.T) {
	// Global + K8s injectors need no cluster access, so this exercises option plumbing without
	// an API server.
	cfg := &config.Config{SecretManagerTypes: []config.SecretManagerType{config.SecretManagerTypeK8s}}

	m, err := NewSecretsMutatorWithOptions(context.Background(), cfg, "flyte", promutils.NewTestScope(),
		WithLimitNamespace("tasks"))
	require.NoError(t, err)
	assert.Equal(t,
		[]config.SecretManagerType{config.SecretManagerTypeGlobal, config.SecretManagerTypeK8s},
		m.enabledSecretManagerTypes)

	legacy, err := NewSecretsMutator(context.Background(), cfg, "flyte", promutils.NewTestScope())
	require.NoError(t, err)
	assert.Equal(t, m.enabledSecretManagerTypes, legacy.enabledSecretManagerTypes)

	o := mutatorOptions{}
	WithLimitNamespace("tasks")(&o)
	assert.Equal(t, "tasks", o.limitNamespace)
}

func TestNewSecretsMutatorFromInjectors(t *testing.T) {
	injectors := map[config.SecretManagerType]SecretsInjector{
		config.SecretManagerTypeK8s: NewK8sSecretsInjector(&config.Config{}),
	}
	m := NewSecretsMutatorFromInjectors([]config.SecretManagerType{config.SecretManagerTypeK8s}, injectors)
	assert.Equal(t, []config.SecretManagerType{config.SecretManagerTypeK8s}, m.enabledSecretManagerTypes)
	assert.Equal(t, injectors, m.injectors)
}
