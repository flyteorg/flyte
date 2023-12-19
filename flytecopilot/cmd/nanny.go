package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/spf13/cobra"

	"github.com/flyteorg/flyte/flytecopilot/cmd/containerwatcher"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
)

type NannyOptions struct {
	*RootOptions
	downloadMode           string
	downloadTimeout        time.Duration
	inputInterface         []byte
	inputDirectoryPath     string
	metadataFormat         string
	metaOutputName         string
	outputDirectoryPath    string
	remoteInputsPath       string
	remoteOutputsPrefix    string
	remoteOutputsRawPrefix string
	typedInterface         []byte
	uploadMode             string
	uploadTimeout          time.Duration
}

type Nanny struct {
	command string
	args    []string
}

func (n *Nanny) WaitToStart(ctx context.Context) error {
	return nil
}

func (n *Nanny) WaitToExit(ctx context.Context) error {
	cmd := exec.Command(n.command, n.args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (n *NannyOptions) Exec(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("no command provided")
	}

	// Download inputs
	downloadOpts := &DownloadOptions{
		RootOptions:         n.RootOptions,
		remoteInputsPath:    n.remoteInputsPath,
		remoteOutputsPrefix: n.remoteOutputsPrefix,
		localDirectoryPath:  n.inputDirectoryPath,
		inputInterface:      n.inputInterface,
		metadataFormat:      n.metadataFormat,
		downloadMode:        n.downloadMode,
		timeout:             n.downloadTimeout,
	}
	if err := downloadOpts.Download(ctx); err != nil {
		return err
	}

	// Initialize nanny
	nanny := &Nanny{
		command: args[0],
		args:    args[1:],
	}

	// Upload outputs
	uploadOpts := &UploadOptions{
		RootOptions:            n.RootOptions,
		remoteOutputsPrefix:    n.remoteOutputsPrefix,
		metaOutputName:         n.metaOutputName,
		remoteOutputsRawPrefix: n.remoteOutputsRawPrefix,
		localDirectoryPath:     n.outputDirectoryPath,
		metadataFormat:         n.metadataFormat,
		uploadMode:             n.uploadMode,
		timeout:                n.uploadTimeout,
		typedInterface:         n.typedInterface,
		startWatcherFn: func(_ context.Context, _ *UploadOptions) (containerwatcher.Watcher, error) {
			return nanny, nil
		},
	}
	if err := uploadOpts.Sidecar(ctx); err != nil {
		return err
	}

	return nil
}

func NewNannyCommand(opts *RootOptions) *cobra.Command {
	nannyOpts := &NannyOptions{
		RootOptions: opts,
	}

	nannyCmd := &cobra.Command{
		Use:   "nanny <opts>",
		Short: "Run a shell command",
		Long:  `Run a shell command with handling for downloading inputs and uploading outputs`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return nannyOpts.Exec(context.Background(), args)
		},
	}

	nannyCmd.Flags().StringVarP(&nannyOpts.downloadMode, "download-mode", "", core.IOStrategy_DOWNLOAD_EAGER.String(), fmt.Sprintf("Download mode to use. Options [%v]", GetDownloadModeVals()))
	nannyCmd.Flags().DurationVarP(&nannyOpts.downloadTimeout, "download-timeout", "", time.Hour*1, "Max time to allow for downloads to complete, default is 1H")
	nannyCmd.Flags().BytesBase64VarP(&nannyOpts.inputInterface, "input-interface", "", nil, "Input interface proto message - core.VariableMap, base64 encoced string")
	nannyCmd.Flags().StringVarP(&nannyOpts.inputDirectoryPath, "to-local-dir", "", "", "The local directory on disk where data should be downloaded to and uploaded from.")
	nannyCmd.Flags().StringVarP(&nannyOpts.metadataFormat, "format", "", core.DataLoadingConfig_JSON.String(), fmt.Sprintf("What should be the output format for the primitive and structured types. Options [%v]", GetFormatVals()))
	nannyCmd.Flags().StringVarP(&nannyOpts.metaOutputName, "meta-output-name", "", "outputs.pb", "The key name under the remoteOutputPrefix that should be return to provide meta information about the outputs on successful execution")
	nannyCmd.Flags().StringVarP(&nannyOpts.outputDirectoryPath, "from-local-dir", "", "", "The local directory on disk where data should be downloaded to and uploaded from.")
	nannyCmd.Flags().StringVarP(&nannyOpts.remoteInputsPath, "from-remote", "f", "", "The remote path/key for inputs in stow store.")
	nannyCmd.Flags().StringVarP(&nannyOpts.remoteOutputsPrefix, "to-output-prefix", "", "", "The remote path/key prefix for outputs in stow store. this is mostly used to write errors.pb.")
	nannyCmd.Flags().StringVarP(&nannyOpts.remoteOutputsRawPrefix, "to-raw-output", "", "", "The remote path/key prefix for outputs in remote store. This is a sandbox directory and all data will be uploaded here.")
	nannyCmd.Flags().BytesBase64VarP(&nannyOpts.typedInterface, "interface", "", nil, "Typed Interface - core.TypedInterface, base64 encoded string of the serialized protobuf")
	nannyCmd.Flags().StringVarP(&nannyOpts.uploadMode, "upload-mode", "", core.IOStrategy_UPLOAD_ON_EXIT.String(), fmt.Sprintf("When should upload start/upload mode. Options [%v]", GetUploadModeVals()))
	nannyCmd.Flags().DurationVarP(&nannyOpts.uploadTimeout, "upload-timeout", "", time.Hour*1, "Max time to allow for uploads to complete, default is 1H")
	return nannyCmd
}
