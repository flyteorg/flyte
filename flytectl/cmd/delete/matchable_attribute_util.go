package delete

import (
	"context"

	"github.com/flyteorg/flytectl/pkg/ext"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flytestdlib/logger"
)

func deleteMatchableAttr(ctx context.Context, project, domain, workflowName string,
	deleter ext.AdminDeleterExtInterface, rsType admin.MatchableResource, dryRun bool) error {
	if len(workflowName) > 0 {
		// Delete the workflow attribute from the admin. If the attribute doesn't exist , admin deesn't return an error and same behavior is followed here
		if dryRun {
			logger.Infof(ctx, "skipping DeleteWorkflowAttributes request (dryRun)")
		} else {
			err := deleter.DeleteWorkflowAttributes(ctx, project, domain, workflowName, rsType)
			if err != nil {
				return err
			}
		}
		logger.Debugf(ctx, "Deleted matchable resources from %v project and domain %v and workflow %v", project, domain, workflowName)
	} else {
		// Delete the project domain attribute from the admin. If the attribute doesn't exist , admin deesn't return an error and same behavior is followed here
		if dryRun {
			logger.Infof(ctx, "skipping DeleteProjectDomainAttributes request (dryRun)")
		} else {
			err := deleter.DeleteProjectDomainAttributes(ctx, project, domain, rsType)
			if err != nil {
				return err
			}
		}
		logger.Debugf(ctx, "Deleted matchable resources from %v project and domain %v", project, domain)
	}
	return nil
}
