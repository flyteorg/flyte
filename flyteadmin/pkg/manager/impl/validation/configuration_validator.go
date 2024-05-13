package validation

import (
	"context"

	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/impl/shared"
	repositoryInterfaces "github.com/flyteorg/flyte/flyteadmin/pkg/repositories/interfaces"
	runtimeInterfaces "github.com/flyteorg/flyte/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
)

func ValidateConfigurationGetRequest(ctx context.Context, db repositoryInterfaces.Repository, config runtimeInterfaces.ApplicationConfiguration, request admin.ConfigurationGetRequest) error {
	if request.Id == nil {
		return shared.GetMissingArgumentError(shared.ID)
	}
	// Get request should only have org (optional), project, and domain set
	if err := ValidateNonemptyStringField(request.Id.Workflow, shared.Workflow); err != nil {
		return err
	}
	if err := ValidateProjectExists(ctx, db, request.Id.Project, request.Id.Org); err != nil {
		return err
	}
	return ValidateDomainExists(ctx, config, request.Id.Domain)
}

func ValidateConfigurationUpdateRequest(request admin.ConfigurationUpdateRequest) error {
	if request.Id == nil {
		return shared.GetMissingArgumentError(shared.ID)
	}
	if request.Id.Workflow != "" && (request.Id.Project == "" || request.Id.Domain == "") {
		return shared.GetInvalidArgumentError(shared.ID)
	}
	if request.Id.Domain != "" && request.Id.Project == "" {
		return shared.GetInvalidArgumentError(shared.ID)
	}
	if request.Configuration == nil {
		return shared.GetMissingArgumentError(shared.Configuration)
	}
	return nil
}

func ValidateWorkflowConfigurationUpdateRequest(ctx context.Context, db repositoryInterfaces.Repository, config runtimeInterfaces.ApplicationConfiguration, request admin.ConfigurationUpdateRequest) error {
	if request.Id == nil {
		return shared.GetMissingArgumentError(shared.ID)
	}
	if err := ValidateEmptyStringField(request.Id.Workflow, shared.Workflow); err != nil {
		return err
	}
	if err := ValidateProjectAndDomain(ctx, db, config, request.Id.Project, request.Id.Domain, request.Id.Org); err != nil {
		return err
	}
	if request.Configuration == nil {
		return shared.GetMissingArgumentError(shared.Configuration)
	}
	return nil
}

func ValidateProjectDomainConfigurationUpdateRequest(ctx context.Context, db repositoryInterfaces.Repository, config runtimeInterfaces.ApplicationConfiguration, request admin.ConfigurationUpdateRequest) error {
	if request.Id == nil {
		return shared.GetMissingArgumentError(shared.ID)
	}
	if err := ValidateNonemptyStringField(request.Id.Workflow, shared.Workflow); err != nil {
		return err
	}
	if err := ValidateProjectAndDomain(ctx, db, config, request.Id.Project, request.Id.Domain, request.Id.Org); err != nil {
		return err
	}
	if request.Configuration == nil {
		return shared.GetMissingArgumentError(shared.Configuration)
	}
	return nil
}

func ValidateProjectConfigurationUpdateRequest(ctx context.Context, db repositoryInterfaces.Repository, request admin.ConfigurationUpdateRequest) error {
	if request.Id == nil {
		return shared.GetMissingArgumentError(shared.ID)
	}
	if err := ValidateNonemptyStringField(request.Id.Workflow, shared.Workflow); err != nil {
		return err
	}
	if err := ValidateNonemptyStringField(request.Id.Domain, shared.Domain); err != nil {
		return err
	}
	if err := ValidateProjectExistsAndActive(ctx, db, request.Id.Project, request.Id.Org); err != nil {
		return err
	}
	if request.Configuration == nil {
		return shared.GetMissingArgumentError(shared.Configuration)
	}
	return nil
}
