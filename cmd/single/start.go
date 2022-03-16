package single

import (
	"context"
	adminEntrypoint "github.com/flyteorg/flyteadmin/cmd/entrypoints"
	propellerEntrypoint "github.com/flyteorg/flytepropeller/pkg/controller"
	propellerConfig "github.com/flyteorg/flytepropeller/pkg/controller/config"

	_ "github.com/golang/glog"
	"github.com/spf13/cobra"
	_ "gorm.io/driver/postgres" // Required to import database driver.

	_ "github.com/flyteorg/flyteplugins/go/tasks/plugins/array/k8s"
	_ "github.com/flyteorg/flyteplugins/go/tasks/plugins/k8s/pod"
)

func startAdmin(ctx context.Context) error {
	return adminEntrypoint.Serve(ctx)
}

func startPropeller(ctx context.Context) error {
	return propellerEntrypoint.StartController(ctx, propellerConfig.GetConfig(), "all")
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "This command will start the flyte native scheduler and periodically get new schedules from the db for scheduling",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		childCtx, cancel := context.WithCancel(ctx)
		defer cancel()
		go func(ctx context.Context) {
			err := startAdmin(ctx)
			if err != nil {
				return
			}
		}(childCtx)

		go func(ctx context.Context) {
			err := startPropeller(ctx)
			if err != nil {
				return
			}
		}(childCtx)

		<-ctx.Done()
		return nil
	},
}

func init() {
	RootCmd.AddCommand(startCmd)
}
