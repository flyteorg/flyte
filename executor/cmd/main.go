package main

import (
	"context"
	"fmt"
	"os"

	"github.com/flyteorg/flyte/v2/app"
	"github.com/flyteorg/flyte/v2/executor"
	executorconfig "github.com/flyteorg/flyte/v2/executor/pkg/config"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/metrics/filters"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	flyteorgv1 "github.com/flyteorg/flyte/v2/executor/api/v1"
	"github.com/flyteorg/flyte/v2/executor/pkg/config"
	"github.com/flyteorg/flyte/v2/executor/pkg/controller"
	"github.com/flyteorg/flyte/v2/executor/pkg/plugin"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery"
	"github.com/flyteorg/flyte/v2/flytestdlib/promutils"
	"github.com/flyteorg/flyte/v2/flytestdlib/storage"

	// Plugin registrations -- blank imports trigger init() which registers plugins with the global registry.
	_ "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/plugins/k8s/pod"
	stdconfig "github.com/flyteorg/flyte/v2/flytestdlib/config"
	"github.com/flyteorg/flyte/v2/flytestdlib/config/viper"
)

func main() {
	a := &app.App{
		Name:  "executor",
		Short: "Executor controller manager for Flyte TaskActions",
		Setup: func(ctx context.Context, sc *app.SetupContext) error {
			cfg := executorconfig.GetConfig()

			// Executor doesn't serve HTTP — it uses controller-runtime's own
			// health probe port. Set a dummy port so the app skeleton starts
			// its HTTP server on a non-conflicting address (or 0 to disable).
			sc.Port = 0

			k8sConfig := ctrl.GetConfigOrDie()
			sc.K8sConfig = k8sConfig

			if err := executor.Setup(ctx, sc); err != nil {
				return fmt.Errorf("executor setup failed: %w", err)
			}

			_ = cfg // config is read inside executor.Setup's worker
			return nil
		},
	}
	if err := a.Run(); err != nil {
		os.Exit(1)
	}
}

func initConfig(cmd *cobra.Command) error {
	configAccessor = viper.NewAccessor(stdconfig.Options{
		SearchPaths: []string{cfgFile, ".", "/etc/flyte/config"},
		StrictMode:  false,
	})

	// Traverse to root command
	rootCmd := cmd
	for rootCmd.Parent() != nil {
		rootCmd = rootCmd.Parent()
	}

	configAccessor.InitializePflags(rootCmd.PersistentFlags())

	return configAccessor.UpdateConfig(context.Background())
}

