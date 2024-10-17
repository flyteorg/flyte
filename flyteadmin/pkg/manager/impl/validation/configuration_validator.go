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
	return v.validateConfigurationID(ctx, request.Id)
}

func (v ConfigurationValidator) ValidateUpdateRequest(ctx context.Context, request admin.ConfigurationUpdateRequest) error {
	if request.Id == nil {
		return shared.GetMissingArgumentError(shared.ID)
	}
	if request.Configuration == nil {
		return shared.GetMissingArgumentError(shared.Configuration)
	}
	return v.validateConfigurationID(ctx, request.Id)
}

func (v ConfigurationValidator) validateConfigurationID(ctx context.Context, configurationID *admin.ConfigurationID) error {
	// If it's a workflow configuration update
	if configurationID.Workflow != "" {
		return v.validateConfigurationIDWorkflow(ctx, configurationID)
	}
	// If it's a project domain configuration update
	if configurationID.Project != "" && configurationID.Domain != "" {
		return v.validateConfigurationIDProjectDomain(ctx, configurationID)
	}
	// If it's a project configuration update
	if configurationID.Project != "" {
		return v.validateConfigurationIDProject(ctx, configurationID)
	}
	// Temporarily allow this to let FE split their changes
	if configurationID.Domain != "" {
		return nil
	}
	// Assume it's an org configuration update
	return v.validateConfigurationIDOrg(configurationID)
}

func (v ConfigurationValidator) validateConfigurationIDWorkflow(ctx context.Context, configurationID *admin.ConfigurationID) error {
	if configurationID == nil {
		return shared.GetMissingArgumentError(shared.ID)
	}
	if err := ValidateEmptyStringField(configurationID.Workflow, shared.Workflow); err != nil {
		return err
	}
	if err := ValidateProjectAndDomain(ctx, v.db, v.config, configurationID.Project, configurationID.Domain, configurationID.Org); err != nil {
		return err
	}
	return nil
}

func (v ConfigurationValidator) validateConfigurationIDProjectDomain(ctx context.Context, configurationID *admin.ConfigurationID) error {
	if configurationID == nil {
		return shared.GetMissingArgumentError(shared.ID)
	}
	if err := ValidateNonemptyStringField(configurationID.Workflow, shared.Workflow); err != nil {
		return err
	}
	if err := ValidateProjectAndDomain(ctx, v.db, v.config, configurationID.Project, configurationID.Domain, configurationID.Org); err != nil {
		return err
	}
	return nil
}

func (v ConfigurationValidator) validateConfigurationIDProject(ctx context.Context, configurationID *admin.ConfigurationID) error {
	if configurationID == nil {
		return shared.GetMissingArgumentError(shared.ID)
	}
	if err := ValidateNonemptyStringField(configurationID.Workflow, shared.Workflow); err != nil {
		return err
	}
	if err := ValidateNonemptyStringField(configurationID.Domain, shared.Domain); err != nil {
		return err
	}
	if err := ValidateProjectExistsAndActive(ctx, v.db, configurationID.Project, configurationID.Org); err != nil {
		return err
	}
	return nil
}

func (v ConfigurationValidator) validateConfigurationIDOrg(configurationID *admin.ConfigurationID) error {
	if configurationID == nil {
		return shared.GetMissingArgumentError(shared.ID)
	}
	if err := ValidateNonemptyStringField(configurationID.Workflow, shared.Workflow); err != nil {
		return err
	}
	if err := ValidateNonemptyStringField(configurationID.Domain, shared.Domain); err != nil {
		return err
	}
	if err := ValidateNonemptyStringField(configurationID.Project, shared.Project); err != nil {
		return err
	}
	return nil
}

func NewConfigurationValidator(db repositoryInterfaces.Repository, config runtimeInterfaces.ApplicationConfiguration) ConfigurationValidator {
	return ConfigurationValidator{
		db:     db,
		config: config,
	}
}
