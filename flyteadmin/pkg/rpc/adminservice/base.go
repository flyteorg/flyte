package adminservice

import (
	"context"
	"fmt"
	"runtime/debug"

	"github.com/flyteorg/flyte/flyteadmin/pkg/async/cloudevent"
	eventWriter "github.com/flyteorg/flyte/flyteadmin/pkg/async/events/implementations"
	"github.com/flyteorg/flyte/flyteadmin/pkg/async/notifications"
	"github.com/flyteorg/flyte/flyteadmin/pkg/async/schedule"
	"github.com/flyteorg/flyte/flyteadmin/pkg/data"
	executionCluster "github.com/flyteorg/flyte/flyteadmin/pkg/executioncluster/impl"
	manager "github.com/flyteorg/flyte/flyteadmin/pkg/manager/impl"
	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/impl/resources"
	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/interfaces"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/errors"
	runtimeIfaces "github.com/flyteorg/flyte/flyteadmin/pkg/runtime/interfaces"
	workflowengineImpl "github.com/flyteorg/flyte/flyteadmin/pkg/workflowengine/impl"
	"github.com/flyteorg/flyte/flyteadmin/plugins"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/service"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
	"github.com/flyteorg/flyte/flytestdlib/storage"
)

type AdminService struct {
	service.UnimplementedAdminServiceServer
	TaskManager              interfaces.TaskInterface
	WorkflowManager          interfaces.WorkflowInterface
	LaunchPlanManager        interfaces.LaunchPlanInterface
	ExecutionManager         interfaces.ExecutionInterface
	NodeExecutionManager     interfaces.NodeExecutionInterface
	TaskExecutionManager     interfaces.TaskExecutionInterface
	ProjectManager           interfaces.ProjectInterface
	ResourceManager          interfaces.ResourceInterface
	NamedEntityManager       interfaces.NamedEntityInterface
	VersionManager           interfaces.VersionInterface
	DescriptionEntityManager interfaces.DescriptionEntityInterface
	MetricsManager           interfaces.MetricsInterface
	Metrics                  AdminMetrics
}

const defaultRetries = 3

