package secret

import (
	"encoding/base64"
	"fmt"
	"math"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"
	"github.com/stretchr/testify/assert"
)

const (
	// Including -, 0, and 1 to tests encoding and decoding escape functionality in addition to specific names
	validSecretIDFmt   = "u__org__testorg__domain__testdomain__project__testproject__key__%s"                            // #nosec G101
	longSecretIDPrefix = "u__org__test-org-with-long-name__domain__development__project__union-health-monitoring__key__" // #nosec G101
)

var azureEncodingDecodingVars = []struct {
	name    string
	decoded string
	encoded string
}{
	{
		name:    "test name without hyphens, escape characters, nor hyphen markers",
		decoded: "u__org__testorg__domain__testdomain__project__testproject__key__testsecret",
		encoded: "1-testorg-testdomain-testproject-ORSXG5DTMVRXEZLU",
	},
	{
		name:    "test name with hyphens",
		decoded: "u__org__test-org__domain__test--domain__project__--testproject--__key__test-secret--",
		encoded: "1-test01org-test0101domain-0101testproject0101-ORSXG5BNONSWG4TFOQWS2",
	},
	{
		name:    "test name with escape characters",
		decoded: "u__org__test0org__domain__test00domain__project__00testproject00__key__testsecret",
		encoded: "1-test00org-test0000domain-0000testproject0000-ORSXG5DTMVRXEZLU",
	},
	{
		name:    "test name that looks escaped",
		decoded: "u__org__test01org__domain__test0102domain__project__01testproject0101__key__testsecret",
		encoded: "1-test001org-test001002domain-001testproject001001-ORSXG5DTMVRXEZLU",
	},
	{
		name:    "test secret name that has invalid Azure characters",
		decoded: "u__org__test01org__domain__test0101domain__project__01testproject0101__key__test_secret__",
		encoded: "1-test001org-test001001domain-001testproject001001-ORSXG5C7ONSWG4TFORPV6",
	},
	{
		name:    "test secret name with capital letters",
		decoded: "u__org__test01org__domain__test0101domain__project__01testproject0101__key__TESTSECRET",
		encoded: "1-test001org-test001001domain-001testproject001001-KRCVGVCTIVBVERKU",
	},
}

func Test_EncodeAzureSecretName(t *testing.T) {
	for _, tt := range azureEncodingDecodingVars {
		t.Run(tt.name, func(t *testing.T) {
			got, err := EncodeAzureSecretName(tt.decoded)
			assert.NoError(t, err)
			assert.Equal(t, tt.encoded, got)
		})
	}

	t.Run("secret name is too long", func(t *testing.T) {
		encodedPrefix, err := EncodeAzureSecretName(longSecretIDPrefix)
		assert.NoError(t, err)
		spaceRemaining := azureMaxSecretNameLength - len(encodedPrefix)
		maxLength := int(math.Floor(float64(spaceRemaining) * 5 / 8))

		b := make([]byte, maxLength+1)
		for i := range b {
			b[i] = 'a'
		}
		_, err = EncodeAzureSecretName(longSecretIDPrefix + string(b))
		assert.Error(t, err)
	})

	t.Run("supports case insensitive secret name", func(t *testing.T) {
		lower, err := EncodeAzureSecretName(fmt.Sprintf(validSecretIDFmt, "testsecret"))
		assert.NoError(t, err)
		upper, err := EncodeAzureSecretName(fmt.Sprintf(validSecretIDFmt, "TESTSECRET"))
		assert.NoError(t, err)
		assert.NotEqual(t, lower, upper)
	})
}

func Test_DecodeAzureSecretName(t *testing.T) {
	for _, tt := range azureEncodingDecodingVars {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DecodeAzureSecretName(tt.encoded)
			assert.NoError(t, err)
			assert.Equal(t, tt.decoded, got)
		})
	}

	t.Run("invalid secret name version", func(t *testing.T) {
		_, err := DecodeAzureSecretName("2-test01org00-test01domain0000-test01project001-testsecret")
		assert.Error(t, err)
	})

	t.Run("incorrect number of parts", func(t *testing.T) {
		_, err := DecodeAzureSecretName("1-test01org00-test01domain0000test01project001-testsecret")
		assert.Error(t, err)
	})

	t.Run("org has unexpected escaped character", func(t *testing.T) {
		invalidSecretName := "1-test02org-testdomain-testproject-ORSXG5DTMVRXEZLU"
		_, err := DecodeAzureSecretName(invalidSecretName)
		assert.Error(t, err)
		assert.EqualError(t, err, fmt.Sprintf("error decoding org from stored secret %s: unexpected escape sequence: 02", invalidSecretName))
	})

	t.Run("domain has unexpected escaped character", func(t *testing.T) {
		invalidSecretName := "1-testorg-test02domain-testproject-ORSXG5DTMVRXEZLU" // #nosec G101
		_, err := DecodeAzureSecretName(invalidSecretName)
		assert.Error(t, err)
		assert.EqualError(t, err, fmt.Sprintf("error decoding domain from stored secret %s: unexpected escape sequence: 02", invalidSecretName))
	})

	t.Run("project has unexpected escaped character", func(t *testing.T) {
		invalidSecretName := "1-testorg-testdomain-test02project-ORSXG5DTMVRXEZLU"
		_, err := DecodeAzureSecretName(invalidSecretName)
		assert.Error(t, err)
		assert.EqualError(t, err, fmt.Sprintf("error decoding project from stored secret %s: unexpected escape sequence: 02", invalidSecretName))
	})
}

