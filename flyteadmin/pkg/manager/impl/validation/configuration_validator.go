package validation

import (
	"context"

	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/impl/shared"
	repositoryInterfaces "github.com/flyteorg/flyte/flyteadmin/pkg/repositories/interfaces"
	runtimeInterfaces "github.com/flyteorg/flyte/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
)

type ConfigurationValidator struct {
	db     repositoryInterfaces.Repository
	config runtimeInterfaces.ApplicationConfiguration
}

func (v ConfigurationValidator) ValidateGetRequest(ctx context.Context, request admin.ConfigurationGetRequest) error {
	if request.Id == nil {
		return shared.GetMissingArgumentError(shared.ID)
	}
	return nil
}

func (v ConfigurationValidator) ValidateProjectDomainGetRequest(ctx context.Context, request admin.ConfigurationGetRequest) error {
	// Get project domain configuration request should only have org (optional), project, and domain set
	if err := ValidateNonemptyStringField(request.Id.Workflow, shared.Workflow); err != nil {
		return err
	}
	if err := ValidateProjectExists(ctx, v.db, request.Id.Project, request.Id.Org); err != nil {
		return err
	}
	return ValidateDomainExists(ctx, v.config, request.Id.Domain)
}

func (v ConfigurationValidator) ValidateDefaultGetRequest(ctx context.Context, request admin.ConfigurationGetRequest) error {
	// Only when the request id exists and only has org (optional) and domain set, this validation would be applicable.
	// So, we only have to check if the domain exists.
	return ValidateDomainExists(ctx, v.config, request.Id.Domain)
}

func (v ConfigurationValidator) ValidateUpdateRequest(ctx context.Context, request admin.ConfigurationUpdateRequest) error {
	if request.Id == nil {
		return shared.GetMissingArgumentError(shared.ID)
	}
	if request.Configuration == nil {
		return shared.GetMissingArgumentError(shared.Configuration)
	}
	// If it's a workflow configuration update
	if request.Id.Workflow != "" {
		return v.validateWorkflowUpdateRequest(ctx, request)
	}
	// If it's a project domain configuration update
	if request.Id.Project != "" && request.Id.Domain != "" {
		return v.validateProjectDomainUpdateRequest(ctx, request)
	}
	// If it's a project configuration update
	if request.Id.Project != "" {
		return v.validateProjectUpdateRequest(ctx, request)
	}
	// Assume it's an org configuration update
	return v.validateOrgUpdateRequest(request)
}

func (v ConfigurationValidator) validateWorkflowUpdateRequest(ctx context.Context, request admin.ConfigurationUpdateRequest) error {
	if request.Id == nil {
		return shared.GetMissingArgumentError(shared.ID)
	}
	if err := ValidateEmptyStringField(request.Id.Workflow, shared.Workflow); err != nil {
		return err
	}
	if err := ValidateProjectAndDomain(ctx, v.db, v.config, request.Id.Project, request.Id.Domain, request.Id.Org); err != nil {
		return err
	}
	if request.Configuration == nil {
		return shared.GetMissingArgumentError(shared.Configuration)
	}
	return nil
}

func (v ConfigurationValidator) validateProjectDomainUpdateRequest(ctx context.Context, request admin.ConfigurationUpdateRequest) error {
	if request.Id == nil {
		return shared.GetMissingArgumentError(shared.ID)
	}
	if err := ValidateNonemptyStringField(request.Id.Workflow, shared.Workflow); err != nil {
		return err
	}
	if err := ValidateProjectAndDomain(ctx, v.db, v.config, request.Id.Project, request.Id.Domain, request.Id.Org); err != nil {
		return err
	}
	if request.Configuration == nil {
		return shared.GetMissingArgumentError(shared.Configuration)
	}
	return nil

}

func (v ConfigurationValidator) validateProjectUpdateRequest(ctx context.Context, request admin.ConfigurationUpdateRequest) error {
	if request.Id == nil {
		return shared.GetMissingArgumentError(shared.ID)
	}
	if err := ValidateNonemptyStringField(request.Id.Workflow, shared.Workflow); err != nil {
		return err
	}
	if err := ValidateNonemptyStringField(request.Id.Domain, shared.Domain); err != nil {
		return err
	}
	if err := ValidateProjectExistsAndActive(ctx, v.db, request.Id.Project, request.Id.Org); err != nil {
		return err
	}
	if request.Configuration == nil {
		return shared.GetMissingArgumentError(shared.Configuration)
	}
	return nil

}

func (v ConfigurationValidator) validateOrgUpdateRequest(request admin.ConfigurationUpdateRequest) error {
	if request.Id == nil {
		return shared.GetMissingArgumentError(shared.ID)
	}
	if err := ValidateNonemptyStringField(request.Id.Workflow, shared.Workflow); err != nil {
		return err
	}
	if err := ValidateNonemptyStringField(request.Id.Domain, shared.Domain); err != nil {
		return err
	}
	if err := ValidateNonemptyStringField(request.Id.Project, shared.Project); err != nil {
		return err
	}
	if request.Configuration == nil {
		return shared.GetMissingArgumentError(shared.Configuration)
	}
	return nil
}

func NewConfigurationValidator(db repositoryInterfaces.Repository, config runtimeInterfaces.ApplicationConfiguration) ConfigurationValidator {
	return ConfigurationValidator{
		db:     db,
		config: config,
	}
}
