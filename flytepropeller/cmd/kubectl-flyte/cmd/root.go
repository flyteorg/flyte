package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/flyteorg/flytestdlib/config/viper"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/version"
	"github.com/spf13/pflag"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	flyteclient "github.com/flyteorg/flytepropeller/pkg/client/clientset/versioned"
	"github.com/spf13/cobra"
)

func init() {
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	err := flag.CommandLine.Parse([]string{})
	if err != nil {
		logger.Error(context.TODO(), "Error in initializing: %v", err)
		os.Exit(-1)
	}
}

type RootOptions struct {
	*clientcmd.ConfigOverrides
	allNamespaces bool
	showSource    bool
	clientConfig  clientcmd.ClientConfig
	restConfig    *rest.Config
	kubeClient    kubernetes.Interface
	flyteClient   flyteclient.Interface
}

func (r *RootOptions) GetTimeoutSeconds() (int64, error) {
	if r.Timeout != "" {
		d, err := time.ParseDuration(r.Timeout)
		if err != nil {
			return 10, err
		}
		return int64(d.Seconds()), nil
	}
	return 10, nil

}

func (r *RootOptions) executeRootCmd() error {
	ctx := context.TODO()
	logger.Infof(ctx, "Go Version: %s", runtime.Version())
	logger.Infof(ctx, "Go OS/Arch: %s/%s", runtime.GOOS, runtime.GOARCH)
	version.LogBuildInformation("kubectl-flyte")
	return fmt.Errorf("use one of the sub-commands")
}

func (r *RootOptions) ConfigureClient() error {
	restConfig, err := r.clientConfig.ClientConfig()
	if err != nil {
		return err
	}
	k, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return err
	}
	fc, err := flyteclient.NewForConfig(restConfig)
	if err != nil {
		return err
	}
	r.restConfig = restConfig
	r.kubeClient = k
	r.flyteClient = fc
	return nil
}

// NewCommand returns a new instance of an argo command
func NewFlyteCommand() *cobra.Command {
	rootOpts := &RootOptions{}
	command := &cobra.Command{
		Use:   "kubectl-flyte",
		Short: "kubectl-flyte allows launching and managing K8s native workflows",
		Long: `Flyte is a serverless workflow processing platform built for native execution on K8s.
      It is extensible and flexible to allow adding new operators and comes with many operators built in`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return rootOpts.ConfigureClient()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return rootOpts.executeRootCmd()
		},
	}

	command.AddCommand(NewDeleteCommand(rootOpts))
	command.AddCommand(NewGetCommand(rootOpts))
	command.AddCommand(NewVisualizeCommand(rootOpts))
	command.AddCommand(NewCreateCommand(rootOpts))
	command.AddCommand(NewCompileCommand(rootOpts))

	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	loadingRules.DefaultClientConfig = &clientcmd.DefaultClientConfig
	rootOpts.ConfigOverrides = &clientcmd.ConfigOverrides{}
	kflags := clientcmd.RecommendedConfigOverrideFlags("")
	command.PersistentFlags().StringVar(&loadingRules.ExplicitPath, "kubeconfig", "", "Path to a kube config. Only required if out-of-cluster")
	clientcmd.BindOverrideFlags(rootOpts.ConfigOverrides, command.PersistentFlags(), kflags)
	rootOpts.clientConfig = clientcmd.NewInteractiveDeferredLoadingClientConfig(loadingRules, rootOpts.ConfigOverrides, os.Stdin)

	command.PersistentFlags().BoolVar(&rootOpts.allNamespaces, "all-namespaces", false, "Enable this flag to execute for all namespaces")
	command.PersistentFlags().BoolVarP(&rootOpts.showSource, "show-source", "s", false, "Show line number for errors")
	command.AddCommand(viper.GetConfigCommand())

	return command
}
