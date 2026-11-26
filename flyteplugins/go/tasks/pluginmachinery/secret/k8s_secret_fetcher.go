package secret

import (
	"context"
	"errors"

	corev1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	stdlibErrors "github.com/flyteorg/flyte/v2/flytestdlib/errors"
	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
)

type K8sSecretFetcher struct {
	getSecret func(ctx context.Context, name string) (*corev1.Secret, error)
}

func (s K8sSecretFetcher) GetSecretValue(ctx context.Context, secretID string) (*SecretValue, error) {
	logger.Infof(ctx, "Got fetch secret Request for %v!", secretID)
	k8sSecretName := EncodeK8sSecretName(secretID)
	secret, err := s.getSecret(ctx, k8sSecretName)

	if err != nil {
		if k8sErrors.IsNotFound(err) {
			wrappedErr := stdlibErrors.Wrapf(ErrCodeSecretNotFound, err, SecretNotFoundErrorFormat, secretID)
			logger.Warn(ctx, wrappedErr)
			return nil, wrappedErr
		} else {
			wrappedErr := stdlibErrors.Wrapf(ErrCodeSecretReadFailure, err, SecretReadFailureErrorFormat, secretID)
			logger.Error(ctx, wrappedErr)
			return nil, wrappedErr
		}
	}

	if _, ok := secret.Data[secretID]; !ok {
		wrappedErr := stdlibErrors.Wrapf(ErrCodeSecretNil, errors.New("secret data is nil"), SecretNilErrorFormat, secretID)
		logger.Error(ctx, wrappedErr)
		return nil, wrappedErr
	}

	// Since all keys and values are merged into the data field on write, we can just return the value in the data field.
	secretValue := &SecretValue{
		BinaryValue: secret.Data[secretID],
	}

	return secretValue, nil
}

// NewK8sSecretFetcher creates a secret value fetcher for K8s
func NewK8sSecretFetcher(secretClient v1.SecretInterface) SecretFetcher {
	return K8sSecretFetcher{
		getSecret: func(ctx context.Context, name string) (*corev1.Secret, error) {
			return secretClient.Get(ctx, name, metav1.GetOptions{})
		},
	}
}

// NewK8sSecretFetcherFromReader creates a secret value fetcher for K8s that reads the secrets in
// namespace through reader, e.g. a client backed by a Secret informer.
func NewK8sSecretFetcherFromReader(reader client.Reader, namespace string) SecretFetcher {
	return K8sSecretFetcher{
		getSecret: func(ctx context.Context, name string) (*corev1.Secret, error) {
			secret := &corev1.Secret{}
			if err := reader.Get(ctx, types.NamespacedName{Namespace: namespace, Name: name}, secret); err != nil {
				return nil, err
			}
			return secret, nil
		},
	}
}
