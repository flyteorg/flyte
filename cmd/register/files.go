package register

import (
	"context"
	"encoding/json"
	"os"

	cmdCore "github.com/lyft/flytectl/cmd/core"
	"github.com/lyft/flytectl/pkg/printer"
	"github.com/lyft/flytestdlib/logger"
)

//go:generate pflags FilesConfig
var (
	filesConfig = &FilesConfig{
		Version:         "v1",
		ContinueOnError: false,
	}
)

// FilesConfig
type FilesConfig struct {
	Version         string `json:"version" pflag:",version of the entity to be registered with flyte."`
	ContinueOnError bool   `json:"continueOnError" pflag:",continue on error when registering files."`
	Archive         bool   `json:"archive" pflag:",pass in archive file either an http link or local path."`
}

const (
	registerFilesShort = "Registers file resources"
	registerFilesLong  = `
Registers all the serialized protobuf files including tasks, workflows and launchplans with default v1 version.
If there are already registered entities with v1 version then the command will fail immediately on the first such encounter.
::

 bin/flytectl register file  _pb_output/* -d development  -p flytesnacks

Using archive file.Currently supported are .tgz and .tar extension files and can be local or remote file served through http/https.
Use --archive flag.

::

 bin/flytectl register files  http://localhost:8080/_pb_output.tar -d development  -p flytesnacks --archive

Using  local tgz file.

::

 bin/flytectl register files  _pb_output.tgz -d development  -p flytesnacks --archive

If you want to continue executing registration on other files ignoring the errors including version conflicts then pass in
the continueOnError flag.

::

 bin/flytectl register file  _pb_output/* -d development  -p flytesnacks --continueOnError

Using short format of continueOnError flag
::

 bin/flytectl register file  _pb_output/* -d development  -p flytesnacks -c

Overriding the default version v1 using version string.
::

 bin/flytectl register file  _pb_output/* -d development  -p flytesnacks -v v2

Change the o/p format has not effect on registration. The O/p is currently available only in table format.

::

 bin/flytectl register file  _pb_output/* -d development  -p flytesnacks -c -o yaml

Usage
`
)

func registerFromFilesFunc(ctx context.Context, args []string, cmdCtx cmdCore.CommandContext) error {
	dataRefs, tmpDir, _err := getSortedFileList(ctx, args)
	if _err != nil {
		logger.Errorf(ctx, "error while un-archiving files in tmp dir due to %v", _err)
		return _err
	}
	logger.Infof(ctx, "Parsing files... Total(%v)", len(dataRefs))
	fastFail := !filesConfig.ContinueOnError
	var registerResults []Result
	for i := 0; i < len(dataRefs) && !(fastFail && _err != nil); i++ {
		registerResults, _err = registerFile(ctx, dataRefs[i], registerResults, cmdCtx)
	}
	payload, _ := json.Marshal(registerResults)
	registerPrinter := printer.Printer{}
	_ = registerPrinter.JSONToTable(payload, projectColumns)
	if tmpDir != "" {
		if _err = os.RemoveAll(tmpDir); _err != nil {
			logger.Errorf(ctx, "unable to delete temp dir %v due to %v", tmpDir, _err)
		}
	}
	return nil
}
