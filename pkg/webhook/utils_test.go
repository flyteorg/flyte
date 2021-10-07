package webhook

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/go-test/deep"

	corev1 "k8s.io/api/core/v1"
)

func Test_hasEnvVar(t *testing.T) {
	type args struct {
		envVars   []corev1.EnvVar
		envVarKey string
	}
	tests := []struct {
		name string
		args args
		want bool
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
			want: true,
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
			want: false,
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
							Value: "my_val_already",
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
