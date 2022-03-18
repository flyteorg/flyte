package single

import (
	"context"
	"fmt"
	"github.com/flyteorg/flyteadmin/pkg/clusterresource"
	"github.com/flyteorg/flyteadmin/pkg/runtime"
	adminServer "github.com/flyteorg/flyteadmin/pkg/server"
	adminScheduler "github.com/flyteorg/flyteadmin/scheduler"
	propellerEntrypoint "github.com/flyteorg/flytepropeller/pkg/controller"
	propellerConfig "github.com/flyteorg/flytepropeller/pkg/controller/config"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/promutils"
	"net/http"

	_ "github.com/golang/glog"
	"github.com/spf13/cobra"
	_ "gorm.io/driver/postgres" // Required to import database driver.

	_ "github.com/flyteorg/flyteplugins/go/tasks/plugins/array/k8s"
	_ "github.com/flyteorg/flyteplugins/go/tasks/plugins/k8s/pod"
)

func GetConsoleHandlers() map[string]func(http.ResponseWriter, *http.Request) {
	handlers := make(map[string]func(http.ResponseWriter, *http.Request))
	// Serves console
	fs := http.StripPrefix("/console/", http.FileServer(http.Dir("./dist")))
	handlers["/console/"] = func(writer http.ResponseWriter, request *http.Request) {
		fmt.Printf("all files under console returning, %s\n", request.URL.Path)
		fs.ServeHTTP(writer, request)
	}
	handlers["/console"] = func(writer http.ResponseWriter, request *http.Request) {
		fmt.Println("returning index.html")
		http.ServeFile(writer, request, "./dist/index.html")
	}

	return handlers
}

func startClusterResourceController(ctx context.Context) error {
	configuration := runtime.NewConfigurationProvider()
	scope := promutils.NewScope(configuration.ApplicationConfiguration().GetTopLevelConfig().MetricsScope).NewSubScope("clusterresource")
	clusterResourceController, err := clusterresource.NewClusterResourceControllerFromConfig(ctx, scope, configuration)
	if err != nil {
		return err
	}
	clusterResourceController.Run()
	logger.Infof(ctx, "ClusterResourceController started running successfully")
	return nil
}

func startAdmin(ctx context.Context) error {
	if err := adminServer.Migrate(ctx); err != nil {
		return err
	}
	if err := adminServer.SeedProjects(ctx, []string{"flytesnacks"}); err != nil {
		return err
	}
	if err := adminScheduler.StartScheduler(ctx); err != nil {
		return err
	}
	if err := startClusterResourceController(ctx); err != nil {
		return err
	}
	return adminServer.Serve(ctx, GetConsoleHandlers())
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
