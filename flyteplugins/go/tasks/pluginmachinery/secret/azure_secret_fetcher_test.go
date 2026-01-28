package secret

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"
	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/secret/mocks"
	stdlibErrors "github.com/flyteorg/flyte/v2/flytestdlib/errors"
)

func setupAzureSecretFetcherTest() (SecretFetcher, *mocks.AzureKeyVaultClient) {
	client := &mocks.AzureKeyVaultClient{}
	return NewAzureSecretFetcher(client), client
}

func Test_AzureSecretFetcher_GetSecretValue(t *testing.T) {
	secretName := "test-secret"
	validSecretIDFmt := "u__org__test-org__domain__test-domain__project__test-project__key__%s" // #nosec G101
	validSecretID := fmt.Sprintf(validSecretIDFmt, secretName)
	validEncodedSecretID, _ := EncodeAzureSecretName(validSecretID)

	sucessfulTests := []struct {
		name        string
		secretIDArg string
		returnValue string
	}{
		{
			"test successful string retrieval",
			validSecretID,
			"test-secret-value",
		},
	}

	ctx := context.Background()

	for _, tt := range sucessfulTests {
		t.Run(tt.name, func(t *testing.T) {
			fetcher, client := setupAzureSecretFetcherTest()
			jsonData, err := json.Marshal(azureSecretValue{
				Type:  azureSecretValueTypeSTRING,
				Value: tt.returnValue,
			})
			assert.NoError(t, err)
			returnValue := string(jsonData)
			mockedResponse := azsecrets.GetSecretResponse{Secret: azsecrets.Secret{Value: &returnValue}}
			client.OnGetSecret(ctx, validEncodedSecretID, azureLatestVersion, nil).Return(mockedResponse, nil)

			sv, err := fetcher.GetSecretValue(ctx, tt.secretIDArg)
			if err != nil {
				t.Errorf("GetSecretValue() got error %v", err)
			}
			assert.Equal(t, tt.returnValue, sv.StringValue)
		})
	}

	sucessfulBinaryTests := []struct {
		name        string
		secretIDArg string
		returnValue *[]byte
	}{
		{
			"test successful binary retrieval",
			validSecretID,
			to.Ptr([]byte{0x48, 0x65, 0x6C, 0x6C, 0x6F}),
		},
	}

	for _, tt := range sucessfulBinaryTests {
		t.Run(tt.name, func(t *testing.T) {
			fetcher, client := setupAzureSecretFetcherTest()
			jsonData, err := json.Marshal(azureSecretValue{
				Type:  azureSecretValueTypeBINARY,
				Value: base64.StdEncoding.EncodeToString(*tt.returnValue),
			})
			assert.NoError(t, err)
			returnValue := string(jsonData)
			mockedResponse := azsecrets.GetSecretResponse{Secret: azsecrets.Secret{Value: &returnValue}}
			client.OnGetSecret(ctx, validEncodedSecretID, azureLatestVersion, nil).Return(mockedResponse, nil)

			sv, err := fetcher.GetSecretValue(ctx, tt.secretIDArg)
			if err != nil {
				t.Errorf("GetSecretValue() got error %v", err)
			}
			assert.Equal(t, *tt.returnValue, sv.BinaryValue)
		})
	}

	t.Run("Azure Key Vault returns not found error", func(t *testing.T) {
		fetcher, client := setupAzureSecretFetcherTest()

		azError := AzureErrorResponse{
			Error: AzureError{
				Code:    "NotFound",
				Message: "Not found",
			},
		}
		rawResponseBody, err := json.Marshal(azError)
		assert.NoError(t, err)
		respError := runtime.NewResponseError(&http.Response{
			StatusCode: http.StatusNotFound,
			Header: http.Header{
				"x-ms-error-code": []string{"NotFound"},
			},
			Body: io.NopCloser(bytes.NewBufferString(string(rawResponseBody))),
		})
		client.OnGetSecret(ctx, validEncodedSecretID, azureLatestVersion, nil).Return(azsecrets.GetSecretResponse{}, respError)

		_, err = fetcher.GetSecretValue(ctx, validSecretID)
		assert.Equal(t, stdlibErrors.Wrapf(ErrCodeSecretNotFound, respError, fmt.Sprintf(SecretNotFoundErrorFormat, validSecretID)), err)
	})

	t.Run("Azure Key Vault returns unexpected error", func(t *testing.T) {
		fetcher, client := setupAzureSecretFetcherTest()
		cause := fmt.Errorf("test-error")
		client.OnGetSecret(ctx, validEncodedSecretID, azureLatestVersion, nil).Return(azsecrets.GetSecretResponse{}, cause)

		_, err := fetcher.GetSecretValue(ctx, validSecretID)
		assert.Equal(t, stdlibErrors.Wrapf(ErrCodeSecretReadFailure, cause, fmt.Sprintf(SecretReadFailureErrorFormat, validSecretID)), err)
	})

	emptyStr := ""
	emptyResultTests := []struct {
		name        string
		secretIDArg string
		returnValue *string
	}{
		{
			"nil",
			validSecretID,
			nil,
		},
		{
			"empty",
			validSecretID,
			&emptyStr,
		},
	}

	for _, tt := range emptyResultTests {
		t.Run(fmt.Sprintf("Azure Key Vault returns %v value", tt.name), func(t *testing.T) {
			fetcher, client := setupAzureSecretFetcherTest()
			client.OnGetSecret(ctx, validEncodedSecretID, azureLatestVersion, nil).Return(azsecrets.GetSecretResponse{Secret: azsecrets.Secret{Value: tt.returnValue}}, nil)
			_, err := fetcher.GetSecretValue(ctx, tt.secretIDArg)
			assert.Equal(t, stdlibErrors.Wrapf(ErrCodeSecretNil, nil, fmt.Sprintf(SecretNilErrorFormat, tt.secretIDArg)), err)
		})
	}
}
