package impl

import (
	"context"
	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/impl/validation"
	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/interfaces"
	repositoryInterfaces "github.com/flyteorg/flyte/flyteadmin/pkg/repositories/interfaces"
	runtimeInterfaces "github.com/flyteorg/flyte/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flytestdlib/storage"
)

type OverrideAttributesManager struct {
	db            repositoryInterfaces.Repository
	config        runtimeInterfaces.ApplicationConfiguration
	storageClient *storage.DataStore
}

func (m *OverrideAttributesManager) GetOverrideAttributes(
	ctx context.Context, request admin.OverrideAttributesGetRequest) (
	*admin.OverrideAttributesGetResponse, error) {
	if err := validation.ValidateOverrideAttributesRequest(request); err != nil {
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
		GlobalAttributes:        getSpecificDocument(document, request.Id.Org, "", ""),
		ProjectAttributes:       getSpecificDocument(document, request.Id.Org, request.Id.Domain, ""),
		ProjectDomainAttributes: getSpecificDocument(document, request.Id.Org, request.Id.Domain, request.Id.Project),
	}, nil
}

func (m *OverrideAttributesManager) UpdateOverrideAttributes(
	ctx context.Context, request admin.OverrideAttributesUpdateRequest) (
	*admin.OverrideAttributesUpdateResponse, error) {
	if err := validation.ValidateOverrideAttributesUpdateRequest(request); err != nil {
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
	updateDocument(document, request)
	return &admin.OverrideAttributesUpdateResponse{}, nil
}

func updateDocument(document *admin.Document, request admin.OverrideAttributesUpdateRequest) {
	if request.Id.Project != "" && request.Id.Domain != "" {
		assignSpecificDocument(document, request.Id.Org, request.Id.Domain, request.Id.Project, request.Attribute)
	}
}

func getSpecificDocument(document *admin.Document, org, domain, project string) *admin.Attributes {
	if orgDocument, ok := document.OrgDocuments[org]; ok {
		if domainDocument, ok := orgDocument.ProjectDocuments[domain]; ok {
			if projectDocument, ok := domainDocument.ProjectDomainDocuments[project]; ok {
				return projectDocument.WorkflowDocuments[""].LaunchPlanDocuments[""].Attributes
			}
		}
	}
	return nil
}

func assignSpecificDocument(document *admin.Document, org, domain, project string, attributes *admin.Attributes) {
	if orgDocument, ok := document.OrgDocuments[org]; ok {
		if domainDocument, ok := orgDocument.ProjectDocuments[domain]; ok {
			if projectDocument, ok := domainDocument.ProjectDomainDocuments[project]; ok {
				projectDocument.WorkflowDocuments[""].LaunchPlanDocuments[""].Attributes = attributes
			}
		}
	}
}

func NewOverrideAttributesManager(db repositoryInterfaces.Repository, config runtimeInterfaces.ApplicationConfiguration) interfaces.OverrideAttributesInterface {
	return &OverrideAttributesManager{
		db:     db,
		config: config,
	}
}
