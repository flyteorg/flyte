package secret

import (
	"encoding/base32"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"
)

const (
	azureSecretNameComponentDelimiter = "-"
	azureMaxSecretNameLength          = 127
	azureEncodingVersion              = 1
	azureEscapeChar                   = "0" // unique single character to escape special characters
	azureHyphenMarker                 = '1' // unique single character to escape hyphens
	EmptyString                       = ""
)

// Custom base32 encoding without padding
var base32Encoding = base32.StdEncoding.WithPadding(base32.NoPadding)

// Azure Key Vault specific encoding for secret names.
// Union introduces __ to deliminate contextual information into the secret.
//
// This codec unwraps the Union specific Key encoding and re-encodes it to a format that is compatible with Azure Key Vault.
// The encoding scheme is versioned and is as follows:
// - Uses 0 as an escape character
// - Uses 1 to escape hyphens
// - Uses 2 to escape underscores
// - Uses base32 encoding for the name component
//
// Ref: https://azure.github.io/PSRule.Rules.Azure/en/rules/Azure.KeyVault.SecretName/
func EncodeAzureSecretName(secretName string) (string, error) {
	components, err := DecodeSecretName(secretName)
	if err != nil {
		return EmptyString, fmt.Errorf("unable to decode original secret name: %w", err)
	}

	encode := func(s string) string {
		s = strings.ReplaceAll(s, azureEscapeChar, azureEscapeChar+azureEscapeChar)
		return strings.ReplaceAll(s, "-", azureEscapeChar+string(azureHyphenMarker))
	}

	parts := []string{
		encode(components.Org),
		encode(components.Domain),
		encode(components.Project),
		base32Encoding.EncodeToString([]byte(components.Name)),
	}

	encoded := fmt.Sprint(azureEncodingVersion) + azureSecretNameComponentDelimiter + strings.Join(parts, azureSecretNameComponentDelimiter)

	if len(encoded) > azureMaxSecretNameLength {
		return EmptyString, fmt.Errorf("encoded secret name with length %d exceeds maximum length: %d", len(encoded), azureMaxSecretNameLength)
	}

	return encoded, nil
}

// Azure Key Vault specific decoding for secret names.
// Reverses EncodeAzureSecretName.
func DecodeAzureSecretName(encodedSecretName string) (string, error) {
	parts := strings.Split(encodedSecretName, azureSecretNameComponentDelimiter)
	if len(parts) != 5 || parts[0] != fmt.Sprint(azureEncodingVersion) {
		return EmptyString, fmt.Errorf("invalid secret name format or version")
	}

	nameBytes, err := base32Encoding.DecodeString(parts[4])
	if err != nil {
		return EmptyString, fmt.Errorf("error decoding name component: %w", err)
	}

	decode := func(s string) (*string, error) {
		var result strings.Builder
		for i := 0; i < len(s); i++ {
			if s[i] == '0' && i+1 < len(s) {
				if s[i+1] == '0' {
					result.WriteByte('0')
				} else if s[i+1] == azureHyphenMarker {
					result.WriteByte('-')
				} else {
					return nil, fmt.Errorf("unexpected escape sequence: %c%c", s[i], s[i+1])
				}
				i++
			} else {
				result.WriteByte(s[i])
			}
		}
		return to.Ptr(result.String()), nil
	}

	org, err := decode(parts[1])
	if err != nil {
		return EmptyString, fmt.Errorf("error decoding org from stored secret %s: %w", encodedSecretName, err)
	}

	domain, err := decode(parts[2])
	if err != nil {
		return EmptyString, fmt.Errorf("error decoding domain from stored secret %s: %w", encodedSecretName, err)
	}

	project, err := decode(parts[3])
	if err != nil {
		return EmptyString, fmt.Errorf("error decoding project from stored secret %s: %w", encodedSecretName, err)
	}

	// Re-encode to Union naming format
	return EncodeSecretName(*org, *domain, *project, string(nameBytes)), nil
}

// Encodes a string value to appropriate Json format for Azure Key Vault storage
func StringToAzureSecret(value string) (string, error) {
	return marshalAzureSecretValue(azureSecretValueTypeSTRING, value)
}

// Encodes a binary value to appropriate Json format for Azure Key Vault storage
func BinaryToAzureSecret(value []byte) (string, error) {
	return marshalAzureSecretValue(azureSecretValueTypeBINARY, base64.StdEncoding.EncodeToString(value))
}

func AzureToUnionSecret(secret azsecrets.Secret) (*SecretValue, error) {
	var sv azureSecretValue
	err := json.Unmarshal([]byte(*secret.Value), &sv)
	if err != nil {
		return nil, err
	}

	if sv.Type == azureSecretValueTypeBINARY {
		bytes, err := base64.StdEncoding.DecodeString(sv.Value)
		if err != nil {
			return nil, err
		}
		return &SecretValue{BinaryValue: bytes}, nil
	}

	return &SecretValue{StringValue: sv.Value}, nil
}

// The AzureErrorResponse object represents an error response from Azure Key Vault API.
//
// Ref: https://learn.microsoft.com/en-us/azure/key-vault/general/authentication-requests-and-responses
type AzureErrorResponse struct {
	Error AzureError `json:"error"`
}

type AzureError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

//go:generate enumer --type=azureSecretValueType --trimprefix=azureSecretValueType -json -yaml

type azureSecretValueType uint8

const (
	azureSecretValueTypeSTRING azureSecretValueType = iota
	azureSecretValueTypeBINARY
)

// A Flyte specific encoding to store strings and binary data as secrets in Azure Key Vault.
// Azure only supports strings and this allows us to store binary data as base64 encoded byte slice.
type azureSecretValue struct {
	Type azureSecretValueType `json:"type"`
	// utf-8 string if type is STRING, based64 encoded bytes if type is BINARY
	Value string `json:"value"`
}

func marshalAzureSecretValue(t azureSecretValueType, value string) (string, error) {
	bytes, err := json.Marshal(azureSecretValue{
		Type:  t,
		Value: value,
	})
	if err != nil {
		return EmptyString, err
	}
	return string(bytes), nil
}
