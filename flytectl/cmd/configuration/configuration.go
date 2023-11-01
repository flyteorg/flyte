package configuration

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/flyteorg/flytectl/pkg/util"

	"github.com/flyteorg/flytectl/pkg/configutil"

	"github.com/flyteorg/flyte/flytestdlib/config/viper"
	initConfig "github.com/flyteorg/flytectl/cmd/config/subcommand/config"
	cmdcore "github.com/flyteorg/flytectl/cmd/core"
	cmdUtil "github.com/flyteorg/flytectl/pkg/commandutils"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/spf13/cobra"
)

// Long descriptions are whitespace sensitive when generating docs using Sphinx.
const (
	initCmdShort = `Generates a Flytectl config file in the user's home directory.`
	initCmdLong  = `Creates a Flytectl config file in Flyte directory i.e ~/.flyte.
	
Generate Sandbox config:
::

 flytectl config init  

Flyte Sandbox is a fully standalone minimal environment for running Flyte. 
Read more about the Sandbox deployment :ref:` + "`here <deploy-sandbox-local>`" + `.

Generate remote cluster config: 
::

 flytectl config init --host=flyte.myexample.com

By default, the connection is secure. 
Read more about remote deployment :ref:` + "`here <Deployment>`" + `.

Generate remote cluster config with insecure connection:
::

 flytectl config init --host=flyte.myexample.com --insecure 

 Generate remote cluster config with separate console endpoint:
 ::

  flytectl config init --host=flyte.myexample.com --console=console.myexample.com

Generate Flytectl config with a storage provider:
::

 flytectl config init --host=flyte.myexample.com --storage
`
)

var endpointPrefix = [3]string{"dns:///", "http://", "https://"}

// CreateConfigCommand will return configuration command
func CreateConfigCommand() *cobra.Command {
	configCmd := viper.GetConfigCommand()

	getResourcesFuncs := map[string]cmdcore.CommandEntry{
		"init": {CmdFunc: configInitFunc, Aliases: []string{""}, ProjectDomainNotRequired: true,
			Short: initCmdShort,
			Long:  initCmdLong, PFlagProvider: initConfig.DefaultConfig},
	}

	configCmd.Flags().BoolVar(&initConfig.DefaultConfig.Force, "force", false, "Force to overwrite the default config file without confirmation")

	cmdcore.AddCommands(configCmd, getResourcesFuncs)
	return configCmd
}

func configInitFunc(ctx context.Context, args []string, cmdCtx cmdcore.CommandContext) error {
	return initFlytectlConfig(os.Stdin)
}

func initFlytectlConfig(reader io.Reader) error {

	if err := util.SetupFlyteDir(); err != nil {
		return err
	}

	templateValues := configutil.ConfigTemplateSpec{
		Host:     "dns:///localhost:30080",
		Insecure: true,
	}
	templateStr := configutil.GetTemplate()

	if len(initConfig.DefaultConfig.Host) > 0 {
		trimHost := trimEndpoint(initConfig.DefaultConfig.Host)
		if !validateEndpointName(trimHost) {
			return fmt.Errorf("%s invalid, please use a valid admin endpoint", trimHost)
		}
		templateValues.Host = fmt.Sprintf("dns:///%s", trimHost)
		templateValues.Insecure = initConfig.DefaultConfig.Insecure
	}
	if len(initConfig.DefaultConfig.Console) > 0 {
		trimConsole := trimEndpoint(initConfig.DefaultConfig.Console)
		if !validateEndpointName(trimConsole) {
			return fmt.Errorf("%s invalid, please use a valid console endpoint", trimConsole)
		}
		templateValues.Console = initConfig.DefaultConfig.Console
	}
	var _err error
	if _, err := os.Stat(configutil.ConfigFile); os.IsNotExist(err) {
		_err = configutil.SetupConfig(configutil.ConfigFile, templateStr, templateValues)
	} else {
		if initConfig.DefaultConfig.Force || cmdUtil.AskForConfirmation(fmt.Sprintf("This action will overwrite an existing config file at [%s]. Do you want to continue?", configutil.ConfigFile), reader) {
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