func Test_EncodeDecodeAzureSecretName_Bijectivity(t *testing.T) {
	for _, tt := range azureEncodingDecodingVars {
		t.Run(tt.name, func(t *testing.T) {
			encoded, err := EncodeAzureSecretName(tt.decoded)
			assert.NoError(t, err)
			decoded, err := DecodeAzureSecretName(encoded)
			assert.NoError(t, err)
			assert.Equal(t, tt.decoded, decoded)
		})
	}
}

func Test_StringToAzureSecret(t *testing.T) {
	tests := []struct {
		name string
		arg  string
	}{
		{
			"empty string",
			"",
		},
		{
			"test-value",
			"test-value",
		},
		{
			"value with spaces",
			" test value ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := StringToAzureSecret(tt.arg)
			assert.NoError(t, err)
			assert.Equal(t, fmt.Sprintf(`{"type":"STRING","value":"%v"}`, tt.arg), res)
			// Test bijective property
			decoded, err := AzureToUnionSecret(azsecrets.Secret{Value: &res})
			assert.NoError(t, err)
			assert.Equal(t, tt.arg, decoded.StringValue)
		})
	}
}

func Test_BinaryToAzureSecret(t *testing.T) {
	tests := []struct {
		name string
		arg  []byte
	}{
		{
			"empty",
			[]byte{},
		},
		{
			"single element",
			[]byte{0x48},
		},
		{
			"non empty",
			[]byte{0x48, 0x65, 0x6C, 0x6C, 0x6F},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := BinaryToAzureSecret(tt.arg)
			assert.NoError(t, err)
			encodedBinary := base64.StdEncoding.EncodeToString(tt.arg)
			assert.Equal(t, fmt.Sprintf(`{"type":"BINARY","value":"%v"}`, encodedBinary), res)
			// Test bijective property
			decoded, err := AzureToUnionSecret(azsecrets.Secret{Value: &res})
			assert.NoError(t, err)
			assert.Equal(t, tt.arg, decoded.BinaryValue)
		})
	}
}

func Test_AzureToUnionSecret(t *testing.T) {
	t.Run("successful string translation", func(t *testing.T) {
		azSecretValue := `{"type":"STRING","value":"test-value"}` //nolint:gosec
		azSecret := azsecrets.Secret{Value: &azSecretValue}
		got, err := AzureToUnionSecret(azSecret)
		assert.NoError(t, err)
		assert.Equal(t, SecretValue{StringValue: "test-value"}, *got)
	})

	t.Run("successful byte translation", func(t *testing.T) {
		expectedSecretValue := []byte{0x48, 0x65, 0x6C, 0x6C, 0x6F}
		azSecretValueFmt := `{"type":"BINARY","value":"%v"}` //nolint:gosec
		azSecretValue := fmt.Sprintf(azSecretValueFmt, base64.StdEncoding.EncodeToString(expectedSecretValue))
		azSecret := azsecrets.Secret{Value: &azSecretValue}
		got, err := AzureToUnionSecret(azSecret)
		assert.NoError(t, err)
		assert.Equal(t, SecretValue{BinaryValue: expectedSecretValue}, *got)
	})

	t.Run("invalid json", func(t *testing.T) {
		azSecretValue := `{"type":"STRING","value":"test-value"` //nolint:gosec
		azSecret := azsecrets.Secret{Value: &azSecretValue}
		_, err := AzureToUnionSecret(azSecret)
		assert.Error(t, err)
	})

	t.Run("non valid base64 encoded binary", func(t *testing.T) {
		azSecretValue := `{"type":"BINARY","value":"test-value"}` //nolint:gosec
		azSecret := azsecrets.Secret{Value: &azSecretValue}
		_, err := AzureToUnionSecret(azSecret)
		assert.Error(t, err)
	})
}
