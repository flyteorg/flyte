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

type OverrideAttributesManager struct {
	db            repositoryInterfaces.Repository
	config        runtimeInterfaces.Configuration
	storageClient *storage.DataStore
}

func (m *OverrideAttributesManager) GetOverrideAttributes(
	ctx context.Context, request admin.OverrideAttributesGetRequest) (
	*admin.OverrideAttributesGetResponse, error) {
	// Validate the request
	if err := validation.ValidateOverrideAttributesGetRequest(request); err != nil {
		return nil, err
	}

	// Get the active override attributes
	overrideAttributes, err := m.db.OverrideAttributesRepo().GetActive(ctx)
	if err != nil {
		return nil, err
	}

	// TODO: Handle the case where the document location is empty
	document := &admin.Document{}
	if err := m.storageClient.ReadProtobuf(ctx, overrideAttributes.DocumentLocation, document); err != nil {
		return nil, err
	}

	// Return the override attributes of different scopes
	return &admin.OverrideAttributesGetResponse{
		Id: &admin.ProjectID{
			Project: request.GetId().GetProject(),
			Domain:  request.GetId().GetDomain(),
		},
		Version:                 document.GetVersion(),
		GlobalAttributes:        getAttributeOfDocument(document, "", ""),
		ProjectAttributes:       getAttributeOfDocument(document, request.Id.Project, ""),
		ProjectDomainAttributes: getAttributeOfDocument(document, request.Id.Project, request.Id.Domain),
	}, nil
}

func (m *OverrideAttributesManager) UpdateOverrideAttributes(
	ctx context.Context, request admin.OverrideAttributesUpdateRequest) (
	*admin.OverrideAttributesUpdateResponse, error) {
	// TODO: These should be done in a transaction

	// Validate the request
	if err := validation.ValidateOverrideAttributesUpdateRequest(request); err != nil {
		return nil, err
	}

	// Get the active override attributes
	overrideAttributes, err := m.db.OverrideAttributesRepo().GetActive(ctx)
	if err != nil {
		return nil, err

	}

	// TODO: Handle the case where the document location is empty
	document := &admin.Document{}
	if err := m.storageClient.ReadProtobuf(ctx, overrideAttributes.DocumentLocation, document); err != nil {
		return nil, err
	}

	// Update the document with the new attributes
	updateAttributeOfDocument(document, request.Id.Project, request.Id.Domain, request.Attributes)

	// Generate a new version for the document
	generatedVersion, err := GenerateRandomString(10)
	if err != nil {
		return nil, err
	}
	document.Version = generatedVersion

	// Offload the updated document
	updatedDocumentLocation, err := common.OffloadOverrideAttributesDocument(ctx, m.storageClient, document, generatedVersion)

	createOverrideAttributesInput := models.OverrideAttributes{
		// random generate a string as version
		Version:          generatedVersion,
		Active:           true,
		DocumentLocation: updatedDocumentLocation,
	}

	// Erase the active override attributes and create the new one
	if err := m.db.OverrideAttributesRepo().EraseActive(ctx); err != nil {
		return nil, err
	}
	if err := m.db.OverrideAttributesRepo().Create(ctx, createOverrideAttributesInput); err != nil {
		return nil, err
	}
	return &admin.OverrideAttributesUpdateResponse{}, nil
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
func getAttributeOfDocument(document *admin.Document, project, domain string) *admin.Attributes {
	documentKey := encodeDocumentKey(project, domain, "")
	if attributes, ok := document.AttributesMap[documentKey]; ok {
		return attributes
	}
	return nil
}

// This function is used to update the attributes of a document based on the project, and domain.
func updateAttributeOfDocument(document *admin.Document, project, domain string, attributes *admin.Attributes) {
	documentKey := encodeDocumentKey(project, domain, "")
	document.AttributesMap[documentKey] = attributes
}

// This function is used to encode the document key based on the org, project, domain, workflow, and launch plan.
func encodeDocumentKey(project, domain, workflow string) string {
	// TODO: This is a temporary solution to encode the document key. We need to come up with a better solution.
	return project + "/" + domain + "/" + workflow
}

func NewOverrideAttributesManager(db repositoryInterfaces.Repository, config runtimeInterfaces.Configuration, storageClient *storage.DataStore) interfaces.OverrideAttributesInterface {
	return &OverrideAttributesManager{
		db:            db,
		config:        config,
		storageClient: storageClient,
	}
}
