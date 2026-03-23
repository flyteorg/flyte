package runs

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/flyteorg/flyte/v2/app"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/actions/actionsconnect"
	flyteappconnect "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/app/appconnect"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/auth/authconnect"
	projectpb "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/project"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/project/projectconnect"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/task/taskconnect"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/trigger/triggerconnect"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow/workflowconnect"
	"github.com/flyteorg/flyte/v2/runs/config"
	"github.com/flyteorg/flyte/v2/runs/migrations"
	"github.com/flyteorg/flyte/v2/runs/repository"
	"github.com/flyteorg/flyte/v2/runs/repository/impl"
	"github.com/flyteorg/flyte/v2/runs/repository/interfaces"
	"github.com/flyteorg/flyte/v2/runs/repository/models"
	"github.com/flyteorg/flyte/v2/runs/service"

	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
)

// Setup registers Run and Task service handlers on the SetupContext mux.
// Requires sc.DB and sc.DataStore to be set. When sc.K8sConfig is provided,
// RunLogsService is also mounted to enable pod log streaming.
func Setup(ctx context.Context, sc *app.SetupContext) error {
	cfg := config.GetConfig()

	if err := migrations.RunMigrations(sc.DB); err != nil {
		return fmt.Errorf("runs: failed to run migrations: %w", err)
	}

	repo := repository.NewRepository(sc.DB, cfg.Database)

	// In unified mode, intra-service calls go through the same mux.
	actionsURL := cfg.ActionsServiceURL
	if sc.BaseURL != "" {
		actionsURL = sc.BaseURL
	}
	actionsClient := actionsconnect.NewActionsServiceClient(
		http.DefaultClient,
		actionsURL,
	)

	runsSvc := service.NewRunService(repo, actionsClient, cfg.StoragePrefix, sc.DataStore)
	taskSvc := service.NewTaskService(repo)

	abortReconciler := service.NewAbortReconciler(repo, actionsClient, service.AbortReconcilerConfig{
		Workers:      5,
		MaxAttempts:  10,
		QueueSize:    1000,
		InitialDelay: time.Second,
		MaxDelay:     5 * time.Minute,
	})
	sc.AddWorker("abort-reconciler", func(ctx context.Context) error {
		return abortReconciler.Run(ctx)
	})

	runsPath, runsHandler := workflowconnect.NewRunServiceHandler(runsSvc)
	sc.Mux.Handle(runsPath, runsHandler)
	logger.Infof(ctx, "Mounted RunService at %s", runsPath)

	internalRunsPath, internalRunsHandler := workflowconnect.NewInternalRunServiceHandler(runsSvc)
	sc.Mux.Handle(internalRunsPath, internalRunsHandler)
	logger.Infof(ctx, "Mounted InternalRunService at %s", internalRunsPath)

	taskPath, taskHandler := taskconnect.NewTaskServiceHandler(taskSvc)
	sc.Mux.Handle(taskPath, taskHandler)
	logger.Infof(ctx, "Mounted TaskService at %s", taskPath)

	translatorSvc := service.NewTranslatorService()
	translatorPath, translatorHandler := workflowconnect.NewTranslatorServiceHandler(translatorSvc)
	sc.Mux.Handle(translatorPath, translatorHandler)
	logger.Infof(ctx, "Mounted TranslatorService at %s", translatorPath)

	identitySvc := service.NewIdentityService()
	identityPath, identityHandler := authconnect.NewIdentityServiceHandler(identitySvc)
	sc.Mux.Handle(identityPath, identityHandler)
	logger.Infof(ctx, "Mounted IdentityService at %s", identityPath)

	appSvc := service.NewAppService()
	appPath, appHandler := flyteappconnect.NewAppServiceHandler(appSvc)
	sc.Mux.Handle(appPath, appHandler)
	logger.Infof(ctx, "Mounted AppService at %s", appPath)

	triggerSvc := service.NewTriggerService()
	triggerPath, triggerHandler := triggerconnect.NewTriggerServiceHandler(triggerSvc)
	sc.Mux.Handle(triggerPath, triggerHandler)
	logger.Infof(ctx, "Mounted TriggerService at %s", triggerPath)

	domains := make([]*projectpb.Domain, 0, len(cfg.Domains))
	for _, d := range cfg.Domains {
		domains = append(domains, &projectpb.Domain{
			Id:   d.ID,
			Name: d.Name,
		})
	}
	projectSvc := service.NewProjectService(impl.NewProjectRepo(sc.DB), domains)
	projectPath, projectHandler := projectconnect.NewProjectServiceHandler(projectSvc)
	sc.Mux.Handle(projectPath, projectHandler)
	logger.Infof(ctx, "Mounted ProjectService at %s", projectPath)

	if sc.K8sConfig != nil {
		logStreamer, err := service.NewK8sLogStreamer(sc.K8sConfig)
		if err != nil {
			return fmt.Errorf("runs: failed to create k8s log streamer: %w", err)
		}
		runLogsSvc := service.NewRunLogsService(repo, logStreamer)
		runLogsPath, runLogsHandler := workflowconnect.NewRunLogsServiceHandler(runLogsSvc)
		sc.Mux.Handle(runLogsPath, runLogsHandler)
		logger.Infof(ctx, "Mounted RunLogsService at %s", runLogsPath)
	}

	if err := seedProjects(ctx, impl.NewProjectRepo(sc.DB), cfg.SeedProjects); err != nil {
		return fmt.Errorf("runs: failed to seed projects: %w", err)
	}

	sc.AddReadyCheck(func(r *http.Request) error {
		sqlDB, err := sc.DB.DB()
		if err != nil {
			return fmt.Errorf("database connection error: %w", err)
		}
		if err := sqlDB.Ping(); err != nil {
			return fmt.Errorf("database ping failed: %w", err)
		}
		return nil
	})

	return nil
}

func seedProjects(ctx context.Context, projectRepo interfaces.ProjectRepo, projects []string) error {
	for _, projectID := range projects {
		if projectID == "" {
			continue
		}

		state := int32(projectpb.ProjectState_PROJECT_STATE_ACTIVE)
		projectModel := &models.Project{
			Identifier: projectID,
			Name:       projectID,
			State:      &state,
		}

		if err := projectRepo.CreateProject(ctx, projectModel); err != nil {
			if errors.Is(err, interfaces.ErrProjectAlreadyExists) {
				continue
			}
			return err
		}

		logger.Infof(ctx, "Seeded project %s", projectID)
	}

	return nil
}
