package runs

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"connectrpc.com/connect/v2"
	"connectrpc.com/connect/v2/connecthttp"
	"connectrpc.com/otelconnect"

	"github.com/flyteorg/flyte/v2/flytestdlib/app"
	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	"github.com/flyteorg/flyte/v2/flytestdlib/otelutils"
	"github.com/flyteorg/flyte/v2/flytestdlib/sentryutils"
	"github.com/flyteorg/flyte/v2/flytestdlib/serviceclient"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/actions/actionsconnect"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/auth/authconnect"
	projectpb "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/project"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/project/projectconnect"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/settings/settingsconnect"
	taskpb "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/task"
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
)

// ServiceName identifies telemetry emitted by this service and its clients.
const ServiceName = "runs-service"

var sentryOperations = map[string]string{
	workflowconnect.RunServiceCreateRunProcedure:        "create_run",
	taskconnect.TaskServiceDeployTaskProcedure:          "deploy_task",
	triggerconnect.TriggerServiceDeployTriggerProcedure: "deploy_trigger",
}

// triggerCountingStream observes the decoded request without consuming it before
// the generated handler receives it.
type triggerCountingStream struct {
	connect.ServerStream
	triggers int
}

func (s *triggerCountingStream) Receive(msg any) error {
	if err := s.ServerStream.Receive(msg); err != nil {
		return err
	}
	if req, ok := msg.(*taskpb.DeployTaskRequest); ok {
		s.triggers = len(req.Triggers)
	}
	return nil
}

// deployTriggerInterceptor counts the triggers that ride along inside a
// DeployTask request. TriggerService.DeployTrigger only ever sees standalone
// trigger deploys, so without this the deploy_trigger count misses every
// trigger declared on a task.
func deployTriggerInterceptor() connect.ServerInterceptor {
	return func(next connect.ServerFunc) connect.ServerFunc {
		return func(ctx context.Context, spec connect.Spec, stream connect.ServerStream) error {
			if spec.Procedure != taskconnect.TaskServiceDeployTaskProcedure {
				return next(ctx, spec, stream)
			}
			observed := &triggerCountingStream{ServerStream: stream}
			err := next(ctx, spec, observed)
			if err == nil && observed.triggers > 0 {
				sentryutils.Count(ctx, "deploy_trigger", int64(observed.triggers))
			}
			return err
		}
	}
}

