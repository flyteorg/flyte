package runs

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"connectrpc.com/connect"
	"connectrpc.com/otelconnect"
	"github.com/flyteorg/flyte/v2/flytestdlib/app"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/actions/actionsconnect"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/auth/authconnect"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/dataproxy/dataproxyconnect"
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
	"github.com/flyteorg/flyte/v2/runs/scheduler"
	"github.com/flyteorg/flyte/v2/runs/service"

	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	"github.com/flyteorg/flyte/v2/flytestdlib/otelutils"
)

const otelServiceName = "runs-service"

// Setup registers Run and Task service handlers on the SetupContext mux.
// Requires sc.DB and sc.DataStore to be set. When sc.K8sConfig is provided,
// RunLogsService is also mounted to enable pod log streaming.
func Setup(ctx context.Context, sc *app.SetupContext) error {
	cfg := config.GetConfig()
	if err := migrations.RunMigrations(ctx, sc.DB); err != nil {
		return fmt.Errorf("runs: failed to run migrations: %w", err)
	}

	otelCfg := otelutils.GetConfig()
	if err := otelutils.RegisterProvidersWithContext(ctx, otelServiceName, otelCfg); err != nil {
		return fmt.Errorf("registering otel providers: %w", err)
	}
	otelInterceptor, err := otelconnect.NewInterceptor(
		otelconnect.WithTracerProvider(otelutils.GetTracerProvider(otelServiceName)),
		otelconnect.WithMeterProvider(otelutils.GetMeterProvider(otelServiceName)),
	)
	if err != nil {
		return fmt.Errorf("creating otel interceptor: %w", err)
	}

	repo, err := repository.NewRepository(sc.DB, cfg.Database)
	if err != nil {
		return fmt.Errorf("runs: failed to create repository: %w", err)
	}

	// In unified mode, intra-service calls go through the same mux.
	actionsURL := cfg.ActionsServiceURL
	if sc.BaseURL != "" {
		actionsURL = sc.BaseURL
	}
	actionsClient := actionsconnect.NewActionsServiceClient(
		http.DefaultClient,
		actionsURL,
		connect.WithInterceptors(otelInterceptor),
	)

	projectsURL := sc.BaseURL
	if projectsURL == "" {
		projectsURL = cfg.ActionsServiceURL
	}
	projectClient := projectconnect.NewProjectServiceClient(
		http.DefaultClient,
		projectsURL,
		connect.WithInterceptors(otelInterceptor),
	)
	dataProxyClient := dataproxyconnect.NewDataProxyServiceClient(http.DefaultClient, projectsURL, connect.WithInterceptors(otelInterceptor))

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

	runsSvc := service.NewRunService(repo, actionsClient, dataProxyClient, projectClient, cfg.StoragePrefix, sc.DataStore, abortReconciler)
	taskSvc := service.NewTaskService(repo, projectClient)

	runsPath, runsHandler := workflowconnect.NewRunServiceHandler(runsSvc, connect.WithInterceptors(otelInterceptor))
	sc.Mux.Handle(runsPath, runsHandler)
	logger.Infof(ctx, "Mounted RunService at %s", runsPath)

	internalRunsPath, internalRunsHandler := workflowconnect.NewInternalRunServiceHandler(runsSvc, connect.WithInterceptors(otelInterceptor))
	sc.Mux.Handle(internalRunsPath, internalRunsHandler)
	logger.Infof(ctx, "Mounted InternalRunService at %s", internalRunsPath)

	taskPath, taskHandler := taskconnect.NewTaskServiceHandler(taskSvc, connect.WithInterceptors(otelInterceptor))
	sc.Mux.Handle(taskPath, taskHandler)
	logger.Infof(ctx, "Mounted TaskService at %s", taskPath)

	identitySvc := service.NewIdentityService()
	identityPath, identityHandler := authconnect.NewIdentityServiceHandler(identitySvc, connect.WithInterceptors(otelInterceptor))
	sc.Mux.Handle(identityPath, identityHandler)
	logger.Infof(ctx, "Mounted IdentityService at %s", identityPath)

	authMetadataSvc := service.NewAuthMetadataService(sc.BaseURL, service.ExternalAuthServerConfig{
		BaseURL:       cfg.AuthMetadata.ExternalAuthServerBaseURL,
		MetadataURL:   cfg.AuthMetadata.ExternalMetadataURL,
		RetryAttempts: cfg.AuthMetadata.RetryAttempts,
		RetryDelay:    cfg.AuthMetadata.RetryDelay,
	}, service.PublicClientConfig{
		ClientID:                 cfg.AuthMetadata.FlyteClient.ClientID,
		RedirectURI:              cfg.AuthMetadata.FlyteClient.RedirectURI,
		Scopes:                   cfg.AuthMetadata.FlyteClient.Scopes,
		Audience:                 cfg.AuthMetadata.FlyteClient.Audience,
		AuthorizationMetadataKey: cfg.AuthMetadata.AuthorizationMetadataKey,
	})
	authMetadataPath, authMetadataHandler := authconnect.NewAuthMetadataServiceHandler(authMetadataSvc, connect.WithInterceptors(otelInterceptor))
	sc.Mux.Handle(authMetadataPath, authMetadataHandler)
	logger.Infof(ctx, "Mounted AuthMetadataService at %s", authMetadataPath)

	// Serve OAuth2 authorization-server metadata at the RFC 8414 well-known path
	// so OAuth2/OIDC discovery clients (flytectl, pyflyte) can find it. When
	// runs.authMetadata.externalAuthServerBaseUrl is set, this proxies the
	// external IdP's (e.g. Okta) metadata document.
	sc.Mux.Handle("/.well-known/oauth-authorization-server", service.OAuth2MetadataHTTPHandler(authMetadataSvc))
	logger.Infof(ctx, "Mounted OAuth2 metadata at /.well-known/oauth-authorization-server")

	triggerSvc := service.NewTriggerService(repo)
	triggerPath, triggerHandler := triggerconnect.NewTriggerServiceHandler(triggerSvc, connect.WithInterceptors(otelInterceptor))
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
	projectPath, projectHandler := projectconnect.NewProjectServiceHandler(projectSvc, connect.WithInterceptors(otelInterceptor))
	sc.Mux.Handle(projectPath, projectHandler)
	logger.Infof(ctx, "Mounted ProjectService at %s", projectPath)

	if sc.K8sConfig != nil {
		logStreamer, err := service.NewK8sLogStreamer(sc.K8sConfig)
		if err != nil {
			return fmt.Errorf("runs: failed to create k8s log streamer: %w", err)
		}
		runLogsSvc := service.NewRunLogsService(repo, logStreamer)
		runLogsPath, runLogsHandler := workflowconnect.NewRunLogsServiceHandler(runLogsSvc, connect.WithInterceptors(otelInterceptor))
		sc.Mux.Handle(runLogsPath, runLogsHandler)
		logger.Infof(ctx, "Mounted RunLogsService at %s", runLogsPath)
	}

	if err := seedProjects(ctx, impl.NewProjectRepo(sc.DB), cfg.SeedProjects); err != nil {
		return fmt.Errorf("runs: failed to seed projects: %w", err)
	}

	if cfg.TriggerScheduler.Enabled {
		runsURL := cfg.ActionsServiceURL
		if sc.BaseURL != "" {
			runsURL = sc.BaseURL
		}
		worker := scheduler.Start(ctx, repo.TriggerRepo(), cfg.TriggerScheduler, runsURL, connect.WithInterceptors(otelInterceptor))
		sc.AddWorker("trigger-scheduler", worker)
		logger.Infof(ctx, "Registered trigger-scheduler worker")
	}

	sc.AddReadyCheck(func(r *http.Request) error {
		if err := sc.DB.PingContext(r.Context()); err != nil {
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
