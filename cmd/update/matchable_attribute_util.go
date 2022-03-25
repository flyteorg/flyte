package update

import (
	"context"
	"fmt"

	sconfig "github.com/flyteorg/flytectl/cmd/config/subcommand"
	"github.com/flyteorg/flytectl/pkg/ext"
)

func DecorateAndUpdateMatchableAttr(ctx context.Context, project, domain, workflowName string,
	updater ext.AdminUpdaterExtInterface, mcDecorator sconfig.MatchableAttributeDecorator, dryRun bool) error {
	matchingAttr := mcDecorator.Decorate()
	if len(workflowName) > 0 {
		// Update the workflow attribute using the admin.
		if dryRun {
			fmt.Printf("skipping UpdateWorkflowAttributes request (dryRun)\n")
		} else {
			err := updater.UpdateWorkflowAttributes(ctx, project, domain, workflowName, matchingAttr)
			if err != nil {
				return err
			}
		}
		fmt.Printf("Updated attributes from %v project and domain %v and workflow %v\n", project, domain, workflowName)
	} else {
		// Update the project domain attribute using the admin.
		if dryRun {
			fmt.Printf("skipping UpdateProjectDomainAttributes request (dryRun)\n")
		} else {
			err := updater.UpdateProjectDomainAttributes(ctx, project, domain, matchingAttr)
			if err != nil {
				return err
			}
		}
		fmt.Printf("Updated attributes from %v project and domain %v\n", project, domain)
	}
	return nil
}
