// Commands for FlytePropeller manager.
package cmd

import (
	"context"
	"flag"
	"os"
	"runtime"

	"github.com/flyteorg/flytestdlib/config"
	"github.com/flyteorg/flytestdlib/config/viper"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/profutils"
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/flyteorg/flytestdlib/version"

	"github.com/flyteorg/flytepropeller/manager"
	managerConfig "github.com/flyteorg/flytepropeller/manager/config"
	propellerConfig "github.com/flyteorg/flytepropeller/pkg/controller/config"
	"github.com/flyteorg/flytepropeller/pkg/signals"
	"github.com/flyteorg/flytepropeller/pkg/utils"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
)

const (
	appName             = "flytepropeller-manager"
	podDefaultNamespace = "flyte"
	podNameEnvVar       = "POD_NAME"
	podNamespaceEnvVar  = "POD_NAMESPACE"
)

var (
	cfgFile        string
	configAccessor = viper.NewAccessor(config.Options{StrictMode: true})
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   appName,
	Short: "Runs FlytePropeller Manager to scale out FlytePropeller by executing multiple instances configured according to the defined sharding scheme.",
	Long: `
FlytePropeller Manager is used to effectively scale out FlyteWorkflow processing among a collection of FlytePropeller instances. Users configure a sharding mechanism (ex. 'hash', 'project', or 'domain') to define the sharding environment.

The FlytePropeller Manager uses a kubernetes PodTemplate to construct the base FlytePropeller PodSpec. This means, apart from the configured sharding scheme, all managed FlytePropeller instances will be identical.

The Manager ensures liveness and correctness by periodically scanning kubernets pods and recovering state (ie. starting missing pods, etc). Live configuration updates are currently unsupported, meaning configuration changes require an application restart.

Sample configuration, illustrating 3 separate sharding techniques, is provided below: 

      manager:
        pod-application: "flytepropeller"
        pod-namespace: "flyte"
        pod-template-name: "flytepropeller-template"
        pod-template-namespace: "flyte"
        scan-interval: 10s
        shard:
          # distribute FlyteWorkflow processing over 3 machines evenly
          type: hash
          pod-count: 3

		  # process the specified projects on defined replicas and all uncovered projects on another
          type: project
		  enableUncoveredReplica: true
          replicas:
            - entities:
              - flytesnacks
            - entities:
              - flyteexamples
              - flytelab

		  # process the 'production' domain on a single instace and all other domains on another
          type: domain
		  enableUncoveredReplica: true
          replicas:
            - entities:
              - production
	`,
	PersistentPreRunE: initConfig,
	Run: func(cmd *cobra.Command, args []string) {
		executeRootCmd(propellerConfig.GetConfig(), managerConfig.GetConfig())
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	version.LogBuildInformation(appName)
	logger.Infof(context.TODO(), "detected %d CPU's\n", runtime.NumCPU())
	if err := rootCmd.Execute(); err != nil {
		logger.Error(context.TODO(), err)
		os.Exit(1)
	}
}

func init() {
	// allows `$ flytepropeller-manager --logtostderr` to work
	klog.InitFlags(flag.CommandLine)
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	err := flag.CommandLine.Parse([]string{})
	if err != nil {
		logAndExit(err)
	}

	// Here you will define your flags and configuration settings. Cobra supports persistent flags, which, if defined
	// here, will be global for your application.
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "",
		"config file (default is $HOME/config.yaml)")

	configAccessor.InitializePflags(rootCmd.PersistentFlags())

	rootCmd.AddCommand(viper.GetConfigCommand())
}

func initConfig(cmd *cobra.Command, _ []string) error {
	configAccessor = viper.NewAccessor(config.Options{
		StrictMode:  false,
		SearchPaths: []string{cfgFile},
	})

	configAccessor.InitializePflags(cmd.PersistentFlags())

	err := configAccessor.UpdateConfig(context.TODO())
	if err != nil {
		return err
	}

	return nil
}

func logAndExit(err error) {
	logger.Error(context.Background(), err)
	os.Exit(-1)
}

func executeRootCmd(propellerCfg *propellerConfig.Config, cfg *managerConfig.Config) {
	baseCtx := context.Background()

	// set up signals so we handle the first shutdown signal gracefully
	ctx := signals.SetupSignalHandler(baseCtx)

	// lookup owner reference
	kubeClient, _, err := utils.GetKubeConfig(ctx, propellerCfg)
	if err != nil {
		logger.Fatalf(ctx, "error building kubernetes clientset [%v]", err)
	}

	ownerReferences := make([]metav1.OwnerReference, 0)
	lookupOwnerReferences := true
	podName, found := os.LookupEnv(podNameEnvVar)
	if !found {
		lookupOwnerReferences = false
	}

	podNamespace, found := os.LookupEnv(podNamespaceEnvVar)
	if !found {
		lookupOwnerReferences = false
		podNamespace = podDefaultNamespace
	}

	if lookupOwnerReferences {
		p, err := kubeClient.CoreV1().Pods(podNamespace).Get(ctx, podName, metav1.GetOptions{})
		if err != nil {
			logger.Fatalf(ctx, "failed to get pod '%v' in namespace '%v' [%v]", podName, podNamespace, err)
		}

		for _, ownerReference := range p.OwnerReferences {
			// must set owner reference controller to false because k8s does not allow setting pod
			// owner references to a controller that does not acknowledge ownership. in this case
			// the owner is technically the FlytePropeller Manager pod and not that pods owner.
			*ownerReference.BlockOwnerDeletion = false
			*ownerReference.Controller = false

			ownerReferences = append(ownerReferences, ownerReference)
		}
	}

	// Add the propeller_manager subscope because the MetricsPrefix only has "flyte:" to get uniform collection of metrics.
	scope := promutils.NewScope(propellerCfg.MetricsPrefix).NewSubScope("propeller_manager")

	go func() {
		err := profutils.StartProfilingServerWithDefaultHandlers(ctx, propellerCfg.ProfilerPort.Port, nil)
		if err != nil {
			logger.Panicf(ctx, "failed to start profiling and metrics server [%v]", err)
		}
	}()

	m, err := manager.New(ctx, propellerCfg, cfg, podNamespace, ownerReferences, kubeClient, scope)
	if err != nil {
		logger.Fatalf(ctx, "failed to start manager [%v]", err)
	} else if m == nil {
		logger.Fatalf(ctx, "failed to start manager, nil manager received")
	}

	if err = m.Run(ctx); err != nil {
		logger.Fatalf(ctx, "error running manager [%v]", err)
	}
}
