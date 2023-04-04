package webhook

import (
	"context"
	"fmt"
	"testing"

	coreIdl "github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytepropeller/pkg/webhook/config"
	"github.com/go-test/deep"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	// We expect these outputs for each successful test
	PodSpec = corev1.PodSpec{
		InitContainers: []corev1.Container{},
		Containers: []corev1.Container{
			{
				Name: "container1",
				Env: []corev1.EnvVar{
					{
						Name:  "FLYTE_SECRETS_DEFAULT_DIR",
						Value: "/etc/flyte/secrets",
					},
					{
						Name: "FLYTE_SECRETS_FILE_PREFIX",
					},
				},
			},
		},
	}
)

func RetrieveUUID(annotations map[string]string) string {
	// helper function to retrieve the random uuid from output before comparing
	var uuid string
	for k := range annotations {
		if len(k) > 39 && k[:39] == "vault.hashicorp.com/agent-inject-secret" {
			uuid = k[40:]
		}
	}
	return uuid
}

func ExpectedKVv1(uuid string) *corev1.Pod {
	// Injects uuid into expected output for KV v1 secrets
	expected := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				"vault.hashicorp.com/agent-inject":                                "true",
				"vault.hashicorp.com/secret-volume-path":                          "/etc/flyte/secrets",
				"vault.hashicorp.com/role":                                        "flyte",
				"vault.hashicorp.com/agent-pre-populate-only":                     "true",
				fmt.Sprintf("vault.hashicorp.com/agent-inject-secret-%s", uuid):   "foo",
				fmt.Sprintf("vault.hashicorp.com/agent-inject-file-%s", uuid):     "foo/bar",
				fmt.Sprintf("vault.hashicorp.com/agent-inject-template-%s", uuid): `{{- with secret "foo" -}}{{ .Data.bar }}{{- end -}}`,
			},
		},
		Spec: PodSpec,
	}
	return expected
}

func ExpectedKVv2(uuid string) *corev1.Pod {
	// Injects uuid into expected output for KV v2 secrets
	expected := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				"vault.hashicorp.com/agent-inject":                                "true",
				"vault.hashicorp.com/secret-volume-path":                          "/etc/flyte/secrets",
				"vault.hashicorp.com/role":                                        "flyte",
				"vault.hashicorp.com/agent-pre-populate-only":                     "true",
				fmt.Sprintf("vault.hashicorp.com/agent-inject-secret-%s", uuid):   "foo",
				fmt.Sprintf("vault.hashicorp.com/agent-inject-file-%s", uuid):     "foo/bar",
				fmt.Sprintf("vault.hashicorp.com/agent-inject-template-%s", uuid): `{{- with secret "foo" -}}{{ .Data.data.bar }}{{- end -}}`,
			},
		},
		Spec: PodSpec,
	}
	return expected
}

func NewInputPod() *corev1.Pod {
	// Need to create a new Pod for every test since annotations are otherwise appended to original reference object
	p := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name: "container1",
				},
			},
		},
	}
	return p
}

func TestVaultSecretManagerInjector_Inject(t *testing.T) {
	inputSecret := &coreIdl.Secret{
		Group: "foo",
		Key:   "bar",
	}

	ctx := context.Background()
	type args struct {
		cfg    config.VaultSecretManagerConfig
		secret *coreIdl.Secret
		p      *corev1.Pod
	}
	tests := []struct {
		name    string
		args    args
		want    func(string) *corev1.Pod
		wantErr bool
	}{
		{
			name: "KVv1 Secret",
			args: args{
				cfg:    config.VaultSecretManagerConfig{Role: "flyte", KVVersion: config.KVVersion1},
				secret: inputSecret,
				p:      NewInputPod(),
			},
			want:    ExpectedKVv1,
			wantErr: false,
		},
		{
			name: "KVv2 Secret",
			args: args{
				cfg:    config.VaultSecretManagerConfig{Role: "flyte", KVVersion: config.KVVersion2},
				secret: inputSecret,
				p:      NewInputPod(),
			},
			want:    ExpectedKVv2,
			wantErr: false,
		},
		{
			name: "Unsupported KV version",
			args: args{
				cfg:    config.VaultSecretManagerConfig{Role: "flyte", KVVersion: 3},
				secret: inputSecret,
				p:      NewInputPod(),
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := NewVaultSecretManagerInjector(tt.args.cfg)
			got, _, err := i.Inject(ctx, tt.args.secret, tt.args.p)

			if (err != nil) != tt.wantErr {
				t.Errorf("Inject() error = %v, wantErr %v", err, tt.wantErr)
				return
			} else if err != nil {
				return
			}

			uuid := RetrieveUUID(got.ObjectMeta.Annotations)
			expected := tt.want(uuid)
			if diff := deep.Equal(got, expected); diff != nil {
				t.Errorf("Inject() Diff = %v\r\n got = %v\r\n want = %v", diff, got, expected)
			}
		})
	}
}
