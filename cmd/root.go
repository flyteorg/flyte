package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytestdlib/config"
	"github.com/flyteorg/flytestdlib/config/viper"
	"github.com/flyteorg/flytestdlib/contextutils"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/flyteorg/flytestdlib/promutils/labeled"
	"github.com/flyteorg/flytestdlib/storage"
	"github.com/flyteorg/flytestdlib/version"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
)

type RootOptions struct {
	*clientcmd.ConfigOverrides
	showSource     bool
	clientConfig   clientcmd.ClientConfig
	Scope          promutils.Scope
	Store          *storage.DataStore
	configAccessor config.Accessor
	cfgFile        string
	// The actual key name that should be created under the remote prefix where the error document is written of the form errors.pb
	errorOutputName string
}

func (r *RootOptions) executeRootCmd() error {
	ctx := context.TODO()
	logger.Infof(ctx, "Go Version: %s", runtime.Version())
	logger.Infof(ctx, "Go OS/Arch: %s/%s", runtime.GOOS, runtime.GOARCH)
	version.LogBuildInformation("flytedata")
	return fmt.Errorf("use one of the sub-commands")
}

func (r RootOptions) UploadError(ctx context.Context, code string, recvErr error, prefix storage.DataReference) error {
	if recvErr == nil {
		recvErr = fmt.Errorf("unknown error")
	}
	errorPath, err := r.Store.ConstructReference(ctx, prefix, r.errorOutputName)
	if err != nil {
		logger.Errorf(ctx, "failed to create error file path err: %s", err)
		return err
	}
	logger.Infof(ctx, "Uploading Error file to path [%s], errFile: %s", errorPath, r.errorOutputName)
	return r.Store.WriteProtobuf(ctx, errorPath, storage.Options{}, &core.ErrorDocument{
		Error: &core.ContainerError{
			Code:    code,
			Message: recvErr.Error(),
			Kind:    core.ContainerError_RECOVERABLE,
		},
	})
}

func PollUntilTimeout(ctx context.Context, pollInterval, timeout time.Duration, condition wait.ConditionFunc) error {
	childCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	return wait.PollUntil(pollInterval, condition, childCtx.Done())
}

func checkAWSCreds() (*credentials.Value, error) {
	sess, err := session.NewSession(&aws.Config{})
	if err != nil {
		return nil, err
	}

	// Determine the AWS credentials from the default credential chain
	creds, err := sess.Config.Credentials.Get()
	if err != nil {
		return nil, err
	}
	if creds.AccessKeyID == "" || creds.SecretAccessKey == "" || creds.SessionToken == "" {
		return nil, fmt.Errorf("invalid data in credential fetch")
	}
	return &creds, nil
}

func waitForAWSCreds(ctx context.Context, timeout time.Duration) error {
	return PollUntilTimeout(ctx, time.Second*5, timeout, func() (bool, error) {
		if creds, err := checkAWSCreds(); err != nil {
			logger.Errorf(ctx, "failed to get AWS credentials: %s", err)
			return false, nil
		} else if creds != nil {
			logger.Infof(ctx, "found AWS credentials from provider: %s", creds.ProviderName)
		}
		return true, nil
	})
}

// NewCommand returns a new instance of the co-pilot root command
func NewDataCommand() *cobra.Command {
	rootOpts := &RootOptions{}
	command := &cobra.Command{
		Use:   "flytedata",
		Short: "flytedata is a simple go binary that can be used to retrieve and upload data from/to remote stow store to local disk.",
		Long:  `flytedata when used with conjunction with flytepropeller eliminates the need to have any flyte library installed inside the container`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if err := rootOpts.initConfig(cmd, args); err != nil {
				return err
			}
			rootOpts.Scope = promutils.NewScope("flyte:data")
			cfg := storage.GetConfig()
			if cfg.Type == storage.TypeS3 {
				if err := waitForAWSCreds(context.Background(), time.Minute*10); err != nil {
					return err
				}
			}
			store, err := storage.NewDataStore(cfg, rootOpts.Scope)
			if err != nil {
				return errors.Wrap(err, "failed to create datastore client")
			}
			rootOpts.Store = store
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return rootOpts.executeRootCmd()
		},
	}

	command.AddCommand(NewDownloadCommand(rootOpts))
	command.AddCommand(NewUploadCommand(rootOpts))

	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	loadingRules.DefaultClientConfig = &clientcmd.DefaultClientConfig
	rootOpts.ConfigOverrides = &clientcmd.ConfigOverrides{}
	kflags := clientcmd.RecommendedConfigOverrideFlags("")
	command.PersistentFlags().StringVar(&loadingRules.ExplicitPath, "kubeconfig", "", "Path to a kube config. Only required if out-of-cluster")
	clientcmd.BindOverrideFlags(rootOpts.ConfigOverrides, command.PersistentFlags(), kflags)
	rootOpts.clientConfig = clientcmd.NewInteractiveDeferredLoadingClientConfig(loadingRules, rootOpts.ConfigOverrides, os.Stdin)

	command.PersistentFlags().StringVar(&rootOpts.cfgFile, "config", "", "config file (default is $HOME/config.yaml)")
	command.PersistentFlags().BoolVarP(&rootOpts.showSource, "show-source", "s", false, "Show line number for errors")
	command.PersistentFlags().StringVar(&rootOpts.errorOutputName, "err-output-name", "errors.pb", "Actual key name under the prefix where the error protobuf should be written to")

	rootOpts.configAccessor = viper.NewAccessor(config.Options{StrictMode: true})
	// Here you will define your flags and configuration settings. Cobra supports persistent flags, which, if defined
	// here, will be global for your application.
	rootOpts.configAccessor.InitializePflags(command.PersistentFlags())

	command.AddCommand(viper.GetConfigCommand())

	return command
}

func (r *RootOptions) initConfig(cmd *cobra.Command, _ []string) error {
	r.configAccessor = viper.NewAccessor(config.Options{
		StrictMode:  true,
		SearchPaths: []string{r.cfgFile},
	})

	rootCmd := cmd
	for rootCmd.Parent() != nil {
		rootCmd = rootCmd.Parent()
	}

	// persistent flags were initially bound to the root command so we must bind to the same command to avoid
	r.configAccessor.InitializePflags(rootCmd.PersistentFlags())

	err := r.configAccessor.UpdateConfig(context.TODO())
	if err != nil {
		return err
	}
	return nil
}

func init() {
	klog.InitFlags(flag.CommandLine)
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	err := flag.CommandLine.Parse([]string{})
	if err != nil {
		logger.Error(context.TODO(), "Error in initializing: %v", err)
		os.Exit(-1)
	}
	labeled.SetMetricKeys(contextutils.ProjectKey, contextutils.DomainKey, contextutils.WorkflowIDKey, contextutils.TaskIDKey)
}
