package get

import (
	"context"

	sconfig "github.com/flyteorg/flytectl/cmd/config/subcommand"
	"github.com/flyteorg/flytectl/pkg/ext"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
)

func FetchAndUnDecorateMatchableAttr(ctx context.Context, project, domain, workflowName string,
	fetcher ext.AdminFetcherExtInterface, unDecorator sconfig.MatchableAttributeUnDecorator, rsType admin.MatchableResource) error {
	if len(workflowName) > 0 {
		// Fetch the workflow attribute from the admin
		workflowAttr, err := fetcher.FetchWorkflowAttributes(ctx,
			project, domain, workflowName, rsType)
		if err != nil {
			return err
		}
		// Update the shadow config with the fetched taskResourceAttribute which can then be written to a file which can then be called for an update.
		unDecorator.UnDecorate(workflowAttr.GetAttributes().GetMatchingAttributes())
	} else {
		// Fetch the project domain attribute from the admin
		projectDomainAttr, err := fetcher.FetchProjectDomainAttributes(ctx, project, domain, rsType)
		if err != nil {
			return err
		}
		// Update the shadow config with the fetched taskResourceAttribute which can then be written to a file which can then be called for an update.
		unDecorator.UnDecorate(projectDomainAttr.GetAttributes().GetMatchingAttributes())
	}
	return nil
}
