package cmd

import (
	"context"

	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	ctrlWebhook "sigs.k8s.io/controller-runtime/pkg/webhook"

	"github.com/flyteorg/flyte/flytepropeller/pkg/controller"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/config"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/executors"
	"github.com/flyteorg/flyte/flytepropeller/pkg/signals"
	"github.com/flyteorg/flyte/flytepropeller/pkg/webhook"
	webhookConfig "github.com/flyteorg/flyte/flytepropeller/pkg/webhook/config"
	"github.com/flyteorg/flyte/flytestdlib/contextutils"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/profutils"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
	"github.com/flyteorg/flyte/flytestdlib/promutils/labeled"
)

var webhookCmd = &cobra.Command{
	Use:     "webhook",
	Aliases: []string{"webhooks"},
	Short:   "Runs Propeller Pod Webhook that listens for certain labels and modify the pod accordingly.",
	Long: `
This command initializes propeller's Pod webhook that enables it to mutate pods whether they are created directly from
plugins or indirectly through the creation of other CRDs (e.g. Spark/Pytorch). 
In order to use this Webhook:
1) Keys need to be mounted to the POD that runs this command; tls.crt should be a CA-issued cert (not a self-signed 
   cert), tls.key as the private key for that cert and, optionally, ca.crt in case tls.crt's CA is not a known 
   Certificate Authority (e.g. in case ca.crt is self-issued).
2) POD_NAME and POD_NAMESPACE environment variables need to be populated because the webhook initialization will lookup
   this pod to copy OwnerReferences into the new MutatingWebhookConfiguration object it'll create to ensure proper
   cleanup.

A sample Container for this webhook might look like this:

      volumes:
        - name: config-volume
          configMap:
            name: flyte-propeller-config-492gkfhbgk
        # Certs secret created by running 'flytepropeller webhook init-certs' 
        - name: webhook-certs
          secret:
            secretName: flyte-pod-webhook
      containers:
        - name: webhook-server
          image: <image>
          command:
            - flytepropeller
          args:
            - webhook
            - --config
            - /etc/flyte/config/*.yaml
          env:
            - name: POD_NAME
              valueFrom:
                fieldRef:
                  fieldPath: metadata.name
            - name: POD_NAMESPACE
              valueFrom:
                fieldRef:
                  fieldPath: metadata.namespace
          volumeMounts:
            - name: config-volume
              mountPath: /etc/flyte/config
              readOnly: true
            # Mount certs from a secret
            - name: webhook-certs
              mountPath: /etc/webhook/certs
              readOnly: true
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runWebhook(context.Background(), config.GetConfig(), webhookConfig.GetConfig())
	},
}

func init() {
	rootCmd.AddCommand(webhookCmd)
}

func runWebhook(origContext context.Context, propellerCfg *config.Config, cfg *webhookConfig.Config) error {
	// set up signals so we handle the first shutdown signal gracefully
	ctx := signals.SetupSignalHandler(origContext)

	// set metric keys
	keys := contextutils.MetricKeysFromStrings(propellerCfg.MetricKeys)
	logger.Infof(context.TODO(), "setting metrics keys to %+v", keys)
	if len(keys) > 0 {
		labeled.SetMetricKeys(keys...)
	}

	webhookScope := promutils.NewScope(cfg.MetricsPrefix).NewSubScope("webhook")
	var namespaceConfigs map[string]cache.Config
	if propellerCfg.LimitNamespace != defaultNamespace {
		namespaceConfigs = map[string]cache.Config{
			propellerCfg.LimitNamespace: {},
		}
	}

	options := manager.Options{
		Cache: cache.Options{
			SyncPeriod:        &propellerCfg.DownstreamEval.Duration,
			DefaultNamespaces: namespaceConfigs,
		},
		NewCache:  executors.NewCache,
		NewClient: executors.BuildNewClientFunc(webhookScope),
		Metrics: metricsserver.Options{
			// Disable metrics serving
			BindAddress: "0",
		},
		WebhookServer: ctrlWebhook.NewServer(ctrlWebhook.Options{
			CertDir: cfg.ExpandCertDir(),
			Port:    cfg.ListenPort,
		}),
	}

	mgr, err := controller.CreateControllerManager(ctx, propellerCfg, options)
	if err != nil {
		logger.Fatalf(ctx, "Failed to create controller manager. Error: %v", err)
		return err
	}

	g, childCtx := errgroup.WithContext(ctx)
	g.Go(func() error {
		err := profutils.StartProfilingServerWithDefaultHandlers(childCtx, propellerCfg.ProfilerPort.Port, nil)
		if err != nil {
			logger.Fatalf(childCtx, "Failed to Start profiling and metrics server. Error: %v", err)
		}
		return err
	})

	g.Go(func() error {
		err := controller.StartControllerManager(childCtx, mgr)
		if err != nil {
			logger.Fatalf(childCtx, "Failed to start controller manager. Error: %v", err)
		}
		return err
	})

	g.Go(func() error {
		err := webhook.Run(childCtx, propellerCfg, cfg, defaultNamespace, &webhookScope, mgr)
		if err != nil {
			logger.Fatalf(childCtx, "Failed to start webhook. Error: %v", err)
		}
		return err
	})

	return g.Wait()
}
