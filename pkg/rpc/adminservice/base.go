package adminservice

import (
	"context"
	"fmt"
	"runtime/debug"

	executionCluster "github.com/lyft/flyteadmin/pkg/executioncluster/impl"

	"github.com/lyft/flytestdlib/profutils"

	"github.com/lyft/flyteadmin/pkg/async/schedule"
	"github.com/lyft/flyteadmin/pkg/data"

	"github.com/lyft/flytestdlib/promutils"
	"github.com/lyft/flytestdlib/storage"

	"github.com/golang/protobuf/proto"
	"github.com/lyft/flytestdlib/logger"

	"github.com/lyft/flyteadmin/pkg/async/notifications"
	manager "github.com/lyft/flyteadmin/pkg/manager/impl"
	"github.com/lyft/flyteadmin/pkg/manager/interfaces"
	"github.com/lyft/flyteadmin/pkg/repositories"
	repositoryConfig "github.com/lyft/flyteadmin/pkg/repositories/config"
	"github.com/lyft/flyteadmin/pkg/runtime"
	workflowengine "github.com/lyft/flyteadmin/pkg/workflowengine/impl"
)

type AdminService struct {
	TaskManager          interfaces.TaskInterface
	WorkflowManager      interfaces.WorkflowInterface
	LaunchPlanManager    interfaces.LaunchPlanInterface
	ExecutionManager     interfaces.ExecutionInterface
	NodeExecutionManager interfaces.NodeExecutionInterface
	TaskExecutionManager interfaces.TaskExecutionInterface
	ProjectManager       interfaces.ProjectInterface
	Metrics              AdminMetrics
}

// Intercepts all admin requests to handle panics during execution.
func (m *AdminService) interceptPanic(ctx context.Context, request proto.Message) {
	err := recover()
	if err == nil {
		return
	}

	m.Metrics.PanicCounter.Inc()
	logger.Fatalf(ctx, "panic-ed for request: [%+v] with err: %v", request, err)
}

const defaultRetries = 3

func NewAdminServer(kubeConfig, master string) *AdminService {
	configuration := runtime.NewConfigurationProvider()
	applicationConfiguration := configuration.ApplicationConfiguration().GetTopLevelConfig()

	adminScope := promutils.NewScope(applicationConfiguration.MetricsScope).NewSubScope("admin")

	defer func() {
		if err := recover(); err != nil {
			adminScope.MustNewCounter("initialization_panic",
				"panics encountered initializating the admin service").Inc()
			logger.Fatalf(context.Background(), fmt.Sprintf("caught panic: %v [%+v]", err, string(debug.Stack())))
		}
	}()

	dbConfigValues := configuration.ApplicationConfiguration().GetDbConfig()
	dbConfig := repositoryConfig.DbConfig{
		Host:         dbConfigValues.Host,
		Port:         dbConfigValues.Port,
		DbName:       dbConfigValues.DbName,
		User:         dbConfigValues.User,
		Password:     dbConfigValues.Password,
		ExtraOptions: dbConfigValues.ExtraOptions,
	}
	db := repositories.GetRepository(
		repositories.POSTGRES, dbConfig, adminScope.NewSubScope("database"))
	storeConfig := storage.GetConfig()
	executionCluster := executionCluster.GetExecutionCluster(
		adminScope.NewSubScope("executor").NewSubScope("cluster"),
		kubeConfig,
		master,
		configuration)
	workflowExecutor := workflowengine.NewFlytePropeller(
		applicationConfiguration.RoleNameKey,
		executionCluster,
		adminScope.NewSubScope("executor").NewSubScope("flytepropeller"))
	logger.Info(context.Background(), "Successfully created a workflow executor engine")
	dataStorageClient, err := storage.NewDataStore(storeConfig, adminScope.NewSubScope("storage"))
	if err != nil {
		logger.Error(context.Background(), "Failed to initialize storage config")
		panic(err)
	}

	publisher := notifications.NewNotificationsPublisher(*configuration.ApplicationConfiguration().GetNotificationsConfig(), adminScope)
	processor := notifications.NewNotificationsProcessor(*configuration.ApplicationConfiguration().GetNotificationsConfig(), adminScope)
	go func() {
		err = processor.StartProcessing()
		if err != nil {
			logger.Errorf(context.Background(), "error with starting processor err: [%v] ", err)
		} else {
			logger.Info(context.Background(), "Successfully started processing notifications.")
		}
	}()

	// Configure workflow scheduler async processes.
	schedulerConfig := configuration.ApplicationConfiguration().GetSchedulerConfig()
	workflowScheduler := schedule.NewWorkflowScheduler(schedule.WorkflowSchedulerConfig{
		Retries:                defaultRetries,
		EventSchedulerConfig:   schedulerConfig.EventSchedulerConfig,
		WorkflowExecutorConfig: schedulerConfig.WorkflowExecutorConfig,
		Scope:                  adminScope,
	})

	eventScheduler := workflowScheduler.GetEventScheduler()
	launchPlanManager := manager.NewLaunchPlanManager(
		db, configuration, eventScheduler, adminScope.NewSubScope("launch_plan_manager"))

	// Configure admin-specific remote data handler (separate from storage)
	remoteDataConfig := configuration.ApplicationConfiguration().GetRemoteDataConfig()
	urlData := data.GetRemoteDataHandler(data.RemoteDataHandlerConfig{
		CloudProvider:            remoteDataConfig.Scheme,
		SignedURLDurationMinutes: remoteDataConfig.SignedURL.DurationMinutes,
		Region:                   remoteDataConfig.Region,
		Retries:                  defaultRetries,
		RemoteDataStoreClient:    dataStorageClient,
	}).GetRemoteURLInterface()

	executionManager := manager.NewExecutionManager(
		db, configuration, dataStorageClient, workflowExecutor, adminScope.NewSubScope("execution_manager"),
		adminScope.NewSubScope("user_execution_metrics"), publisher, urlData)

	scheduledWorkflowExecutor := workflowScheduler.GetWorkflowExecutor(executionManager, launchPlanManager)
	logger.Info(context.Background(), "Successfully initialized a new scheduled workflow executor")
	go func() {
		scheduledWorkflowExecutor.Run()
		logger.Info(context.Background(), "Successfully started running the scheduled workflow executor")
	}()

	// Serve profiling endpoints.
	go func() {
		err := profutils.StartProfilingServerWithDefaultHandlers(
			context.Background(), applicationConfiguration.ProfilerPort, nil)
		if err != nil {
			logger.Panicf(context.Background(), "Failed to Start profiling and Metrics server. Error, %v", err)
		}
	}()

	logger.Info(context.Background(), "Initializing a new AdminService")
	return &AdminService{
		TaskManager: manager.NewTaskManager(db, configuration, workflowengine.NewCompiler(),
			adminScope.NewSubScope("task_manager")),
		WorkflowManager: manager.NewWorkflowManager(
			db, configuration, workflowengine.NewCompiler(), dataStorageClient, applicationConfiguration.MetadataStoragePrefix,
			adminScope.NewSubScope("workflow_manager")),
		LaunchPlanManager: launchPlanManager,
		ExecutionManager:  executionManager,
		NodeExecutionManager: manager.NewNodeExecutionManager(
			db, adminScope.NewSubScope("node_execution_manager"), urlData),
		TaskExecutionManager: manager.NewTaskExecutionManager(
			db, adminScope.NewSubScope("task_execution_manager"), urlData),
		ProjectManager: manager.NewProjectManager(db, configuration),
		Metrics:        InitMetrics(adminScope),
	}
}
