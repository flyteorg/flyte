package cmd

import (
	"context"
	"fmt"
	"path"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/flyteorg/flyte/flytecopilot/cmd/containerwatcher"
	"github.com/flyteorg/flyte/flytecopilot/data"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/storage"
)

const (
	StartFile   = "_START"
	SuccessFile = "_SUCCESS"
	ErrorFile   = "_ERROR"
)

type UploadOptions struct {
	*RootOptions
	// The remote prefix where all the meta outputs or error should be uploaded of the form s3://bucket/prefix
	remoteOutputsPrefix string
	// Name like outputs.pb under the remoteOutputsPrefix that should be created to upload the metaOutputs
	metaOutputName string
	// The remote prefix where all the raw outputs should be uploaded of the form s3://bucket/prefix/
	remoteOutputsRawPrefix string
	// Local directory path where the sidecar should look for outputs. Use uploadConfigFilePath to override this for individual outputs.
	localDirectoryPath string
	// Local directory path, used with uploadConfigFilePath.
	uploadConfigDir string
	// Path to a JSON file configuring uploads. It maps an output to the path the sidecar should look for its upload.
	uploadConfigFilePath string
	// Non primitive types will be dumped in this output format
	metadataFormat   string
	uploadMode       string
	timeout          time.Duration
	typedInterface   []byte
	startWatcherType containerwatcher.WatcherType
	exitWatcherType  containerwatcher.WatcherType
}

func (u *UploadOptions) createWatcher(_ context.Context, w containerwatcher.WatcherType) (containerwatcher.Watcher, error) {
	switch w {
	case containerwatcher.WatcherTypeSignal:
		return containerwatcher.SignalWatcher{}, nil
	case containerwatcher.WatcherTypeNoop:
		return containerwatcher.NoopWatcher{}, nil
	}
	return nil, fmt.Errorf("unsupported watcher type")
}

func hydrateUploadConfigs(configs map[string]data.FileIOConfig, vars *core.VariableMap, localDirectoryPath string) map[string]data.FileIOConfig {
	for varName := range vars.GetVariables() {
		if _, ok := configs[varName]; !ok {
			configs[varName] = data.FileIOConfig{
				Path:         path.Join(localDirectoryPath, varName),
				VariableName: varName,
			}
		}
	}
	return configs
}

func (u *UploadOptions) uploader(ctx context.Context) error {
	if u.typedInterface == nil {
		logger.Infof(ctx, "No output interface provided. Assuming Void outputs.")
		return nil
	}

	iface := &core.TypedInterface{}
	if err := proto.Unmarshal(u.typedInterface, iface); err != nil {
		logger.Errorf(ctx, "Bad interface passed, failed to unmarshal err: %s", err)
		return errors.Wrap(err, "Bad interface passed, failed to unmarshal, expected core.TypedInterface")
	}
	outputInterface := iface.GetOutputs()

	if iface.GetOutputs() == nil || iface.Outputs.Variables == nil || len(iface.GetOutputs().GetVariables()) == 0 {
		logger.Infof(ctx, "Empty output interface received. Assuming void outputs. Sidecar will exit immediately.")
		return nil
	}

	f, ok := core.DataLoadingConfig_LiteralMapFormat_value[u.metadataFormat]
	if !ok {
		return fmt.Errorf("incorrect input data format specified, given [%s], possible values [%+v]", u.metadataFormat, GetFormatVals())
	}

	m, ok := core.IOStrategy_UploadMode_value[u.uploadMode]
	if !ok {
		return fmt.Errorf("incorrect input upload mode specified, given [%s], possible values [%+v]", u.uploadMode, GetUploadModeVals())
	}

	logger.Infof(ctx, "Creating start watcher type: %s", u.startWatcherType)
	w, err := u.createWatcher(ctx, u.startWatcherType)
	if err != nil {
		return err
	}

	logger.Infof(ctx, "Waiting for Container to exit.")
	if err := w.WaitToExit(ctx); err != nil {
		logger.Errorf(ctx, "Failed waiting for container to exit. Err: %s", err)
		return err
	}

	logger.Infof(ctx, "Container Exited! uploading data.")

	// TODO maybe we should just take the meta output path as an input argument
	toOutputPath, err := u.Store.ConstructReference(ctx, storage.DataReference(u.remoteOutputsPrefix), u.metaOutputName)
	if err != nil {
		return err
	}

	logger.Infof(ctx, "Loading upload configs from %s", u.uploadConfigFilePath)
	var uploadConfigs map[string]data.FileIOConfig
	if u.uploadConfigFilePath != "" {
		var err error
		uploadConfigs, err = data.LoadFileIOConfigs(u.uploadConfigFilePath, u.uploadConfigDir)
		if err != nil {
			return fmt.Errorf("failed to load upload configs: %w", err)
		}
	} else {
		uploadConfigs = make(map[string]data.FileIOConfig)
	}
	hydrateUploadConfigs(uploadConfigs, outputInterface, u.localDirectoryPath)

	errorFilePath := path.Join(u.localDirectoryPath, ErrorFile)
	dl := data.NewUploader(ctx, u.Store, core.DataLoadingConfig_LiteralMapFormat(f), core.IOStrategy_UploadMode(m), ErrorFile)

	childCtx, cancelFn := context.WithTimeout(ctx, u.timeout)
	defer cancelFn()
	if err := dl.RecursiveUpload(childCtx, outputInterface, uploadConfigs, errorFilePath, toOutputPath, storage.DataReference(u.remoteOutputsRawPrefix)); err != nil {
		logger.Errorf(ctx, "Uploading failed, err %s", err)
		return err
	}

	logger.Infof(ctx, "Uploader completed successfully!")
	return nil
}

