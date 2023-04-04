package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/flyteorg/flytecopilot/data"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/storage"
	"github.com/spf13/cobra"
)

type DownloadOptions struct {
	*RootOptions
	remoteInputsPath    string
	remoteOutputsPrefix string
	localDirectoryPath  string
	inputInterface      []byte
	metadataFormat      string
	downloadMode        string
	timeout             time.Duration
}

func GetFormatVals() []string {
	var vals []string
	for k := range core.DataLoadingConfig_LiteralMapFormat_value {
		vals = append(vals, k)
	}
	return vals
}

func GetDownloadModeVals() []string {
	var vals []string
	for k := range core.IOStrategy_DownloadMode_value {
		vals = append(vals, k)
	}
	return vals
}

func GetUploadModeVals() []string {
	var vals []string
	for k := range core.IOStrategy_UploadMode_value {
		vals = append(vals, k)
	}
	return vals
}

func (d *DownloadOptions) Download(ctx context.Context) error {
	if d.remoteOutputsPrefix == "" {
		return fmt.Errorf("to-output-prefix is required")
	}

	// We need remote outputs prefix to write and error file
	err := func() error {
		if d.localDirectoryPath == "" {
			return fmt.Errorf("to-local-dir is required")
		}
		if d.remoteInputsPath == "" {
			return fmt.Errorf("from-remote is required")
		}
		f, ok := core.DataLoadingConfig_LiteralMapFormat_value[d.metadataFormat]
		if !ok {
			return fmt.Errorf("incorrect input download format specified, given [%s], possible values [%+v]", d.metadataFormat, GetFormatVals())
		}

		m, ok := core.IOStrategy_DownloadMode_value[d.downloadMode]
		if !ok {
			return fmt.Errorf("incorrect input download mode specified, given [%s], possible values [%+v]", d.downloadMode, GetDownloadModeVals())
		}
		dl := data.NewDownloader(ctx, d.Store, core.DataLoadingConfig_LiteralMapFormat(f), core.IOStrategy_DownloadMode(m))
		childCtx := ctx
		cancelFn := func() {}
		if d.timeout > 0 {
			childCtx, cancelFn = context.WithTimeout(ctx, d.timeout)
		}
		defer cancelFn()
		err := dl.DownloadInputs(childCtx, storage.DataReference(d.remoteInputsPath), d.localDirectoryPath)
		if err != nil {
			logger.Errorf(ctx, "Downloading failed, err %s", err)
			return err
		}
		return nil
	}()

	if err != nil {
		if err2 := d.UploadError(ctx, "InputDownloadFailed", err, storage.DataReference(d.remoteOutputsPrefix)); err2 != nil {
			logger.Errorf(ctx, "Failed to write error document, err :%s", err2)
			return err2
		}
	}
	return nil
}

func NewDownloadCommand(opts *RootOptions) *cobra.Command {

	downloadOpts := &DownloadOptions{
		RootOptions: opts,
	}

	// deleteCmd represents the delete command
	downloadCmd := &cobra.Command{
		Use:   "download <opts>",
		Short: "downloads flytedata from the remotepath to a local directory.",
		Long:  `Currently it looks at the outputs.pb and creates one file per variable.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return downloadOpts.Download(context.Background())
		},
	}

	downloadCmd.Flags().StringVarP(&downloadOpts.remoteInputsPath, "from-remote", "f", "", "The remote path/key for inputs in stow store.")
	downloadCmd.Flags().StringVarP(&downloadOpts.remoteOutputsPrefix, "to-output-prefix", "", "", "The remote path/key prefix for outputs in stow store. this is mostly used to write errors.pb.")
	downloadCmd.Flags().StringVarP(&downloadOpts.localDirectoryPath, "to-local-dir", "o", "", "The local directory on disk where data should be downloaded.")
	downloadCmd.Flags().StringVarP(&downloadOpts.metadataFormat, "format", "m", core.DataLoadingConfig_JSON.String(), fmt.Sprintf("What should be the output format for the primitive and structured types. Options [%v]", GetFormatVals()))
	downloadCmd.Flags().StringVarP(&downloadOpts.downloadMode, "download-mode", "d", core.IOStrategy_DOWNLOAD_EAGER.String(), fmt.Sprintf("Download mode to use. Options [%v]", GetDownloadModeVals()))
	downloadCmd.Flags().DurationVarP(&downloadOpts.timeout, "timeout", "t", time.Hour*1, "Max time to allow for downloads to complete, default is 1H")
	downloadCmd.Flags().BytesBase64VarP(&downloadOpts.inputInterface, "input-interface", "i", nil, "Input interface proto message - core.VariableMap, base64 encoced string")
	return downloadCmd
}
