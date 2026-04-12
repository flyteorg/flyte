package runs

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/flyteorg/flyte/v2/flytestdlib/app"
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
	"github.com/flyteorg/flyte/v2/runs/scheduler"
	"github.com/flyteorg/flyte/v2/runs/service"
	authservice "github.com/flyteorg/flyte/v2/runs/service/auth"
	"github.com/flyteorg/flyte/v2/runs/service/auth/authzserver"
	authConfig "github.com/flyteorg/flyte/v2/runs/service/auth/config"

	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
)

// authSecretsDir is the directory where cookie hash/block keys and other auth
// secrets are mounted (matches the flyte-binary chart volume mount path).
const authSecretsDir = "/etc/secrets"

// Setup registers Run and Task service handlers on the SetupContext mux.
// Requires sc.DB and sc.DataStore to be set. When sc.K8sConfig is provided,
// RunLogsService is also mounted to enable pod log streaming.
func Setup(ctx context.Context, sc *app.SetupContext) error {
	cfg := config.GetConfig()
	if err := migrations.RunMigrations(ctx, sc.DB); err != nil {
		return fmt.Errorf("runs: failed to run migrations: %w", err)
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
	)

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

	runsSvc := service.NewRunService(repo, actionsClient, cfg.StoragePrefix, sc.DataStore, abortReconciler)
	taskSvc := service.NewTaskService(repo)

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

	if cfg.Security.UseAuth {
		if err := setupAuth(ctx, sc); err != nil {
			return fmt.Errorf("runs: failed to set up auth: %w", err)
		}
	} else {
		// When auth is disabled, still mount a stub AuthMetadataService so
		// clients performing metadata discovery get a coherent response.
		authMetadataSvc := service.NewAuthMetadataService(sc.BaseURL)
		authMetadataPath, authMetadataHandler := authconnect.NewAuthMetadataServiceHandler(authMetadataSvc)
		sc.Mux.Handle(authMetadataPath, authMetadataHandler)
		logger.Infof(ctx, "Mounted stub AuthMetadataService at %s", authMetadataPath)
	}

	appSvc := service.NewAppService()
	appPath, appHandler := flyteappconnect.NewAppServiceHandler(appSvc)
	sc.Mux.Handle(appPath, appHandler)
	logger.Infof(ctx, "Mounted AppService at %s", appPath)

	triggerSvc := service.NewTriggerService(repo)
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

	if cfg.TriggerScheduler.Enabled {
		runsURL := cfg.ActionsServiceURL
		if sc.BaseURL != "" {
			runsURL = sc.BaseURL
		}
		worker := scheduler.Start(ctx, repo.TriggerRepo(), cfg.TriggerScheduler, runsURL)
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

// setupAuth wires up the external-mode OAuth2 resource server, OIDC browser
// handlers, AuthMetadataService / IdentityService, and a bearer-token
// validating HTTP middleware on the shared mux. It requires that the auth
// config section is populated and that cookie hash/block keys are present as
// files under authSecretsDir.
func setupAuth(ctx context.Context, sc *app.SetupContext) error {
	authCfg := authConfig.GetConfig()

	// Mount the real AuthMetadataService backed by the configured issuer.
	authMetadataSvc := authzserver.NewAuthMetadataService(*authCfg)
	authPath, authHandler := authconnect.NewAuthMetadataServiceHandler(authMetadataSvc)
	sc.Mux.Handle(authPath, authHandler)
	logger.Infof(ctx, "Mounted AuthMetadataService at %s", authPath)

	identitySvc := authservice.NewUserInfoProvider()
	identityPath, identityHandler := authconnect.NewIdentityServiceHandler(identitySvc)
	sc.Mux.Handle(identityPath, identityHandler)
	logger.Infof(ctx, "Mounted IdentityService at %s", identityPath)

	hashKey, err := readSecretFile(authConfig.SecretNameCookieHashKey)
	if err != nil {
		return err
	}
	blockKey, err := readSecretFile(authConfig.SecretNameCookieBlockKey)
	if err != nil {
		return err
	}

	// Load the OIDC client secret used during the OAuth2 code exchange. The
	// filename is configurable so that a deployment can swap the secret name
	// without redeploying the binary.
	oidcClientSecretName := authCfg.UserAuth.OpenID.ClientSecretName
	if oidcClientSecretName == "" {
		oidcClientSecretName = authConfig.SecretNameOIdCClientSecret
	}
	oidcClientSecret, err := readSecretFile(oidcClientSecretName)
	if err != nil {
		return err
	}

	// Validate tokens issued by the configured external authorization server.
	// If BaseURL is empty, the resource server falls back to the first authorizedUri.
	var fallbackURL authConfig.URL
	if len(authCfg.AuthorizedURIs) > 0 {
		fallbackURL = authCfg.AuthorizedURIs[0]
	}
	resourceServer, err := authzserver.NewOAuth2ResourceServer(ctx, authCfg.AppAuth.ExternalAuthServer, fallbackURL)
	if err != nil {
		return fmt.Errorf("failed to create OAuth2 resource server: %w", err)
	}

	authCtx, err := authservice.NewAuthContext(ctx, *authCfg, resourceServer, hashKey, blockKey, oidcClientSecret)
	if err != nil {
		return fmt.Errorf("failed to create auth context: %w", err)
	}

	// Register /login, /callback, /logout, /.well-known/openid-configuration.
	authservice.RegisterHandlers(ctx, sc.Mux, authCtx.HandlerConfig())
	logger.Infof(ctx, "Registered OIDC browser handlers (/login, /callback, /logout)")

	// Chain the bearer/cookie auth middleware with any existing middleware
	// (e.g. CORS). Ordering: request -> CORS -> auth -> mux.
	prev := sc.Middleware
	authMw := authservice.GetAuthenticationHTTPInterceptor(authCtx.HandlerConfig())
	sc.Middleware = func(next http.Handler) http.Handler {
		wrapped := authMw(next)
		if prev != nil {
			wrapped = prev(wrapped)
		}
		return wrapped
	}
	logger.Infof(ctx, "Auth middleware installed; audience=%s", authCfg.AppAuth.ExternalAuthServer.BaseURL.String())

	return nil
}

// readSecretFile reads a base64-encoded key file from authSecretsDir and
// returns the trimmed string contents.
func readSecretFile(name string) (string, error) {
	path := filepath.Join(authSecretsDir, name)
	b, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read auth secret %s: %w", path, err)
	}
	return strings.TrimSpace(string(b)), nil
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
