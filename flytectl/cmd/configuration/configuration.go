package configuration

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/flyteorg/flytestdlib/logger"

	"github.com/flyteorg/flytectl/pkg/util"

	"github.com/flyteorg/flytectl/pkg/configutil"

	initConfig "github.com/flyteorg/flytectl/cmd/config/subcommand/config"
	cmdcore "github.com/flyteorg/flytectl/cmd/core"
	cmdUtil "github.com/flyteorg/flytectl/pkg/commandutils"
	"github.com/flyteorg/flytestdlib/config/viper"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

// Long descriptions are whitespace sensitive when generating docs using Sphinx.
const (
	initCmdShort = `Generates FlyteCTL config file in the user's home directory.`
	initCmdLong  = `Creates a FlyteCTL config file in Flyte directory i.e ~/.flyte
	
Generates sandbox config. Flyte Sandbox is a fully standalone minimal environment for running Flyte. Read more about sandbox https://docs.flyte.org/en/latest/deployment/sandbox.html

::

 flytectl config init  

Generates remote cluster config, By default connection is secure. Read more about the remote deployment https://docs.flyte.org/en/latest/deployment/index.html
	
::

 flytectl config init --host=flyte.myexample.com

Generates remote cluster config with insecure connection

::

 flytectl config init --host=flyte.myexample.com --insecure 

Generates FlyteCTL config with a storage provider
::

 flytectl config init --host=flyte.myexample.com --storage
`
)

var prompt = promptui.Select{
	Label: "Select Storage Provider",
	Items: []string{"S3", "GCS"},
}

var endpointPrefix = [3]string{"dns:///", "http://", "https://"}

// CreateConfigCommand will return configuration command
func CreateConfigCommand() *cobra.Command {
	configCmd := viper.GetConfigCommand()

	getResourcesFuncs := map[string]cmdcore.CommandEntry{
		"init": {CmdFunc: configInitFunc, Aliases: []string{""}, ProjectDomainNotRequired: true,
			Short: initCmdShort,
			Long:  initCmdLong, PFlagProvider: initConfig.DefaultConfig},
	}

	cmdcore.AddCommands(configCmd, getResourcesFuncs)
	return configCmd
}

func configInitFunc(ctx context.Context, args []string, cmdCtx cmdcore.CommandContext) error {
	return initFlytectlConfig(ctx, os.Stdin)
}

func initFlytectlConfig(ctx context.Context, reader io.Reader) error {

	if err := util.SetupFlyteDir(); err != nil {
		return err
	}

	templateValues := configutil.ConfigTemplateSpec{
		Host:     "dns:///localhost:30081",
		Insecure: true,
	}
	templateStr := configutil.GetSandboxTemplate()

	if len(initConfig.DefaultConfig.Host) > 0 {
		trimHost := trimEndpoint(initConfig.DefaultConfig.Host)
		if !validateEndpointName(trimHost) {
			return errors.New("Please use a valid endpoint")
		}
		templateValues.Host = fmt.Sprintf("dns:///%s", trimHost)
		templateValues.Insecure = initConfig.DefaultConfig.Insecure
		templateStr = configutil.AdminConfigTemplate
		if initConfig.DefaultConfig.Storage {
			templateStr = configutil.GetAWSCloudTemplate()
			_, result, err := prompt.Run()
			if err != nil {
				return err
			}
			if strings.ToUpper(result) == "GCS" {
				templateStr = configutil.GetGoogleCloudTemplate()
			}
		} else {
			logger.Infof(ctx, "Init flytectl config for remote cluster, Please update your storage config in %s. Learn more about the config here https://docs.flyte.org/projects/flytectl/en/latest/index.html#configure", configutil.ConfigFile)
		}
	}
	var _err error
	if _, err := os.Stat(configutil.ConfigFile); os.IsNotExist(err) {
		_err = configutil.SetupConfig(configutil.ConfigFile, templateStr, templateValues)
	} else {
		if cmdUtil.AskForConfirmation(fmt.Sprintf("This action will overwrite an existing config file at [%s]. Do you want to continue?", configutil.ConfigFile), reader) {
			if err := os.Remove(configutil.ConfigFile); err != nil {
				return err
			}
			_err = configutil.SetupConfig(configutil.ConfigFile, templateStr, templateValues)
		}
	}
	if _err != nil {
		return _err
	}
	fmt.Printf("Init flytectl config file at [%s]", configutil.ConfigFile)
	return nil
}

func trimEndpoint(hostname string) string {
	for _, prefix := range endpointPrefix {
		hostname = strings.TrimPrefix(hostname, prefix)
	}
	return hostname

}

func validateEndpointName(endPoint string) bool {
	var validate = false
	if endPoint == "localhost" {
		return true
	}
	if err := is.URL.Validate(endPoint); err != nil {
		return false
	}
	endPointParts := strings.Split(endPoint, ":")
	if len(endPointParts) <= 2 && len(endPointParts) > 0 {
		if err := is.DNSName.Validate(endPointParts[0]); !errors.Is(err, is.ErrDNSName) && err == nil {
			validate = true
		}
		if err := is.IP.Validate(endPointParts[0]); !errors.Is(err, is.ErrIP) && err == nil {
			validate = true
		}
		if len(endPointParts) == 2 {
			if err := is.Port.Validate(endPointParts[1]); err != nil {
				return false
			}
		}
	}

	return validate
}
