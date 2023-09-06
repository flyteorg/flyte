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

func ExpectedExtraAnnotation(uuid string) *corev1.Pod {
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
				"vault.hashicorp.com/auth-config-type":                            "gce",
			},
		},
		Spec: PodSpec,
	}
	return expected
}

func ExpectedExistingRoleAnnotation(uuid string) *corev1.Pod {
	expected := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				"vault.hashicorp.com/agent-inject":                                "true",
				"vault.hashicorp.com/secret-volume-path":                          "/etc/flyte/secrets",
				"vault.hashicorp.com/role":                                        "my-role",
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

func ExpectedConfigAnnotation(uuid string) *corev1.Pod {
	expected := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				"vault.hashicorp.com/agent-inject":                                "true",
				"vault.hashicorp.com/secret-volume-path":                          "/etc/flyte/secrets",
				"vault.hashicorp.com/role":                                        "flyte",
				"vault.hashicorp.com/agent-pre-populate-only":                     "false",
				fmt.Sprintf("vault.hashicorp.com/agent-inject-secret-%s", uuid):   "foo",
				fmt.Sprintf("vault.hashicorp.com/agent-inject-file-%s", uuid):     "foo/bar",
				fmt.Sprintf("vault.hashicorp.com/agent-inject-template-%s", uuid): `{{- with secret "foo" -}}{{ .Data.data.bar }}{{- end -}}`,
			},
		},
		Spec: PodSpec,
	}
	return expected
}

func ExpectedDB(uuid string) *corev1.Pod {
	expected := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				"vault.hashicorp.com/agent-inject":                              "true",
				"vault.hashicorp.com/secret-volume-path":                        "/etc/flyte/secrets",
				"vault.hashicorp.com/role":                                      "flyte",
				"vault.hashicorp.com/agent-pre-populate-only":                   "true",
				fmt.Sprintf("vault.hashicorp.com/agent-inject-secret-%s", uuid): "foo",
				fmt.Sprintf("vault.hashicorp.com/agent-inject-file-%s", uuid):   "foo/bar",
			},
		},
		Spec: PodSpec,
	}
	return expected
}

func NewInputPod(annotations map[string]string) *corev1.Pod {
	// Need to create a new Pod for every test since annotations are otherwise appended to original reference object
	p := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: annotations,
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
			name: "KVv1 Secret Group Version argument overwrites config",
			args: args{
				cfg: config.VaultSecretManagerConfig{Role: "flyte", KVVersion: config.KVVersion2},
				secret: &coreIdl.Secret{
					Group:        "foo",
					Key:          "bar",
					GroupVersion: "kv1",
				},
				p: NewInputPod(map[string]string{}),
			},
			want:    ExpectedKVv1,
			wantErr: false,
		},
		{
			name: "KVv2 Secret Group Version argument overwrites config",
			args: args{
				cfg: config.VaultSecretManagerConfig{Role: "flyte", KVVersion: config.KVVersion1},
				secret: &coreIdl.Secret{
					Group:        "foo",
					Key:          "bar",
					GroupVersion: "kv2",
				},
				p: NewInputPod(map[string]string{}),
			},
			want:    ExpectedKVv2,
			wantErr: false,
		},
		{
			name: "Extra annotations from config are added",
			args: args{
				cfg: config.VaultSecretManagerConfig{Role: "flyte", KVVersion: config.KVVersion2, Annotations: map[string]string{
					"vault.hashicorp.com/auth-config-type": "gce",
				}},
				secret: inputSecret,
				p:      NewInputPod(map[string]string{}),
			},
			want:    ExpectedExtraAnnotation,
			wantErr: false,
		},
		{
			name: "Already present annotation is not overwritten",
			args: args{
				cfg:    config.VaultSecretManagerConfig{Role: "flyte", KVVersion: config.KVVersion2, Annotations: map[string]string{}},
				secret: inputSecret,
				p: NewInputPod(map[string]string{
					"vault.hashicorp.com/role": "my-role",
				}),
			},
			want:    ExpectedExistingRoleAnnotation,
			wantErr: false,
		},
		{
			name: "Config annotation overwrites system default annotation",
			args: args{
				cfg: config.VaultSecretManagerConfig{Role: "flyte", KVVersion: config.KVVersion2, Annotations: map[string]string{
					"vault.hashicorp.com/agent-pre-populate-only": "false", // override vault.hashicorp.com/agent-pre-populate-only
				}},
				secret: inputSecret,
				p:      NewInputPod(map[string]string{}),
			},
			want:    ExpectedConfigAnnotation,
			wantErr: false,
		},
		{
			name: "DB Secret backend enginge is supported",
			args: args{
				cfg: config.VaultSecretManagerConfig{Role: "flyte", KVVersion: config.KVVersion1},
				secret: &coreIdl.Secret{
					Group:        "foo",
					Key:          "bar",
					GroupVersion: "db",
				},
				p: NewInputPod(map[string]string{}),
			},
			want:    ExpectedDB,
			wantErr: false,
		},
		{
			name: "Legacy config option V1 is still supported",
			args: args{
				cfg: config.VaultSecretManagerConfig{Role: "flyte", KVVersion: config.KVVersion1},
				secret: &coreIdl.Secret{
					Group: "foo",
					Key:   "bar",
				},
				p: NewInputPod(map[string]string{}),
			},
			want:    ExpectedKVv1,
			wantErr: false,
		},
		{
			name: "Legacy config option V2 is still supported",
			args: args{
				cfg: config.VaultSecretManagerConfig{Role: "flyte", KVVersion: config.KVVersion2},
				secret: &coreIdl.Secret{
					Group: "foo",
					Key:   "bar",
				},
				p: NewInputPod(map[string]string{}),
			},
			want:    ExpectedKVv2,
			wantErr: false,
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