func (u *UploadOptions) Sidecar(ctx context.Context) error {

	if err := u.uploader(ctx); err != nil {
		logger.Errorf(ctx, "Uploading failed, err %s", err)
		if err := u.UploadError(ctx, "OutputUploadFailed", err, storage.DataReference(u.remoteOutputsPrefix)); err != nil {
			logger.Errorf(ctx, "Failed to write error document, err :%s", err)
			return err
		}
	}
	return nil
}

func NewUploadCommand(opts *RootOptions) *cobra.Command {

	uploadOptions := &UploadOptions{
		RootOptions: opts,
	}

	// deleteCmd represents the delete command
	uploadCmd := &cobra.Command{
		Use:   "sidecar <opts>",
		Short: "uploads flyteData from the localpath to a remote dir.",
		Long:  `Currently it looks at the outputs.pb and creates one file per variable.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return uploadOptions.Sidecar(context.Background())
		},
	}

	uploadCmd.Flags().StringVarP(&uploadOptions.remoteOutputsPrefix, "to-output-prefix", "o", "", "The remote path/key prefix for output metadata in stow store.")
	uploadCmd.Flags().StringVarP(&uploadOptions.remoteOutputsRawPrefix, "to-raw-output", "x", "", "The remote path/key prefix for outputs in remote store. This is a sandbox directory and all data will be uploaded here.")
	uploadCmd.Flags().StringVarP(&uploadOptions.localDirectoryPath, "from-local-dir", "f", "", "The local directory on disk where data will be available for upload. Use the upload-config-* flags to override this for individual outputs.")
	uploadCmd.Flags().StringVarP(&uploadOptions.uploadConfigDir, "upload-config-dir", "", "", "The local directory on disk where the sidecar should look for upload configs.")
	uploadCmd.Flags().StringVarP(&uploadOptions.uploadConfigFilePath, "upload-config-file-path", "", "", "Path to a JSON file configuring uploads. It maps output variable names to their local file paths for uploads.")
	uploadCmd.Flags().StringVarP(&uploadOptions.metadataFormat, "format", "m", core.DataLoadingConfig_JSON.String(), fmt.Sprintf("What should be the output format for the primitive and structured types. Options [%v]", GetFormatVals()))
	uploadCmd.Flags().StringVarP(&uploadOptions.uploadMode, "upload-mode", "u", core.IOStrategy_UPLOAD_ON_EXIT.String(), fmt.Sprintf("When should upload start/upload mode. Options [%v]", GetUploadModeVals()))
	uploadCmd.Flags().StringVarP(&uploadOptions.metaOutputName, "meta-output-name", "", "outputs.pb", "The key name under the remoteOutputPrefix that should be return to provide meta information about the outputs on successful execution")
	uploadCmd.Flags().DurationVarP(&uploadOptions.timeout, "timeout", "t", time.Hour*1, "Max time to allow for uploads to complete, default is 1H")
	uploadCmd.Flags().BytesBase64VarP(&uploadOptions.typedInterface, "interface", "i", nil, "Typed Interface - core.TypedInterface, base64 encoded string of the serialized protobuf")
	uploadCmd.Flags().DurationVarP(&uploadOptions.timeout, "start-timeout", "", 0, "Deprecated: Use --timeout instead. Specifies the maximum duration to allow uploads to complete. Retained for backward compatibility.")
	uploadCmd.Flags().StringVarP(&uploadOptions.startWatcherType, "start-watcher-type", "", containerwatcher.WatcherTypeSignal, fmt.Sprintf("Sidecar will wait for container before starting upload process. Watcher type makes the type configurable. Available Type %+v", containerwatcher.AllWatcherTypes))
	uploadCmd.Flags().StringVarP(&uploadOptions.exitWatcherType, "exit-watcher-type", "", containerwatcher.WatcherTypeSignal, fmt.Sprintf("Sidecar will wait for completion of the container before starting upload process. Watcher type makes the type configurable. Available Type %+v", containerwatcher.AllWatcherTypes))
	return uploadCmd
}
