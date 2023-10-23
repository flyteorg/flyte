package update

import (
	"context"
	"fmt"
	"os"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	sconfig "github.com/flyteorg/flytectl/cmd/config/subcommand"
	cmdCore "github.com/flyteorg/flytectl/cmd/core"
	cmdUtil "github.com/flyteorg/flytectl/pkg/commandutils"
	"github.com/flyteorg/flytectl/pkg/ext"
)

func DecorateAndUpdateMatchableAttr(
	ctx context.Context,
	cmdCtx cmdCore.CommandContext,
	project, domain, workflow string,
	resourceType admin.MatchableResource,
	attributeDecorator sconfig.MatchableAttributeDecorator,
	dryRun bool,
	force bool,
) error {
	if project == "" {
		return fmt.Errorf("project is required")
	}
	if domain == "" && workflow != "" {
		return fmt.Errorf("domain is required")
	}

	switch {
	case workflow != "":
		return updateWorkflowMatchableAttributes(ctx, cmdCtx, project, domain, workflow, resourceType, attributeDecorator, dryRun, force)
	case domain != "":
		return updateProjectDomainMatchableAttributes(ctx, cmdCtx, project, domain, resourceType, attributeDecorator, dryRun, force)
	default:
		return updateProjectMatchableAttributes(ctx, cmdCtx, project, resourceType, attributeDecorator, dryRun, force)
	}
}

func updateProjectMatchableAttributes(
	ctx context.Context,
	cmdCtx cmdCore.CommandContext,
	project string,
	resourceType admin.MatchableResource,
	attributeDecorator sconfig.MatchableAttributeDecorator,
	dryRun bool,
	force bool,
) error {
	if project == "" {
		panic("project is empty")
	}

	response, err := cmdCtx.AdminFetcherExt().FetchProjectAttributes(ctx, project, resourceType)
	if err != nil && !ext.IsNotFoundError(err) {
		return fmt.Errorf("update project %s matchable attributes: could not fetch attributes: %w", project, err)
	}

	oldMatchingAttributes := response.GetAttributes().GetMatchingAttributes()
	newMatchingAttributes := attributeDecorator.Decorate()

	if confirmed, err := confirmMatchableAttributeUpdate(oldMatchingAttributes, newMatchingAttributes, dryRun, force); err != nil || !confirmed {
		return err
	}

	if err := cmdCtx.AdminUpdaterExt().UpdateProjectAttributes(ctx, project, newMatchingAttributes); err != nil {
		return fmt.Errorf("update project %s matchable attributes: update failed: %w", project, err)
	}

	fmt.Printf("Updated attributes from %s project\n", project)
	return nil
}

func updateProjectDomainMatchableAttributes(
	ctx context.Context,
	cmdCtx cmdCore.CommandContext,
	project, domain string,
	resourceType admin.MatchableResource,
	attributeDecorator sconfig.MatchableAttributeDecorator,
	dryRun bool,
	force bool,
) error {
	if project == "" {
		panic("project is empty")
	}
	if domain == "" {
		panic("domain is empty")
	}

	response, err := cmdCtx.AdminFetcherExt().FetchProjectDomainAttributes(ctx, project, domain, resourceType)
	if err != nil && !ext.IsNotFoundError(err) {
		return fmt.Errorf("update project %s domain %s matchable attributes: could not fetch attributes: %w", project, domain, err)
	}

	oldMatchingAttributes := response.GetAttributes().GetMatchingAttributes()
	newMatchingAttributes := attributeDecorator.Decorate()

	if confirmed, err := confirmMatchableAttributeUpdate(oldMatchingAttributes, newMatchingAttributes, dryRun, force); err != nil || !confirmed {
		return err
	}

	if err := cmdCtx.AdminUpdaterExt().UpdateProjectDomainAttributes(ctx, project, domain, newMatchingAttributes); err != nil {
		return fmt.Errorf("update project %s domain %s matchable attributes: update failed: %w", project, domain, err)
	}

	fmt.Printf("Updated attributes from %s project and domain %s\n", project, domain)
	return nil
}

func updateWorkflowMatchableAttributes(
	ctx context.Context,
	cmdCtx cmdCore.CommandContext,
	project, domain, workflow string,
	resourceType admin.MatchableResource,
	attributeDecorator sconfig.MatchableAttributeDecorator,
	dryRun bool,
	force bool,
) error {
	if project == "" {
		panic("project is empty")
	}
	if domain == "" {
		panic("domain is empty")
	}
	if workflow == "" {
		panic("workflow is empty")
	}

	response, err := cmdCtx.AdminFetcherExt().FetchWorkflowAttributes(ctx, project, domain, workflow, resourceType)
	if err != nil && !ext.IsNotFoundError(err) {
		return fmt.Errorf("update project %s domain %s workflow %s matchable attributes: could not fetch attributes: %w", project, domain, workflow, err)
	}

	oldMatchingAttributes := response.GetAttributes().GetMatchingAttributes()
	newMatchingAttributes := attributeDecorator.Decorate()

	if confirmed, err := confirmMatchableAttributeUpdate(oldMatchingAttributes, newMatchingAttributes, dryRun, force); err != nil || !confirmed {
		return err
	}

	if err := cmdCtx.AdminUpdaterExt().UpdateWorkflowAttributes(ctx, project, domain, workflow, newMatchingAttributes); err != nil {
		return fmt.Errorf("update project %s domain %s workflow %s matchable attributes: update failed: %w", project, domain, workflow, err)
	}

	fmt.Printf("Updated attributes from %s project and domain %s and workflow %s\n", project, domain, workflow)
	return nil
}

func confirmMatchableAttributeUpdate(old, new *admin.MatchingAttributes, dryRun, force bool) (bool, error) {
	patch, err := DiffAsYaml(diffPathBefore, diffPathAfter, old.GetTarget(), new.GetTarget())
	if err != nil {
		return false, fmt.Errorf("update matchable attributes: %w", err)
	}

	if patch == "" {
		fmt.Printf("No changes detected. Skipping the update.\n")
		return false, nil
	}

	fmt.Printf("The following changes are to be applied.\n%s\n", patch)

	if dryRun {
		fmt.Printf("Skipping update request (dryRun)\n")
		return false, nil
	}

	if !force && !cmdUtil.AskForConfirmation("Continue?", os.Stdin) {
		return false, fmt.Errorf("update aborted by user")
	}

	return true, nil
}
