package register

import (
	"context"
	"encoding/json"
	"fmt"
	cmdCore "github.com/lyft/flytectl/cmd/core"
	"github.com/lyft/flytectl/pkg/printer"
	"github.com/lyft/flytestdlib/logger"
	"io/ioutil"
	"sort"
)

const(
	registerFilesShort = "Registers file resources"
	registerFilesLong  = `
Registers all the serialized protobuf files including tasks, workflows and launchplans with default v1 version.
If there are already registered entities with v1 version then the command will fail immediately on the first such encounter.
::

 bin/flytectl register file  _pb_output/* -d development  -p flytesnacks

If you want to continue executing registration on other files ignoring the errors including version conflicts then pass in
the skipOnError flag.

::

 bin/flytectl register file  _pb_output/* -d development  -p flytesnacks --skipOnError

Using short format of skipOnError flag
::

 bin/flytectl register file  _pb_output/* -d development  -p flytesnacks -s

Overriding the default version v1 using version string.
::

 bin/flytectl register file  _pb_output/* -d development  -p flytesnacks -v v2

Change the o/p format has not effect on registration. The O/p is currently available only in table format.

::

 bin/flytectl register file  _pb_output/* -d development  -p flytesnacks -s -o yaml

Usage
`
)

func registerFromFilesFunc(ctx context.Context, args []string, cmdCtx cmdCore.CommandContext) error {
	files := args
	sort.Strings(files)
	logger.Infof(ctx, "Parsing files... Total(%v)", len(files))
	logger.Infof(ctx, "Params version %v", filesConfig.version)
	var registerResults [] RegisterResult
	adminPrinter := printer.Printer{}
	fastFail := !filesConfig.skipOnError
	logger.Infof(ctx, "Fail fast %v", fastFail)
	var _err error
	for i := 0; i< len(files) && !(fastFail && _err != nil) ; i++ {
		absFilePath := files[i]
		var registerResult RegisterResult
		logger.Infof(ctx, "Parsing  %v", absFilePath)
		fileContents, err := ioutil.ReadFile(absFilePath)
		if err != nil {
			registerResult =  RegisterResult{Name: absFilePath, Status: "Failed", Info: fmt.Sprintf("Error reading file due to %v", err)}
			registerResults = append(registerResults, registerResult)
			_err = err
			continue
		}
		spec, err := unMarshalContents(ctx, fileContents, absFilePath)
		if err != nil {
			registerResult =  RegisterResult{Name: absFilePath, Status: "Failed", Info: fmt.Sprintf("Error unmarshalling file due to %v", err)}
			registerResults = append(registerResults, registerResult)
			_err = err
			continue
		}
		if err := hydrateSpec(spec); err != nil {
			registerResult =  RegisterResult{Name: absFilePath, Status: "Failed", Info: fmt.Sprintf("Error hydrating spec due to %v", err)}
			registerResults = append(registerResults, registerResult)
			_err = err
			continue
		}
		logger.Debugf(ctx, "Hydrated spec : %v", getJsonSpec(spec))
		if err := register(ctx, spec, cmdCtx); err != nil {
			registerResult =  RegisterResult{Name: absFilePath, Status: "Failed", Info: fmt.Sprintf("Error registering file due to %v", err)}
			registerResults = append(registerResults, registerResult)
			_err = err
			continue
		}
		registerResult =  RegisterResult{Name: absFilePath, Status: "Success", Info: "Successfully registered file"}
		registerResults = append(registerResults, registerResult)
	}
	payload, _ := json.Marshal(registerResults)
	adminPrinter.JSONToTable(payload, projectColumns)
	return nil
}
