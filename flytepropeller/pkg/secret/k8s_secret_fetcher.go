package secret

import (
	"context"
	"errors"
	"fmt"

	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"

	stdlibErrors "github.com/flyteorg/flyte/flytestdlib/errors"
	"github.com/flyteorg/flyte/flytestdlib/logger"
)

type K8sSecretFetcher struct {
	secretClient v1.SecretInterface
}

func (s K8sSecretFetcher) GetSecretValue(ctx context.Context, secretID string) (*SecretValue, error) {
	logger.Infof(ctx, "Got fetch secret Request for %v!", secretID)
	k8sSecretName := EncodeK8sSecretName(secretID)
	secret, err := s.secretClient.Get(ctx, k8sSecretName, metav1.GetOptions{})

	if err != nil {
		if k8sErrors.IsNotFound(err) {
			wrappedErr := stdlibErrors.Wrapf(ErrCodeSecretNotFound, err, fmt.Sprintf(SecretNotFoundErrorFormat, secretID))
			logger.Warn(ctx, wrappedErr)
			return nil, wrappedErr
		} else {
			wrappedErr := stdlibErrors.Wrapf(ErrCodeSecretReadFailure, err, fmt.Sprintf(SecretReadFailureErrorFormat, secretID))
			logger.Error(ctx, wrappedErr)
			return nil, wrappedErr
		}
	}

	if _, ok := secret.Data[secretID]; !ok {
		wrappedErr := stdlibErrors.Wrapf(ErrCodeSecretNil, errors.New("secret data is nil"), fmt.Sprintf(SecretNilErrorFormat, secretID))
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
		secretClient: secretClient,
	}
}