// Setup registers Run and Task service handlers on the SetupContext mux.
// Requires sc.DB and sc.DataStore to be set.
func Setup(ctx context.Context, sc *app.SetupContext) error {
	cfg := config.GetConfig()
	sc.AddProfilingServer("runs", cfg.PprofPort)
	if err := migrations.RunMigrations(ctx, sc.DB); err != nil {
		return fmt.Errorf("runs: failed to run migrations: %w", err)
	}

	otelCfg := otelutils.GetConfig()
	if err := otelutils.RegisterProvidersWithContext(ctx, ServiceName, otelCfg); err != nil {
		return fmt.Errorf("registering otel providers: %w", err)
	}
	meterProvider := otelutils.GetMeterProvider(ServiceName)
	otelInterceptor, err := otelconnect.NewServerInterceptor(
		otelconnect.WithTracerProvider(otelutils.GetTracerProvider(ServiceName)),
		otelconnect.WithMeterProvider(meterProvider),
	)
	if err != nil {
		return fmt.Errorf("creating otel interceptor: %w", err)
	}
	otelClientInterceptor, err := otelconnect.NewClientInterceptor(
		otelconnect.WithTracerProvider(otelutils.GetTracerProvider(ServiceName)),
		otelconnect.WithMeterProvider(meterProvider),
	)
	if err != nil {
		return fmt.Errorf("creating otel client interceptor: %w", err)
	}

	interceptors := []connect.ServerInterceptor{otelInterceptor}
	if sentryutils.Init(ctx, ServiceName) {
		interceptors = append(interceptors, sentryutils.Interceptor(sentryOperations), deployTriggerInterceptor())
		sc.AddWorker("sentry-flush", func(ctx context.Context) error {
			<-ctx.Done()
			sentryutils.Flush()
			return nil
		})
	}

	repo, err := repository.NewRepository(
		sc.DB,
		cfg.Database,
		repository.NewNotificationConfig(
			cfg.NotificationBufferLimit,
			cfg.NotifyRetryMinBackoff.Duration,
			cfg.NotifyRetryMaxBackoff.Duration,
			cfg.WatchBufferSize,
			cfg.NotifyBatchSize,
		),
		meterProvider,
	)
	if err != nil {
		return fmt.Errorf("runs: failed to create repository: %w", err)
	}

	settingsRepo := impl.NewSettingsRepo(sc.DB)

	// In unified mode, intra-service calls go through the same mux.
	actionsServiceCfg := cfg.ActionsService
	if sc.BaseURL != "" {
		actionsServiceCfg.URL = sc.BaseURL
	}
	actionsHTTPClient, err := serviceclient.NewHTTPClient(ctx, http.DefaultClient, actionsServiceCfg)
	if err != nil {
		return fmt.Errorf("runs: configure actions service client: %w", err)
	}
	actionsClient := actionsconnect.NewActionsServiceClient(connect.NewClient(connecthttp.NewTransport(
		actionsHTTPClient,
		actionsServiceCfg.URL,
		connecthttp.WithReadMaxBytes(0),
	), otelClientInterceptor,
	))

	runsServiceURL := fmt.Sprintf("http://localhost:%d", cfg.Server.Port)
	projectClient := projectconnect.NewProjectServiceClient(connect.NewClient(connecthttp.NewTransport(
		http.DefaultClient,
		runsServiceURL,
		connecthttp.WithReadMaxBytes(0),
	), otelClientInterceptor,
	))
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

	runsSvc := service.NewRunService(repo, settingsRepo, actionsClient, projectClient, cfg.StoragePrefix, sc.DataStore, abortReconciler, cfg.AuthMetadata.ExternalAuthServerBaseURL, cfg.TrustForwardedIdentityHeaders, cfg.IdentityHeaders)
	taskSvc := service.NewTaskService(repo, projectClient)

	runsHandler := connect.NewServer(interceptors...)
	workflowconnect.RegisterRunServiceHandler(runsHandler, runsSvc)
	connecthttp.Mount(sc.Mux, runsHandler, connecthttp.WithReadMaxBytes(0))
	runsPath := "/" + workflowconnect.RunServiceName + "/"
	logger.Infof(ctx, "Mounted RunService at %s", runsPath)

	internalRunsHandler := connect.NewServer(otelInterceptor)
	workflowconnect.RegisterInternalRunServiceHandler(internalRunsHandler, runsSvc)
	connecthttp.Mount(sc.Mux, internalRunsHandler, connecthttp.WithReadMaxBytes(0))
	internalRunsPath := "/" + workflowconnect.InternalRunServiceName + "/"
	logger.Infof(ctx, "Mounted InternalRunService at %s", internalRunsPath)

	taskHandler := connect.NewServer(interceptors...)
	taskconnect.RegisterTaskServiceHandler(taskHandler, taskSvc)
	connecthttp.Mount(sc.Mux, taskHandler, connecthttp.WithReadMaxBytes(0))
	taskPath := "/" + taskconnect.TaskServiceName + "/"
	logger.Infof(ctx, "Mounted TaskService at %s", taskPath)

	identitySvc := service.NewIdentityService(cfg.AuthMetadata.ExternalAuthServerBaseURL, cfg.TrustForwardedIdentityHeaders, cfg.IdentityHeaders)
	identityHandler := connect.NewServer(otelInterceptor)
	authconnect.RegisterIdentityServiceHandler(identityHandler, identitySvc)
	connecthttp.Mount(sc.Mux, identityHandler, connecthttp.WithReadMaxBytes(0))
	identityPath := "/" + authconnect.IdentityServiceName + "/"
	logger.Infof(ctx, "Mounted IdentityService at %s", identityPath)

	authMetadataSvc := service.NewAuthMetadataService(sc.BaseURL, cfg.AuthMetadata)
	authMetadataHandler := connect.NewServer(otelInterceptor)
	authconnect.RegisterAuthMetadataServiceHandler(authMetadataHandler, authMetadataSvc)
	connecthttp.Mount(sc.Mux, authMetadataHandler, connecthttp.WithReadMaxBytes(0))
	authMetadataPath := "/" + authconnect.AuthMetadataServiceName + "/"
	logger.Infof(ctx, "Mounted AuthMetadataService at %s", authMetadataPath)

	// Serve OAuth2 authorization-server metadata at the RFC 8414 well-known path
	// so OAuth2/OIDC discovery clients (flyte-sdk) can find it. When
	// runs.authMetadata.externalAuthServerBaseUrl is set, this proxies the
	// external IdP's (e.g. Okta) metadata document.
	sc.Mux.Handle("/.well-known/oauth-authorization-server", service.OAuth2MetadataHTTPHandler(authMetadataSvc))
	logger.Infof(ctx, "Mounted OAuth2 metadata at /.well-known/oauth-authorization-server")

	triggerSvc := service.NewTriggerService(repo)
	triggerHandler := connect.NewServer(interceptors...)
	triggerconnect.RegisterTriggerServiceHandler(triggerHandler, triggerSvc)
	connecthttp.Mount(sc.Mux, triggerHandler, connecthttp.WithReadMaxBytes(0))
	triggerPath := "/" + triggerconnect.TriggerServiceName + "/"
	logger.Infof(ctx, "Mounted TriggerService at %s", triggerPath)

	domains := make([]*projectpb.Domain, 0, len(cfg.Domains))
	for _, d := range cfg.Domains {
		domains = append(domains, &projectpb.Domain{
			Id:   d.ID,
			Name: d.Name,
		})
	}
	projectSvc := service.NewProjectService(impl.NewProjectRepo(sc.DB), domains)
	projectHandler := connect.NewServer(otelInterceptor)
	projectconnect.RegisterProjectServiceHandler(projectHandler, projectSvc)
	connecthttp.Mount(sc.Mux, projectHandler, connecthttp.WithReadMaxBytes(0))
	projectPath := "/" + projectconnect.ProjectServiceName + "/"
	logger.Infof(ctx, "Mounted ProjectService at %s", projectPath)

	settingsSvc := service.NewSettingsService(settingsRepo)
	settingsHandler := connect.NewServer(otelInterceptor)
	settingsconnect.RegisterSettingsServiceHandler(settingsHandler, settingsSvc)
	connecthttp.Mount(sc.Mux, settingsHandler, connecthttp.WithReadMaxBytes(0))
	settingsPath := "/" + settingsconnect.SettingsServiceName + "/"
	logger.Infof(ctx, "Mounted SettingsService at %s", settingsPath)

	if err := seedProjects(ctx, impl.NewProjectRepo(sc.DB), cfg.SeedProjects); err != nil {
		return fmt.Errorf("runs: failed to seed projects: %w", err)
	}

	if cfg.TriggerScheduler.Enabled {
		worker := scheduler.Start(ctx, repo.TriggerRepo(), cfg.TriggerScheduler, http.DefaultClient, runsServiceURL, otelClientInterceptor)
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
