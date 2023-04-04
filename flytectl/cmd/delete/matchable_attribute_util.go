package delete

import (
	"context"
	"fmt"

	"github.com/flyteorg/flytectl/pkg/ext"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
)

func deleteMatchableAttr(ctx context.Context, project, domain, workflowName string,
	deleter ext.AdminDeleterExtInterface, rsType admin.MatchableResource, dryRun bool) error {
	if len(workflowName) > 0 {
		// Delete the workflow attribute from the admin. If the attribute doesn't exist , admin deesn't return an error and same behavior is followed here
		if dryRun {
			fmt.Print("skipping DeleteWorkflowAttributes request (dryRun)\n")
		} else {
			err := deleter.DeleteWorkflowAttributes(ctx, project, domain, workflowName, rsType)
			if err != nil {
				return err
			}
		}
		fmt.Printf("Deleted matchable resources from %v project and domain %v and workflow %v\n", project, domain, workflowName)
	} else {
		// Delete the project domain attribute from the admin. If the attribute doesn't exist , admin deesn't return an error and same behavior is followed here
		if dryRun {
			fmt.Print("skipping DeleteProjectDomainAttributes request (dryRun)\n")
		} else {
			if len(domain) == 0 {
				err := deleter.DeleteProjectAttributes(ctx, project, rsType)
				if err != nil {
					return err
				}
				fmt.Printf("Deleted matchable resources from %v project \n", project)
			} else {
				err := deleter.DeleteProjectDomainAttributes(ctx, project, domain, rsType)
				if err != nil {
					return err
				}
				fmt.Printf("Deleted matchable resources from %v project and domain %v\n", project, domain)
			}
		}

	}
	return nil
}
