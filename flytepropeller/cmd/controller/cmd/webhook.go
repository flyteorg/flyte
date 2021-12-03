package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	webhookConfig "github.com/flyteorg/flytepropeller/pkg/webhook/config"

	apiErrors "k8s.io/apimachinery/pkg/api/errors"

	"k8s.io/client-go/kubernetes"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/flyteorg/flytepropeller/pkg/controller/executors"
	"github.com/flyteorg/flytepropeller/pkg/signals"
	"github.com/flyteorg/flytepropeller/pkg/utils"
	"github.com/flyteorg/flytepropeller/pkg/webhook"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/profutils"
	"github.com/flyteorg/flytestdlib/promutils"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/flyteorg/flytepropeller/pkg/controller/config"
	"github.com/spf13/cobra"
)

const (
	PodNameEnvVar       = "POD_NAME"
	PodNamespaceEnvVar  = "POD_NAMESPACE"
	podDefaultNamespace = "default"
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

	raw, err := json.Marshal(cfg)
	if err != nil {
		return err
	}

	fmt.Println(string(raw))

	kubeClient, kubecfg, err := utils.GetKubeConfig(ctx, propellerCfg)
	if err != nil {
		return err
	}

	// Add the propeller subscope because the MetricsPrefix only has "flyte:" to get uniform collection of metrics.
	propellerScope := promutils.NewScope(cfg.MetricsPrefix).NewSubScope("propeller").NewSubScope(safeMetricName(propellerCfg.LimitNamespace))
	webhookScope := propellerScope.NewSubScope("webhook")

	go func() {
		err := profutils.StartProfilingServerWithDefaultHandlers(ctx, propellerCfg.ProfilerPort.Port, nil)
		if err != nil {
			logger.Panicf(ctx, "Failed to Start profiling and metrics server. Error: %v", err)
		}
	}()

	limitNamespace := ""
	if propellerCfg.LimitNamespace != defaultNamespace {
		limitNamespace = propellerCfg.LimitNamespace
	}

	secretsWebhook := webhook.NewPodMutator(cfg, webhookScope)

	// Creates a MutationConfig to instruct ApiServer to call this service whenever a Pod is being created.
	err = createMutationConfig(ctx, kubeClient, secretsWebhook)
	if err != nil {
		return err
	}

	mgr, err := manager.New(kubecfg, manager.Options{
		Port:          cfg.ListenPort,
		CertDir:       cfg.CertDir,
		Namespace:     limitNamespace,
		SyncPeriod:    &propellerCfg.DownstreamEval.Duration,
		ClientBuilder: executors.NewFallbackClientBuilder(webhookScope),
	})

	if err != nil {
		logger.Fatalf(ctx, "Failed to initialize controller run-time manager. Error: %v", err)
	}

	err = secretsWebhook.Register(ctx, mgr)
	if err != nil {
		logger.Fatalf(ctx, "Failed to register webhook with manager. Error: %v", err)
	}

	logger.Infof(ctx, "Starting controller-runtime manager")
	return mgr.Start(ctx)
}

func createMutationConfig(ctx context.Context, kubeClient *kubernetes.Clientset, webhookObj *webhook.PodMutator) error {
	shouldAddOwnerRef := true
	podName, found := os.LookupEnv(PodNameEnvVar)
	if !found {
		shouldAddOwnerRef = false
	}

	podNamespace, found := os.LookupEnv(PodNamespaceEnvVar)
	if !found {
		shouldAddOwnerRef = false
		podNamespace = podDefaultNamespace
	}

	mutateConfig, err := webhookObj.CreateMutationWebhookConfiguration(podNamespace)
	if err != nil {
		return err
	}

	if shouldAddOwnerRef {
		// Lookup the pod to retrieve its UID
		p, err := kubeClient.CoreV1().Pods(podNamespace).Get(ctx, podName, metav1.GetOptions{})
		if err != nil {
			logger.Infof(ctx, "Failed to get Pod [%v/%v]. Error: %v", podNamespace, podName, err)
			return fmt.Errorf("failed to get pod. Error: %w", err)
		}

		mutateConfig.OwnerReferences = p.OwnerReferences
	}

	logger.Infof(ctx, "Creating MutatingWebhookConfiguration [%v/%v]", mutateConfig.GetNamespace(), mutateConfig.GetName())

	_, err = kubeClient.AdmissionregistrationV1().MutatingWebhookConfigurations().Create(ctx, mutateConfig, metav1.CreateOptions{})
	var statusErr *apiErrors.StatusError
	if err != nil && errors.As(err, &statusErr) && statusErr.Status().Reason == metav1.StatusReasonAlreadyExists {
		logger.Infof(ctx, "Failed to create MutatingWebhookConfiguration. Will attempt to update. Error: %v", err)
		obj, getErr := kubeClient.AdmissionregistrationV1().MutatingWebhookConfigurations().Get(ctx, mutateConfig.Name, metav1.GetOptions{})
		if getErr != nil {
			logger.Infof(ctx, "Failed to get MutatingWebhookConfiguration. Error: %v", getErr)
			return err
		}

		obj.Webhooks = mutateConfig.Webhooks
		_, err = kubeClient.AdmissionregistrationV1().MutatingWebhookConfigurations().Update(ctx, obj, metav1.UpdateOptions{})
		if err == nil {
			logger.Infof(ctx, "Successfully updated existing mutating webhook config.")
		}

		return err
	}

	return nil
}
