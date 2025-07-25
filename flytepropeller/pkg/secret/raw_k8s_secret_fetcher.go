package secret

import (
	"context"
	"errors"
	"fmt"
	"strings"

	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	stdlibErrors "github.com/flyteorg/flyte/flytestdlib/errors"
	"github.com/flyteorg/flyte/flytestdlib/logger"
)

type RawK8sSecretFetcher struct {
	kubeClientset *kubernetes.Clientset
}

func (s RawK8sSecretFetcher) GetSecretValue(ctx context.Context, secretID string) (*SecretValue, error) {
	logger.Infof(ctx, "Got fetch secret Request for %v", secretID)
	secretNameComponents, err := DecodeSecretName(secretID)
	if err != nil {
		wrappedErr := stdlibErrors.Wrapf(ErrCodeSecretReadFailure, err, fmt.Sprintf(SecretReadFailureErrorFormat, secretID))
		logger.Error(ctx, wrappedErr)
		return nil, wrappedErr
	}

	secretNamespaceGroupKey := strings.Split(secretNameComponents.Name, NamespaceGroupKeyDelimiter)
	if len(secretNamespaceGroupKey) != 3 {
		wrappedErr := stdlibErrors.Wrapf(ErrCodeSecretReadFailure, errors.New("invalid secret ID format"), fmt.Sprintf(SecretReadFailureErrorFormat, secretID))
		logger.Error(ctx, wrappedErr)
		return nil, wrappedErr
	}
	secretNamespace, secretGroup, secretKey := secretNamespaceGroupKey[0], secretNamespaceGroupKey[1], secretNamespaceGroupKey[2]
	secretClient := s.kubeClientset.CoreV1().Secrets(secretNamespace)
	secret, err := secretClient.Get(ctx, secretGroup, metav1.GetOptions{})

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

	secretBinaryValue, ok := secret.Data[secretKey]
	if !ok {
		wrappedErr := stdlibErrors.Wrapf(ErrCodeSecretNil, errors.New("secret data is nil"), fmt.Sprintf(SecretNilErrorFormat, secretID))
		logger.Error(ctx, wrappedErr)
		return nil, wrappedErr
	}

	// Since all keys and values are merged into the data field on write, we can just return the value in the data field.
	secretValue := &SecretValue{
		BinaryValue: secretBinaryValue,
	}

	return secretValue, nil
}

// NewRawK8sSecretFetcher creates a secret value fetcher for K8s
func NewRawK8sSecretFetcher(kubeClientset *kubernetes.Clientset) SecretFetcher {
	return RawK8sSecretFetcher{
		kubeClientset: kubeClientset,
	}
}
