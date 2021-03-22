package webhook

import (
	"path/filepath"
	"reflect"
	"testing"

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
			if got := UpdateVolumeMounts(tt.args.containers, tt.args.volumeMount); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UpdateVolumeMounts() = %v, want %v", got, tt.want)
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
			if got := UpdateEnvVars(tt.args.containers, tt.args.envVar); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UpdateEnvVars() = %v, want %v", got, tt.want)
			}
		})
	}
}
