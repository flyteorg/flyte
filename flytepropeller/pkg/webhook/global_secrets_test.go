package webhook

import (
	"context"
	"fmt"
	"testing"

	"github.com/flyteorg/flytepropeller/pkg/webhook/config"

	"github.com/stretchr/testify/assert"

	"github.com/go-test/deep"

	coreIdl "github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytepropeller/pkg/webhook/mocks"
	"github.com/stretchr/testify/mock"
	corev1 "k8s.io/api/core/v1"
)

func TestGlobalSecrets_Inject(t *testing.T) {
	secretFound := &mocks.GlobalSecretProvider{}
	secretFound.OnGetForSecretMatch(mock.Anything, mock.Anything).Return("my_password", nil)

	secretNotFound := &mocks.GlobalSecretProvider{}
	secretNotFound.OnGetForSecretMatch(mock.Anything, mock.Anything).Return("", fmt.Errorf("secret not found"))

	inputPod := corev1.Pod{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name: "container1",
				},
			},
		},
	}

	successPod := corev1.Pod{
		Spec: corev1.PodSpec{
			InitContainers: []corev1.Container{},
			Containers: []corev1.Container{
				{
					Name: "container1",
					Env: []corev1.EnvVar{
						{
							Name:  "FLYTE_SECRETS_ENV_PREFIX",
							Value: "_FSEC_",
						},
						{
							Name:  "_FSEC_GROUP_HELLO",
							Value: "my_password",
						},
					},
				},
			},
		},
	}

	type args struct {
		secret *coreIdl.Secret
		p      *corev1.Pod
	}
	tests := []struct {
		name             string
		envSecretManager GlobalSecretProvider
		args             args
		want             *corev1.Pod
		wantErr          bool
	}{
		{name: "require group", envSecretManager: secretFound, args: args{secret: &coreIdl.Secret{Key: "hello"}, p: &corev1.Pod{}},
			want: &corev1.Pod{}, wantErr: true},
		{name: "simple", envSecretManager: secretFound, args: args{secret: &coreIdl.Secret{Group: "group", Key: "hello"}, p: &inputPod},
			want: &successPod, wantErr: false},
		{name: "require file", envSecretManager: secretFound, args: args{secret: &coreIdl.Secret{Key: "hello", MountRequirement: coreIdl.Secret_FILE},
			p: &corev1.Pod{}},
			want: &corev1.Pod{}, wantErr: true},
		{name: "not found", envSecretManager: secretNotFound, args: args{secret: &coreIdl.Secret{Key: "hello", MountRequirement: coreIdl.Secret_FILE},
			p: &corev1.Pod{}},
			want: &corev1.Pod{}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := GlobalSecrets{
				envSecretManager: tt.envSecretManager,
			}

			assert.Equal(t, config.SecretManagerTypeGlobal, g.Type())

			got, _, err := g.Inject(context.Background(), tt.args.secret, tt.args.p)
			if (err != nil) != tt.wantErr {
				t.Errorf("Inject() error = %v, wantErr %v", err, tt.wantErr)
				return
			} else if err != nil {
				return
			}

			if diff := deep.Equal(got, tt.want); diff != nil {
				t.Errorf("Inject() Diff = %v\r\n got = %v\r\n want = %v", diff, got, tt.want)
			}
		})
	}
}
