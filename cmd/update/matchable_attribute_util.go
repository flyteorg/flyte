package update

import (
	"context"

	sconfig "github.com/flyteorg/flytectl/cmd/config/subcommand"
	"github.com/flyteorg/flytectl/pkg/ext"
	"github.com/flyteorg/flytestdlib/logger"
)

func DecorateAndUpdateMatchableAttr(ctx context.Context, project, domain, workflowName string,
	updater ext.AdminUpdaterExtInterface, mcDecorator sconfig.MatchableAttributeDecorator, dryRun bool) error {
	matchingAttr := mcDecorator.Decorate()
	if len(workflowName) > 0 {
		// Update the workflow attribute using the admin.
		if dryRun {
			logger.Infof(ctx, "skipping UpdateWorkflowAttributes request (dryRun)")
		} else {
			err := updater.UpdateWorkflowAttributes(ctx, project, domain, workflowName, matchingAttr)
			if err != nil {
				return err
			}
		}
		logger.Debugf(ctx, "Updated attributes from %v project and domain %v and workflow %v", project, domain, workflowName)
	} else {
		// Update the project domain attribute using the admin.
		if dryRun {
			logger.Infof(ctx, "skipping UpdateProjectDomainAttributes request (dryRun)")
		} else {
			err := updater.UpdateProjectDomainAttributes(ctx, project, domain, matchingAttr)
			if err != nil {
				return err
			}
		}
		logger.Debugf(ctx, "Updated attributes from %v project and domain %v", project, domain)
	}
	return nil
}
