package secret

import (
	"errors"
	"fmt"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/go-test/deep"
	corev1 "k8s.io/api/core/v1"
)

const (
	testOrg        = "test-org"
	testDomain     = "test-domain"
	testProject    = "test-project"
	testSecretName = "test-secret"
)

func Test_hasEnvVar(t *testing.T) {
	type args struct {
		envVars   []corev1.EnvVar
		envVarKey string
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "exists",
			args: args{
				envVars: []corev1.EnvVar{
					{
						Name: "ENV_VAR_1",
					},
				},
				envVarKey: "ENV_VAR_1",
			},
			want: 0,
		},

		{
			name: "doesn't exist",
			args: args{
				envVars: []corev1.EnvVar{
					{
						Name: "ENV_VAR_1",
					},
				},
				envVarKey: "ENV_VAR",
			},
			want: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hasEnvVar(tt.args.envVars, tt.args.envVarKey); got != tt.want {
				t.Errorf("hasEnvVar() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUpdateVolumeMounts(t *testing.T) {
	type args struct {
		containers  []corev1.Container
		volumeMount corev1.VolumeMount
	}
	tests := []struct {
		name string
		args args
		want []corev1.Container
	}{
		{
			name: "volume",
			args: args{
				containers: []corev1.Container{
					{
						Name: "my_container",
					},
				},
				volumeMount: corev1.VolumeMount{
					Name:      "my_secret",
					ReadOnly:  true,
					MountPath: filepath.Join(filepath.Join(K8sSecretPathPrefix...), "my_secret"),
				},
			},
			want: []corev1.Container{
				{
					Name: "my_container",
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      "my_secret",
							ReadOnly:  true,
							MountPath: filepath.Join(filepath.Join(K8sSecretPathPrefix...), "my_secret"),
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := AppendVolumeMounts(tt.args.containers, tt.args.volumeMount); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AppendVolumeMounts() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUpdateEnvVars(t *testing.T) {
	type args struct {
		containers []corev1.Container
		envVar     corev1.EnvVar
	}
	tests := []struct {
		name string
		args args
		want []corev1.Container
	}{
		{
			name: "env vars already exists",
			args: args{
				containers: []corev1.Container{
					{
						Name: "my_container",
						Env: []corev1.EnvVar{
							{
								Name:  "my_secret",
								Value: "my_val_already",
							},
						},
					},
				},
				envVar: corev1.EnvVar{
					Name:  "my_secret",
					Value: "my_val",
				},
			},
			want: []corev1.Container{
				{
					Name: "my_container",
					Env: []corev1.EnvVar{
						{
							Name:  "my_secret",
							Value: "my_val",
						},
					},
				},
			},
		},
		{
			name: "env vars already added",
			args: args{
				containers: []corev1.Container{
					{
						Name: "my_container",
					},
				},
				envVar: corev1.EnvVar{
					Name:  "my_secret",
					Value: "my_val",
				},
			},
			want: []corev1.Container{
				{
					Name: "my_container",
					Env: []corev1.EnvVar{
						{
							Name:  "my_secret",
							Value: "my_val",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := AppendEnvVars(tt.args.containers, tt.args.envVar); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AppendEnvVars() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAppendVolume(t *testing.T) {
	type args struct {
		volumes []corev1.Volume
		volume  corev1.Volume
	}
	tests := []struct {
		name string
		args args
		want []corev1.Volume
	}{
		{name: "append secret", args: args{volumes: []corev1.Volume{}, volume: corev1.Volume{Name: "new_secret"}}, want: []corev1.Volume{{Name: "new_secret"}}},
		{name: "existing other volumes", args: args{volumes: []corev1.Volume{{Name: "existing"}}, volume: corev1.Volume{Name: "new_secret"}}, want: []corev1.Volume{{Name: "existing"}, {Name: "new_secret"}}},
		{name: "existing secret volume",
			args: args{
				volumes: []corev1.Volume{{
					Name: "existing", VolumeSource: corev1.VolumeSource{Secret: &corev1.SecretVolumeSource{SecretName: "foo", Items: []corev1.KeyToPath{{Key: "existingKey"}}}},
				}},
				volume: corev1.Volume{Name: "new_secret", VolumeSource: corev1.VolumeSource{Secret: &corev1.SecretVolumeSource{SecretName: "foo", Items: []corev1.KeyToPath{{Key: "newKey"}}}}}},
			want: []corev1.Volume{{
				Name: "existing", VolumeSource: corev1.VolumeSource{Secret: &corev1.SecretVolumeSource{SecretName: "foo", Items: []corev1.KeyToPath{{Key: "existingKey"}, {Key: "newKey"}}}},
			}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AppendVolume(tt.args.volumes, tt.args.volume)
			if diff := deep.Equal(got, tt.want); diff != nil {
				t.Errorf("AppendVolume() = %v, want %v", got, tt.want)
				t.Errorf("Diff: %v", diff)
			}
		})
	}
}

func Test_EncodeSecretName(t *testing.T) {
	tests := []struct {
		name       string
		org        string
		domain     string
		project    string
		secretName string
		want       string
	}{
		{
			name:       "test name without domain nor project",
			org:        testOrg,
			domain:     "",
			project:    "",
			secretName: testSecretName,
			want:       "u__org__test-org__domain____project____key__test-secret",
		},
		{
			name:       "test name without project",
			org:        testOrg,
			domain:     testDomain,
			project:    "",
			secretName: testSecretName,
			want:       "u__org__test-org__domain__test-domain__project____key__test-secret",
		},
		{
			name:       "test name with project and domain",
			org:        testOrg,
			domain:     testDomain,
			project:    testProject,
			secretName: testSecretName,
			want:       "u__org__test-org__domain__test-domain__project__test-project__key__test-secret",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := EncodeSecretName(tt.org, tt.domain, tt.project, tt.secretName); got != tt.want {
				t.Errorf("EncodeSecretName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_EncodeSecretNamePrefix(t *testing.T) {
	tests := []struct {
		name    string
		org     string
		domain  string
		project string
		want    string
	}{
		{
			name:    "test with org, domain, and project",
			org:     testOrg,
			domain:  testDomain,
			project: testProject,
			want:    "u__org__test-org__domain__test-domain__project__test-project",
		},
		{
			name:    "test with org, domain, but no project",
			org:     testOrg,
			domain:  testDomain,
			project: "",
			want:    "u__org__test-org__domain__test-domain",
		},
		{
			name:    "test with org but no domain nor project",
			org:     testOrg,
			domain:  "",
			project: "",
			want:    "u__org__test-org",
		},
		{
			name:    "test with no org, domain, nor project",
			org:     "",
			domain:  "",
			project: "",
			want:    "u__org__",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := EncodeSecretNamePrefix(tt.org, tt.domain, tt.project); got != tt.want {
				t.Errorf("EncodeSecretNamePrefix() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_DecodeSecretName(t *testing.T) {
	tests := []struct {
		name string
		arg  string
		want SecretNameComponents
	}{
		{
			name: "test name without domain nor project",
			arg:  "u__org__test-org__domain____project____key__test-secret",
			want: SecretNameComponents{
				Org:     testOrg,
				Domain:  "",
				Project: "",
				Name:    testSecretName,
			},
		},
		{
			name: "test name without project",
			arg:  "u__org__test-org__domain__test-domain__project____key__test-secret",
			want: SecretNameComponents{
				Org:     testOrg,
				Domain:  testDomain,
				Project: "",
				Name:    testSecretName,
			},
		},
		{
			name: "test name with project and domain",
			arg:  "u__org__test-org__domain__test-domain__project__test-project__key__test-secret",
			want: SecretNameComponents{
				Org:     testOrg,
				Domain:  testDomain,
				Project: testProject,
				Name:    testSecretName,
			},
		},
		{
			name: "test name with key that has underscores",
			arg:  "u__org__test-org__domain__test-domain__project__test-project__key__test__secret",
			want: SecretNameComponents{
				Org:     testOrg,
				Domain:  testDomain,
				Project: testProject,
				Name:    "test__secret",
			},
		},
		{
			name: "test name with key that ends with underscores",
			arg:  "u__org__test-org__domain__test-domain__project__test-project__key__test-secret__",
			want: SecretNameComponents{
				Org:     testOrg,
				Domain:  testDomain,
				Project: testProject,
				Name:    "test-secret__",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, err := DecodeSecretName(tt.arg); err != nil || !reflect.DeepEqual(*got, tt.want) {
				t.Errorf("DecodeSecretName() = %v, want %v", got, tt.want)
			}
		})
	}

	notEnoughParts := []string{
		"u_org__test-org__domain__test-domain__project__test-project__key__test-secret",
		"u__org_test-org__domain__test-domain__project__test-project__key__test-secret",
		"u__org__test-org_domain__test-domain__project__test-project__key__test-secret",
		"u__org__test-org__domain_test-domain__project__test-project__key__test-secret",
		"u__org__test-org__domain__test-domain_project__test-project__key__test-secret",
		"u__org__test-org__domain__test-domain__project_test-project__key__test-secret",
		"u__org__test-org__domain__test-domain__project__test-project_key__test-secret",
		"u__org__test-org__domain__test-domain__project__test-project__key_test-secret",
	}

	for _, arg := range notEnoughParts {
		t.Run(fmt.Sprintf("test name with not enough parts: %s", arg), func(t *testing.T) {
			expectedErr := errors.New(secretNameInvalidNotEnoughPartsMsg)
			if got, err := DecodeSecretName(arg); err == nil || errors.Is(err, expectedErr) {
				t.Errorf("DecodeSecretName() = %v, %v; want error %v", got, err, expectedErr)
			}
		})
	}

	unexpectedParts := []string{
		"v__org__test-org__domain__test-domain__project__test-project__key__test-secret",
		"u__smorg__test-org__domain__test-domain__project__test-project__key__test-secret",
		"u__org__test-org__doman__test-domain__project__test-project__key__test-secret",
		"u__org__test-org__domain__test-domain__projects__test-project__key__test-secret",
		"u__org__test-org__domain__test-domain__project__test-project__ky__test-secret",
	}

	for _, arg := range unexpectedParts {
		t.Run(fmt.Sprintf("test name with unexpected parts: %s", arg), func(t *testing.T) {
			expectedErr := errors.New(secretNameInvalidUnexpectedPartsMsg)
			if got, err := DecodeSecretName(arg); err == nil || errors.Is(err, expectedErr) {
				t.Errorf("DecodeSecretName() = %v, %v; want error %v", got, err, expectedErr)
			}
		})
	}
}

func Test_EncodeDecodeSecretName_Bijectivity(t *testing.T) {
	tests := []struct {
		name       string
		org        string
		domain     string
		project    string
		secretName string
		arg        SecretNameComponents
	}{
		{
			name:       "test name without domain nor project",
			org:        testOrg,
			domain:     "",
			project:    "",
			secretName: testSecretName,
		},
		{
			name:       "test name without project",
			org:        testOrg,
			domain:     testDomain,
			project:    "",
			secretName: testSecretName,
		},
		{
			name:       "test name with project and domain",
			org:        testOrg,
			domain:     testDomain,
			project:    testProject,
			secretName: testSecretName,
		},
		{
			name:       "test name with key that has underscores",
			org:        testOrg,
			domain:     testDomain,
			project:    testProject,
			secretName: "test__secret",
		},
		{
			name:       "test name with key that ends with underscores",
			org:        testOrg,
			domain:     testDomain,
			project:    testProject,
			secretName: "test-secret__",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded := EncodeSecretName(tt.org, tt.domain, tt.project, tt.secretName)
			decoded, err := DecodeSecretName(encoded)
			if err != nil {
				t.Errorf("DecodeSecretName() = %v, want nil", err)
			}
			if !reflect.DeepEqual(*decoded, SecretNameComponents{
				Org:     tt.org,
				Domain:  tt.domain,
				Project: tt.project,
				Name:    tt.secretName,
			}) {
				t.Errorf("DecodeSecretName() = %v, want %v", *decoded, tt.arg)
			}
		})
	}
}
