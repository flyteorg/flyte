package configurations

import (
	"crypto/rand"
	"math/big"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
)

func GetGlobalConfigurationFromDocument(document *admin.ConfigurationDocument) *admin.Configuration {
	return getConfigurationFromDocument(document, "", "", "")
}

func GetProjectConfigurationFromDocument(document *admin.ConfigurationDocument, project string) *admin.Configuration {
	return getConfigurationFromDocument(document, project, "", "")
}

func GetProjectDomainConfigurationFromDocument(document *admin.ConfigurationDocument, project, domain string) *admin.Configuration {
	return getConfigurationFromDocument(document, project, domain, "")
}

func GetWorkflowConfigurationFromDocument(document *admin.ConfigurationDocument, project, domain, workflow string) *admin.Configuration {
	return getConfigurationFromDocument(document, project, domain, workflow)
}

// TODO: Check if this function is implemented already
func GenerateRandomString(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	var result string
	charsetLength := big.NewInt(int64(len(charset)))
	for i := 0; i < length; i++ {
		randomNumber, err := rand.Int(rand.Reader, charsetLength)
		if err != nil {
			return "", err // Return the error if there was a problem generating the random number.
		}
		result += string(charset[randomNumber.Int64()])
	}
	return result, nil
}

// This function is used to get the attributes of a document based on the project, and domain.
func getConfigurationFromDocument(document *admin.ConfigurationDocument, project, domain, workflow string) *admin.Configuration {
	documentKey := encodeDocumentKey(project, domain, workflow)
	if _, ok := document.Configurations[documentKey]; !ok {
		document.Configurations[documentKey] = &admin.Configuration{}
	}
	return document.Configurations[documentKey]
}

// This function is used to update the attributes of a document based on the project, and domain.
func updateConfigurationToDocument(document *admin.ConfigurationDocument, configuration *admin.Configuration, project, domain, workflow string) {
	documentKey := encodeDocumentKey(project, domain, workflow)
	document.Configurations[documentKey] = configuration
}

// This function is used to encode the document key based on the org, project, domain, workflow, and launch plan.
func encodeDocumentKey(project, domain, workflow string) string {
	// TODO: This is a temporary solution to encode the document key. We need to come up with a better solution.
	return project + "/" + domain + "/" + workflow
}
