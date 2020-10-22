package get

import (
	"context"
	"github.com/golang/protobuf/proto"
	"github.com/lyft/flytectl/cmd/config"
	cmdCore "github.com/lyft/flytectl/cmd/core"
	"github.com/lyft/flytectl/pkg/adminutils"
	"github.com/lyft/flytectl/pkg/printer"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/lyft/flytestdlib/logger"
)

var launchplanColumns = []printer.Column{
	{"Version", "$.id.version"},
	{"Name", "$.id.name"},
	{"Type", "$.closure.compiledTask.template.type"},
	{"State", "$.spec.state"},
	{"Schedule", "$.spec.entityMetadata.schedule"},
}

func LaunchplanToProtoMessages(l []*admin.LaunchPlan) []proto.Message {
	messages := make([]proto.Message, 0, len(l))
	for _, m := range l {
		messages = append(messages, m)
	}
	return messages
}

func getLaunchPlanFunc(ctx context.Context, args []string, cmdCtx cmdCore.CommandContext) error {
	launchPlanPrinter := printer.Printer{}

	if len(args) == 1 {
		name := args[0]
		launchPlan, err := cmdCtx.AdminClient().ListLaunchPlans(ctx, &admin.ResourceListRequest{
			Limit: 10,
			Id: &admin.NamedEntityIdentifier{
				Project: config.GetConfig().Project,
				Domain:  config.GetConfig().Domain,
				Name:    name,
			},
		})
		if err != nil {
			return err
		}
		logger.Debugf(ctx, "Retrieved %v excutions", len(launchPlan.LaunchPlans))
		err = launchPlanPrinter.Print(config.GetConfig().MustOutputFormat(), launchplanColumns, LaunchplanToProtoMessages(launchPlan.LaunchPlans)...)
		if err != nil {
			return err
		}
		return nil
	}

	launchPlans, err := adminutils.GetAllNamedEntities(ctx, cmdCtx.AdminClient().ListLaunchPlanIds, adminutils.ListRequest{Project: config.GetConfig().Project, Domain: config.GetConfig().Domain})
	if err != nil {
		return err
	}
	logger.Debugf(ctx, "Retrieved %v launch plans", len(launchPlans))
	return launchPlanPrinter.Print(config.GetConfig().MustOutputFormat(), entityColumns, adminutils.NamedEntityToProtoMessage(launchPlans)...)
	return nil
}
