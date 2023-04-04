package auth

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"k8s.io/client-go/rest"

	"github.com/flyteorg/flytestdlib/logger"
	kubeErrors "k8s.io/apimachinery/pkg/api/errors"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/flyteorg/flyteadmin/pkg/config"
	executioncluster "github.com/flyteorg/flyteadmin/pkg/executioncluster/impl"
	"github.com/flyteorg/flyteadmin/pkg/executioncluster/interfaces"
	"github.com/flyteorg/flyteadmin/pkg/runtime"
	"github.com/flyteorg/flytestdlib/errors"
	"github.com/flyteorg/flytestdlib/promutils"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/client-go/kubernetes"
)

const (
	PodNamespaceEnvVar  = "POD_NAMESPACE"
	podDefaultNamespace = "default"
)

var (
	secretName       string
	secretsLocalPath string
	forceUpdate      bool
)

func GetCreateSecretsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use: "create",
		Long: `Creates a new secret (or noop if one exists unless --force is provided) using keys found in the provided path.
If POD_NAMESPACE env var is set, the secret will be created in that namespace.
`,
		Example: `
Create a secret using default name (flyte-admin-auth) in default namespace
flyteadmin secret create --fromPath=/path/in/container

Override an existing secret if one exists (reads secrets from default path /etc/secrets/):
flyteadmin secret create --name "my-auth-secrets" --force
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return persistSecrets(context.Background(), cmd.Flags())
		},
	}
	cmd.Flags().StringVar(&secretName, "name", "flyte-admin-auth", "Chooses secret name to create/update")
	cmd.Flags().StringVar(&secretsLocalPath, "fromPath", filepath.Join(string(os.PathSeparator), "etc", "secrets"), "Chooses secret name to create/update")
	cmd.Flags().BoolVarP(&forceUpdate, "force", "f", false, "Whether to update the secret if one exists")

	return cmd
}

func persistSecrets(ctx context.Context, _ *pflag.FlagSet) error {
	configuration := runtime.NewConfigurationProvider()
	scope := promutils.NewScope(configuration.ApplicationConfiguration().GetTopLevelConfig().MetricsScope)
	initializationErrorCounter := scope.NewSubScope("secrets").MustNewCounter(
		"flyteclient_initialization_error",
		"count of errors encountered initializing a flyte client from kube config")

	var listTargetsProvider interfaces.ListTargetsInterface
	var err error
	if len(configuration.ClusterConfiguration().GetClusterConfigs()) == 0 {
		serverConfig := config.GetConfig()
		listTargetsProvider, err = executioncluster.NewInCluster(initializationErrorCounter, serverConfig.KubeConfig, serverConfig.Master)
	} else {
		listTargetsProvider, err = executioncluster.NewListTargets(initializationErrorCounter, executioncluster.NewExecutionTargetProvider(), configuration.ClusterConfiguration())
	}
	if err != nil {
		return err
	}

	targets := listTargetsProvider.GetValidTargets()
	// Since we are targeting the cluster Admin is running in, this list should contain exactly one item
	if len(targets) != 1 {
		return fmt.Errorf("expected exactly 1 valid target cluster. Found [%v]", len(targets))
	}
	var clusterCfg rest.Config
	for _, target := range targets {
		// We've just ascertained targets contains exactly 1 item, so we can safely assume we'll assign the clusterCfg
		// from that one item now.
		clusterCfg = target.Config
	}

	kubeClient, err := kubernetes.NewForConfig(&clusterCfg)
	if err != nil {
		return errors.Wrapf("INIT", err, "Error building kubernetes clientset")
	}

	podNamespace, found := os.LookupEnv(PodNamespaceEnvVar)
	if !found {
		podNamespace = podDefaultNamespace
	}

	secretsData, err := buildK8sSecretData(ctx, secretsLocalPath)
	if err != nil {
		return errors.Wrapf("INIT", err, "Error building k8s secret's data field.")
	}

	secretsClient := kubeClient.CoreV1().Secrets(podNamespace)
	newSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: podNamespace,
		},
		Type: corev1.SecretTypeOpaque,
		Data: secretsData,
	}

	_, err = secretsClient.Create(ctx, newSecret, metav1.CreateOptions{})

	if err != nil && kubeErrors.IsAlreadyExists(err) {
		if forceUpdate {
			logger.Infof(ctx, "A secret already exists with the same name. Attempting to update it.")
			_, err = secretsClient.Update(ctx, newSecret, metav1.UpdateOptions{})
		} else {
			var existingSecret *corev1.Secret
			existingSecret, err = secretsClient.Get(ctx, newSecret.Name, metav1.GetOptions{})
			if err != nil {
				logger.Infof(ctx, "Failed to retrieve existing secret. Error: %v", err)
				return err
			}

			if existingSecret.Data == nil {
				existingSecret.Data = map[string][]byte{}
			}

			needsUpdate := false
			for key, val := range secretsData {
				if _, found := existingSecret.Data[key]; !found {
					existingSecret.Data[key] = val
					needsUpdate = true
				}
			}

			if needsUpdate {
				_, err = secretsClient.Update(ctx, existingSecret, metav1.UpdateOptions{})
				if err != nil && kubeErrors.IsConflict(err) {
					logger.Infof(ctx, "Another instance of flyteadmin has updated the same secret. Ignoring this update")
					err = nil
				}
			}
		}

		return err
	}

	return err
}

func buildK8sSecretData(_ context.Context, localPath string) (map[string][]byte, error) {
	secretsData := make(map[string][]byte, 4)

	err := filepath.Walk(localPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		data, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		secretsData[strings.TrimPrefix(path, filepath.Dir(path)+string(filepath.Separator))] = data
		return nil
	})

	if err != nil {
		return nil, err
	}

	return secretsData, nil
}