func NewAdminServer(ctx context.Context, pluginRegistry *plugins.Registry, configuration runtimeIfaces.Configuration,
	kubeConfig, master string, dataStorageClient *storage.DataStore, adminScope promutils.Scope) *AdminService {
	applicationConfiguration := configuration.ApplicationConfiguration().GetTopLevelConfig()

	panicCounter := adminScope.MustNewCounter("initialization_panic",
		"panics encountered initializing the admin service")

	defer func() {
		if err := recover(); err != nil {
			panicCounter.Inc()
			logger.Fatalf(ctx, fmt.Sprintf("caught panic: %v [%+v]", err, string(debug.Stack())))
		}
	}()

	databaseConfig := configuration.ApplicationConfiguration().GetDbConfig()
	logConfig := logger.GetConfig()

	db, err := repositories.GetDB(ctx, databaseConfig, logConfig)
	if err != nil {
		logger.Fatal(ctx, err)
	}
	dbScope := adminScope.NewSubScope("database")
	repo := repositories.NewGormRepo(
		db, errors.NewPostgresErrorTransformer(adminScope.NewSubScope("errors")), dbScope)
	execCluster := executionCluster.GetExecutionCluster(
		adminScope.NewSubScope("executor").NewSubScope("cluster"),
		kubeConfig,
		master,
		configuration,
		repo)
	workflowBuilder := workflowengineImpl.NewFlyteWorkflowBuilder(
		adminScope.NewSubScope("builder").NewSubScope("flytepropeller"))
	workflowExecutor := workflowengineImpl.NewK8sWorkflowExecutor(configuration, execCluster, workflowBuilder, dataStorageClient)
	logger.Info(ctx, "Successfully created a workflow executor engine")
	pluginRegistry.RegisterDefault(plugins.PluginIDWorkflowExecutor, workflowExecutor)

	publisher := notifications.NewNotificationsPublisher(*configuration.ApplicationConfiguration().GetNotificationsConfig(), adminScope)
	processor := notifications.NewNotificationsProcessor(*configuration.ApplicationConfiguration().GetNotificationsConfig(), adminScope)
	eventPublisher := notifications.NewEventsPublisher(*configuration.ApplicationConfiguration().GetExternalEventsConfig(), adminScope)
	go func() {
		logger.Info(ctx, "Started processing notifications.")
		processor.StartProcessing()
	}()

	// Configure workflow scheduler async processes.
	schedulerConfig := configuration.ApplicationConfiguration().GetSchedulerConfig()
	workflowScheduler := schedule.NewWorkflowScheduler(repo, schedule.WorkflowSchedulerConfig{
		Retries:         defaultRetries,
		SchedulerConfig: *schedulerConfig,
		Scope:           adminScope,
	})

	eventScheduler := workflowScheduler.GetEventScheduler()

	launchPlanManager := manager.NewLaunchPlanManager(
		repo, configuration, eventScheduler, adminScope.NewSubScope("launch_plan_manager"))

	// Configure admin-specific remote data handler (separate from storage)
	remoteDataConfig := configuration.ApplicationConfiguration().GetRemoteDataConfig()
	urlData := data.GetRemoteDataHandler(data.RemoteDataHandlerConfig{
		CloudProvider:            remoteDataConfig.Scheme,
		SignedURLDurationMinutes: remoteDataConfig.SignedURL.DurationMinutes,
		SigningPrincipal:         remoteDataConfig.SignedURL.SigningPrincipal,
		Region:                   remoteDataConfig.Region,
		Retries:                  defaultRetries,
		RemoteDataStoreClient:    dataStorageClient,
	}).GetRemoteURLInterface()

	cloudEventPublisher := cloudevent.NewCloudEventsPublisher(ctx, repo, dataStorageClient, urlData, *configuration.ApplicationConfiguration().GetCloudEventsConfig(), *configuration.ApplicationConfiguration().GetRemoteDataConfig(), adminScope)

	workflowManager := manager.NewWorkflowManager(
		repo, configuration, workflowengineImpl.NewCompiler(), dataStorageClient, applicationConfiguration.GetMetadataStoragePrefix(),
		adminScope.NewSubScope("workflow_manager"))
	namedEntityManager := manager.NewNamedEntityManager(repo, configuration, adminScope.NewSubScope("named_entity_manager"))
	descriptionEntityManager := manager.NewDescriptionEntityManager(repo, configuration, adminScope.NewSubScope("description_entity_manager"))

	executionEventWriter := eventWriter.NewWorkflowExecutionEventWriter(repo, applicationConfiguration.GetAsyncEventsBufferSize())
	go func() {
		executionEventWriter.Run()
	}()

	executionManager := manager.NewExecutionManager(repo, pluginRegistry, configuration, dataStorageClient,
		adminScope.NewSubScope("execution_manager"), adminScope.NewSubScope("user_execution_metrics"),
		publisher, urlData, workflowManager, namedEntityManager, eventPublisher, cloudEventPublisher, executionEventWriter)
	versionManager := manager.NewVersionManager()

	scheduledWorkflowExecutor := workflowScheduler.GetWorkflowExecutor(executionManager, launchPlanManager)
	logger.Info(ctx, "Successfully initialized a new scheduled workflow executor")
	go func() {
		logger.Info(ctx, "Starting the scheduled workflow executor")
		scheduledWorkflowExecutor.Run()
	}()

	nodeExecutionEventWriter := eventWriter.NewNodeExecutionEventWriter(repo, applicationConfiguration.GetAsyncEventsBufferSize())
	go func() {
		nodeExecutionEventWriter.Run()
	}()

	nodeExecutionManager := manager.NewNodeExecutionManager(repo, configuration, applicationConfiguration.GetMetadataStoragePrefix(), dataStorageClient,
		adminScope.NewSubScope("node_execution_manager"), urlData, eventPublisher, cloudEventPublisher, nodeExecutionEventWriter)
	taskExecutionManager := manager.NewTaskExecutionManager(repo, configuration, dataStorageClient,
		adminScope.NewSubScope("task_execution_manager"), urlData, eventPublisher, cloudEventPublisher)

	logger.Info(ctx, "Initializing a new AdminService")
	return &AdminService{
		TaskManager: manager.NewTaskManager(repo, configuration, workflowengineImpl.NewCompiler(),
			adminScope.NewSubScope("task_manager")),
		WorkflowManager:          workflowManager,
		LaunchPlanManager:        launchPlanManager,
		ExecutionManager:         executionManager,
		NamedEntityManager:       namedEntityManager,
		DescriptionEntityManager: descriptionEntityManager,
		VersionManager:           versionManager,
		NodeExecutionManager:     nodeExecutionManager,
		TaskExecutionManager:     taskExecutionManager,
		ProjectManager:           manager.NewProjectManager(repo, configuration),
		ResourceManager:          resources.NewResourceManager(repo, configuration.ApplicationConfiguration()),
		MetricsManager: manager.NewMetricsManager(workflowManager, executionManager, nodeExecutionManager,
			taskExecutionManager, adminScope.NewSubScope("metrics_manager")),
		Metrics: InitMetrics(adminScope),
	}
}