// nolint:gocyclo
func run() error {
	cfg := config.GetConfig()

	opts := zap.Options{
		Development: true,
	}
	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	var tlsOpts []func(*tls.Config)

	// if the enable-http2 flag is false (the default), http/2 should be disabled
	// due to its vulnerabilities. More specifically, disabling http/2 will
	// prevent from being vulnerable to the HTTP/2 Stream Cancellation and
	// Rapid Reset CVEs. For more information see:
	// - https://github.com/advisories/GHSA-qppj-fm5r-hxr3
	// - https://github.com/advisories/GHSA-4374-p667-p6c8
	disableHTTP2 := func(c *tls.Config) {
		setupLog.Info("disabling http/2")
		c.NextProtos = []string{"http/1.1"}
	}

	if !cfg.EnableHTTP2 {
		tlsOpts = append(tlsOpts, disableHTTP2)
	}

	// Initial webhook TLS options
	webhookTLSOpts := tlsOpts
	webhookServerOptions := webhook.Options{
		TLSOpts: webhookTLSOpts,
	}

	if len(cfg.WebhookCertPath) > 0 {
		setupLog.Info("Initializing webhook certificate watcher using provided certificates",
			"webhook-cert-path", cfg.WebhookCertPath, "webhook-cert-name", cfg.WebhookCertName, "webhook-cert-key", cfg.WebhookCertKey)

		webhookServerOptions.CertDir = cfg.WebhookCertPath
		webhookServerOptions.CertName = cfg.WebhookCertName
		webhookServerOptions.KeyName = cfg.WebhookCertKey
	}

	webhookServer := webhook.NewServer(webhookServerOptions)

	// Metrics endpoint is enabled in 'config/default/kustomization.yaml'. The Metrics options configure the server.
	// More info:
	// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.22.1/pkg/metrics/server
	// - https://book.kubebuilder.io/reference/metrics.html
	metricsServerOptions := metricsserver.Options{
		BindAddress:   cfg.MetricsBindAddress,
		SecureServing: cfg.MetricsSecure,
		TLSOpts:       tlsOpts,
	}

	if cfg.MetricsSecure {
		// FilterProvider is used to protect the metrics endpoint with authn/authz.
		// These configurations ensure that only authorized users and service accounts
		// can access the metrics endpoint. The RBAC are configured in 'config/rbac/kustomization.yaml'. More info:
		// https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.22.1/pkg/metrics/filters#WithAuthenticationAndAuthorization
		metricsServerOptions.FilterProvider = filters.WithAuthenticationAndAuthorization
	}

	// If the certificate is not specified, controller-runtime will automatically
	// generate self-signed certificates for the metrics server. While convenient for development and testing,
	// this setup is not recommended for production.
	if len(cfg.MetricsCertPath) > 0 {
		setupLog.Info("Initializing metrics certificate watcher using provided certificates",
			"metrics-cert-path", cfg.MetricsCertPath, "metrics-cert-name", cfg.MetricsCertName, "metrics-cert-key", cfg.MetricsCertKey)

		metricsServerOptions.CertDir = cfg.MetricsCertPath
		metricsServerOptions.CertName = cfg.MetricsCertName
		metricsServerOptions.KeyName = cfg.MetricsCertKey
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		Metrics:                metricsServerOptions,
		WebhookServer:          webhookServer,
		HealthProbeBindAddress: cfg.HealthProbeBindAddress,
		LeaderElection:         cfg.LeaderElect,
		LeaderElectionID:       "abf369a8.flyte.org",
		// LeaderElectionReleaseOnCancel defines if the leader should step down voluntarily
		// when the Manager ends. This requires the binary to immediately end when the
		// Manager is stopped, otherwise, this setting is unsafe. Setting this significantly
		// speeds up voluntary leader transitions as the new leader don't have to wait
		// LeaseDuration time first.
		//
		// In the default scaffold provided, the program ends immediately after
		// the manager stops, so would be fine to enable this option. However,
		// if you are doing or is intended to do any operation such as perform cleanups
		// after the manager stops then its usage might be unsafe.
		// LeaderElectionReleaseOnCancel: true,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	// Create DataStore for plugin I/O (task templates, inputs, outputs)
	dataStore, err := storage.NewDataStore(storage.GetConfig(), promutils.NewScope("executor:storage"))
	if err != nil {
		setupLog.Error(err, "unable to create data store")
		os.Exit(1)
	}

	// Create SetupContext for plugin initialization.
	// ctrl.Manager satisfies pluginsCore.KubeClient (GetClient + GetCache).
	setupCtx := plugin.NewSetupContext(
		mgr,  // KubeClient
		nil,  // SecretManager -- TODO: implement
		nil,  // ResourceRegistrar -- not needed for executor
		nil,  // EnqueueOwner -- not needed, controller-runtime handles reconciliation
		nil,  // EnqueueLabels
		"TaskAction",
		promutils.NewScope("executor"),
	)

	// Initialize plugin registry from the global pluginmachinery singleton.
	// Plugins are registered via blank imports above (e.g. pod plugin).
	registry := plugin.NewRegistry(setupCtx, pluginmachinery.PluginRegistry())
	if err := registry.Initialize(context.Background()); err != nil {
		setupLog.Error(err, "unable to initialize plugin registry")
		os.Exit(1)
	}

	reconciler := controller.NewTaskActionReconciler(
		mgr.GetClient(),
		mgr.GetScheme(),
		registry,
		dataStore,
	)
	reconciler.Recorder = mgr.GetEventRecorderFor("taskaction-controller")
	if err := reconciler.SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "TaskAction")
		os.Exit(1)
	}

	// +kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}

	return nil
}
