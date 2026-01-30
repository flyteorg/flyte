package secret

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	k8sTesting "k8s.io/client-go/testing"

	stdlibErrors "github.com/flyteorg/flyte/v2/flytestdlib/errors"
)

const (
	testNamespace = "test-namespace"
)

func TestGetSecretValue(t *testing.T) {
	testCases := []struct {
		name             string
		secretsName      []string
		targetSecretName string
		expectSuccess    bool
		expectedError    error
		includeData      bool
		reactors         []func(action k8sTesting.Action) (handled bool, ret runtime.Object, err error)
	}{
		{
			name: "Get Secret Successful",
			secretsName: []string{
				"test-secret",
			},
			targetSecretName: "test-secret",
			expectSuccess:    true,
			expectedError:    nil,
			includeData:      true,
		},
		{
			name:             "Get Secret Not Found",
			secretsName:      []string{},
			targetSecretName: "test-secret",
			expectSuccess:    false,
			expectedError:    stdlibErrors.Wrapf(ErrCodeSecretNotFound, k8sErrors.NewNotFound(v1.Resource("secrets"), EncodeK8sSecretName("test-secret")), fmt.Sprintf(SecretNotFoundErrorFormat, "test-secret")),
			includeData:      false,
		},
		{
			name: "Get Secret Failed With Wrong Data",
			secretsName: []string{
				"test-secret",
			},
			targetSecretName: "test-secret",
			expectSuccess:    false,
			expectedError:    stdlibErrors.Wrapf(ErrCodeSecretNil, errors.New("secret data is nil"), fmt.Sprintf(SecretNilErrorFormat, "test-secret")),
			includeData:      false,
		},
		{
			name: "Get Secret Failed With K8s Client Error",
			secretsName: []string{
				"test-secret",
			},
			targetSecretName: "test-secret",
			expectSuccess:    false,
			expectedError:    stdlibErrors.Wrapf(ErrCodeSecretReadFailure, errors.New("k8s client error"), fmt.Sprintf(SecretReadFailureErrorFormat, "test-secret")),
			includeData:      false,
			reactors: []func(action k8sTesting.Action) (handled bool, ret runtime.Object, err error){
				func(action k8sTesting.Action) (handled bool, ret runtime.Object, err error) {
					if action.GetVerb() == "get" && action.GetResource().Resource == "secrets" {
						return true, nil, errors.New("k8s client error")
					}
					return false, nil, nil
				},
			},
		},
	}
	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			testSecrets := lo.Map(test.secretsName, func(secretName string, _ int) runtime.Object {
				secret := &v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      EncodeK8sSecretName(secretName),
						Namespace: testNamespace,
					},
				}
				if test.includeData {
					secret.Data = map[string][]byte{
						secretName: []byte(secretName),
					}
				}
				return secret
			})
			fakeClientset := fake.NewSimpleClientset(testSecrets...)
			if len(test.reactors) > 0 {
				for _, reactor := range test.reactors {
					fakeClientset.PrependReactor("*", "secrets", reactor)
				}
			}
			fetcher := NewK8sSecretFetcher(fakeClientset.CoreV1().Secrets(testNamespace))
			secret, err := fetcher.GetSecretValue(context.TODO(), test.targetSecretName)
			if test.expectSuccess {
				assert.NoError(t, err)
				assert.NotNil(t, secret)
				assert.Equal(t, []byte(test.targetSecretName), secret.BinaryValue)
			} else {
				assert.Error(t, err)
				assert.Equal(t, test.expectedError.Error(), err.Error())
			}
		})
	}
}
