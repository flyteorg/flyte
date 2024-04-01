package impl

import (
	"context"
	"github.com/flyteorg/flyte/flyteadmin/pkg/common"
	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/impl/validation"
	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/interfaces"
	repositoryInterfaces "github.com/flyteorg/flyte/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/models"
	runtimeInterfaces "github.com/flyteorg/flyte/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flytestdlib/storage"
	"time"
)

type OverrideAttributesManager struct {
	db            repositoryInterfaces.Repository
	config        runtimeInterfaces.ApplicationConfiguration
	storageClient *storage.DataStore
}

func (m *OverrideAttributesManager) GetOverrideAttributes(
	ctx context.Context, request admin.OverrideAttributesGetRequest) (
	*admin.OverrideAttributesGetResponse, error) {
	if err := validation.ValidateOverrideAttributesGetRequest(request); err != nil {
		return nil, err
	}
	attributes, err := m.db.OverrideAttributesRepo().GetActive(ctx)
	if err != nil {
		return nil, err
	}
	document := &admin.Document{}
	if err := m.storageClient.ReadProtobuf(ctx, attributes.DocumentLocation, document); err != nil {
		return nil, err
	}
	return &admin.OverrideAttributesGetResponse{
		GlobalAttributes:        getAttributeOfDocument(document, request.Id.Org, "", ""),
		ProjectAttributes:       getAttributeOfDocument(document, request.Id.Org, request.Id.Project, ""),
		ProjectDomainAttributes: getAttributeOfDocument(document, request.Id.Org, request.Id.Project, request.Id.Domain),
	}, nil
}

func (m *OverrideAttributesManager) UpdateOverrideAttributes(
	ctx context.Context, request admin.OverrideAttributesUpdateRequest) (
	*admin.OverrideAttributesUpdateResponse, error) {
	if err := validation.ValidateOverrideAttributesUpdateRequest(request); err != nil {
		return nil, err
	}

	// Get the active override attributes
	attributes, err := m.db.OverrideAttributesRepo().GetActive(ctx)
	if err != nil {
		return nil, err

	}
	document := &admin.Document{}
	if err := m.storageClient.ReadProtobuf(ctx, attributes.DocumentLocation, document); err != nil {
		return nil, err
	}

	// Update the document with the new attributes
	updateAttributeOfDocument(document, request.Id.Org, request.Id.Project, request.Id.Domain, request.Attributes)

	// Offload the updated document
	updatedDocumentLocation, err := common.OffloadOverrideAttributesDocument(ctx, m.storageClient, document)
	createOverrideAttributesInput := models.OverrideAttributes{
		Version:          time.Now(),
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

func getAttributeOfDocument(document *admin.Document, org, project, domain string) *admin.Attributes {
	documentKey := encodeDocumentKey(org, project, domain, "", "")
	if attributes, ok := document.AttributesMap[documentKey]; ok {
		return attributes
	}
	return nil
}

func updateAttributeOfDocument(document *admin.Document, org, project, domain string, attributes *admin.Attributes) {
	documentKey := encodeDocumentKey(org, project, domain, "", "")
	document.AttributesMap[documentKey] = attributes

}

func encodeDocumentKey(org, project, domain, workflow, launch_plan string) string {
	return org + "/" + project + "/" + domain + "/" + workflow + "/" + launch_plan
}

func NewOverrideAttributesManager(db repositoryInterfaces.Repository, config runtimeInterfaces.ApplicationConfiguration) interfaces.OverrideAttributesInterface {
	return &OverrideAttributesManager{
		db:     db,
		config: config,
	}
}
