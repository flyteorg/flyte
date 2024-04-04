package impl

import (
	"context"
	"crypto/rand"
	"github.com/flyteorg/flyte/flyteadmin/pkg/common"
	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/impl/validation"
	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/interfaces"
	repositoryInterfaces "github.com/flyteorg/flyte/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/models"
	runtimeInterfaces "github.com/flyteorg/flyte/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flytestdlib/storage"
	"math/big"
)

type ConfigurationManager struct {
	db            repositoryInterfaces.Repository
	config        runtimeInterfaces.Configuration
	storageClient *storage.DataStore
	cache         *admin.ConfigurationDocument
}

func (m *ConfigurationManager) GetConfiguration(
	ctx context.Context, request admin.ConfigurationGetRequest) (
	*admin.ConfigurationGetResponse, error) {
	// Validate the request
	if err := validation.ValidateConfigurationGetRequest(request); err != nil {
		return nil, err
	}

	// Get the active override attributes
	activeConfiguration, err := m.db.ConfigurationRepo().GetActive(ctx)
	if err != nil {
		return nil, err
	}

	// If the cache is not nil and the version of the cache is the same as the active override attributes, return the cache
	if m.cache != nil && m.cache.GetVersion() == activeConfiguration.Version {
		return &admin.ConfigurationGetResponse{
			Id: &admin.ProjectID{
				Project: request.GetId().GetProject(),
				Domain:  request.GetId().GetDomain(),
			},
			Version:                    m.cache.GetVersion(),
			GlobalConfiguration:        getConfigurationFromDocument(m.cache, "", ""),
			ProjectConfiguration:       getConfigurationFromDocument(m.cache, request.Id.Project, ""),
			ProjectDomainConfiguration: getConfigurationFromDocument(m.cache, request.Id.Project, request.Id.Domain),
		}, nil
	}

	// TODO: Handle the case where the document location is empty
	document := &admin.ConfigurationDocument{}
	if err := m.storageClient.ReadProtobuf(ctx, activeConfiguration.DocumentLocation, document); err != nil {
		return nil, err
	}

	// Cache the document
	m.cache = document

	// Return the override attributes of different scopes
	return &admin.ConfigurationGetResponse{
		Id: &admin.ProjectID{
			Project: request.GetId().GetProject(),
			Domain:  request.GetId().GetDomain(),
		},
		Version:                    document.GetVersion(),
		GlobalConfiguration:        getConfigurationFromDocument(document, "", ""),
		ProjectConfiguration:       getConfigurationFromDocument(document, request.Id.Project, ""),
		ProjectDomainConfiguration: getConfigurationFromDocument(document, request.Id.Project, request.Id.Domain),
	}, nil
}

func (m *ConfigurationManager) UpdateConfiguration(
	ctx context.Context, request admin.ConfigurationUpdateRequest) (
	*admin.ConfigurationUpdateResponse, error) {
	// Validate the request
	if err := validation.ValidateConfigurationUpdateRequest(request); err != nil {
		return nil, err
	}

	// Get the active override attributes
	activeConfiguration, err := m.db.ConfigurationRepo().GetActive(ctx)
	if err != nil {
		return nil, err
	}

	// TODO: Handle the case where the document location is empty
	document := &admin.ConfigurationDocument{}
	if err := m.storageClient.ReadProtobuf(ctx, activeConfiguration.DocumentLocation, document); err != nil {
		return nil, err
	}

	// Update the document with the new attributes
	updateConfigurationToDocument(document, request.Id.Project, request.Id.Domain, request.Configuration)

	// Generate a new version for the document
	generatedVersion, err := GenerateRandomString(10)
	if err != nil {
		return nil, err
	}
	document.Version = generatedVersion

	// Offload the updated document
	updatedDocumentLocation, err := common.OffloadConfigurationDocument(ctx, m.storageClient, document, generatedVersion)

	newConfiguration := models.Configuration{
		// random generate a string as version
		Version:          generatedVersion,
		Active:           true,
		DocumentLocation: updatedDocumentLocation,
	}

	// Erase the active override attributes and create the new one
	var versionToUpdate string
	if request.VersionToUpdate != "" {
		versionToUpdate = request.VersionToUpdate
	} else {
		versionToUpdate = activeConfiguration.Version
	}
	if err := m.db.ConfigurationRepo().EraseActiveAndCreate(ctx, versionToUpdate, newConfiguration); err != nil {
		return nil, err
	}
	return &admin.ConfigurationUpdateResponse{}, nil
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
func getConfigurationFromDocument(document *admin.ConfigurationDocument, project, domain string) *admin.Configuration {
	documentKey := encodeDocumentKey(project, domain, "")
	if attributes, ok := document.Configurations[documentKey]; ok {
		return attributes
	}
	return nil
}

// This function is used to update the attributes of a document based on the project, and domain.
func updateConfigurationToDocument(document *admin.ConfigurationDocument, project, domain string, attributes *admin.Configuration) {
	documentKey := encodeDocumentKey(project, domain, "")
	document.Configurations[documentKey] = attributes
}

// This function is used to encode the document key based on the org, project, domain, workflow, and launch plan.
func encodeDocumentKey(project, domain, workflow string) string {
	// TODO: This is a temporary solution to encode the document key. We need to come up with a better solution.
	return project + "/" + domain + "/" + workflow
}

func NewConfigurationManager(db repositoryInterfaces.Repository, config runtimeInterfaces.Configuration, storageClient *storage.DataStore) interfaces.ConfigurationInterface {
	return &ConfigurationManager{
		db:            db,
		config:        config,
		storageClient: storageClient,
		cache:         nil,
	}
}
