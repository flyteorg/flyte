package secret

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"

	stdlibErrors "github.com/flyteorg/flyte/v2/flytestdlib/errors"
	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
)

const (
	azureLatestVersion string = "" // Empty string represents the latest version of the secret
)

type AzureSecretFetcher struct {
	// The Azure Key Vault client to use for fetching secrets
	client AzureKeyVaultClient
}

func (a AzureSecretFetcher) GetSecretValue(ctx context.Context, secretID string) (*SecretValue, error) {
	secretName, err := EncodeAzureSecretName(secretID)
	if err != nil {
		logger.Errorf(ctx, "Failed to encode secret name %v: %w", secretID, err)
		return nil, stdlibErrors.Wrapf(ErrCodeSecretReadFailure, err, fmt.Sprintf(SecretReadFailureErrorFormat, secretID))
	}
	resp, err := a.client.GetSecret(ctx, secretName, azureLatestVersion, nil)
	if err != nil {
		var notFound *azcore.ResponseError
		if errors.As(err, &notFound) && notFound.StatusCode == http.StatusNotFound {
			logger.Errorf(ctx, "Azure secret not found: %w", secretID)
			return nil, stdlibErrors.Wrapf(ErrCodeSecretNotFound, err, fmt.Sprintf(SecretNotFoundErrorFormat, secretID))
		}
		logger.Errorf(ctx, "Failed to fetch secret %v: %w", secretID, err)
		return nil, stdlibErrors.Wrapf(ErrCodeSecretReadFailure, err, fmt.Sprintf(SecretReadFailureErrorFormat, secretID))
	}

	if resp.Value == nil || len(*resp.Value) == 0 {
		logger.Errorf(ctx, "Secret %v returned empty value", secretID)
		return nil, stdlibErrors.Wrapf(ErrCodeSecretNil, err, fmt.Sprintf(SecretNilErrorFormat, secretID))
	}

	sv, err := AzureToUnionSecret(resp.Secret)
	if err != nil {
		logger.Errorf(ctx, "Failed to convert Azure secret %v to Union secret", secretID)
		return nil, stdlibErrors.Wrapf(ErrCodeSecretReadFailure, err, fmt.Sprintf(SecretReadFailureErrorFormat, secretID))
	}
	return sv, nil
}

func NewAzureSecretFetcher(client AzureKeyVaultClient) SecretFetcher {
	return &AzureSecretFetcher{client: client}
}

func getAzureKVSecretsClient(vaultURI string) (*azsecrets.Client, error) {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get azure credentials: %w", err)
	}

	client, err := azsecrets.NewClient(vaultURI, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create azure secrets client: %w", err)
	}

	return client, nil
}
