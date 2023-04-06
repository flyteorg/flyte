package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/flyteorg/flytecopilot/cmd/containerwatcher"
	"github.com/flyteorg/flytecopilot/data"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/storage"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
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
	// Local directory path where the sidecar should look for outputs.
	localDirectoryPath string
	// Non primitive types will be dumped in this output format
	metadataFormat        string
	uploadMode            string
	timeout               time.Duration
	containerStartTimeout time.Duration
	typedInterface        []byte
	startWatcherType      containerwatcher.WatcherType
	exitWatcherType       containerwatcher.WatcherType
	containerInfo         containerwatcher.ContainerInformation
}

func (u *UploadOptions) createWatcher(ctx context.Context, w containerwatcher.WatcherType) (containerwatcher.Watcher, error) {
	switch w {
	case containerwatcher.WatcherTypeKubeAPI:
		// TODO, in this case container info should have namespace and podname and we can get it using downwardapi
		// TODO https://kubernetes.io/docs/tasks/inject-data-application/environment-variable-expose-pod-information/
		return containerwatcher.NewKubeAPIWatcher(ctx, u.RootOptions.clientConfig, u.containerInfo)
	case containerwatcher.WatcherTypeFile:
		return containerwatcher.NewSuccessFileWatcher(ctx, u.localDirectoryPath, StartFile, SuccessFile, ErrorFile)
	case containerwatcher.WatcherTypeSharedProcessNS:
		return containerwatcher.NewSharedProcessNSWatcher(ctx, time.Second*2, 2)
	case containerwatcher.WatcherTypeNoop:
		return containerwatcher.NoopWatcher{}, nil
	}
	return nil, fmt.Errorf("unsupported watcher type")
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
	outputInterface := iface.Outputs

	if iface.Outputs == nil || iface.Outputs.Variables == nil || len(iface.Outputs.Variables) == 0 {
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

	logger.Infof(ctx, "Waiting for Container to start with timeout %s.", u.containerStartTimeout)
	childCtx, cancelFn := context.WithTimeout(ctx, u.containerStartTimeout)
	defer cancelFn()
	err = w.WaitToStart(childCtx)
	if err != nil && err != containerwatcher.ErrTimeout {
		return err
	}

	if err != nil {
		logger.Warnf(ctx, "Container start detection aborted, :%s", err.Error())
	}

	if u.startWatcherType != u.exitWatcherType {
		logger.Infof(ctx, "Creating watcher type: %s", u.exitWatcherType)
		w, err = u.createWatcher(ctx, u.exitWatcherType)
		if err != nil {
			return err
		}
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

	dl := data.NewUploader(ctx, u.Store, core.DataLoadingConfig_LiteralMapFormat(f), core.IOStrategy_UploadMode(m), ErrorFile)

	childCtx, cancelFn = context.WithTimeout(ctx, u.timeout)
	defer cancelFn()
	if err := dl.RecursiveUpload(childCtx, outputInterface, u.localDirectoryPath, toOutputPath, storage.DataReference(u.remoteOutputsRawPrefix)); err != nil {
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
	uploadCmd.Flags().StringVarP(&uploadOptions.localDirectoryPath, "from-local-dir", "f", "", "The local directory on disk where data will be available for upload.")
	uploadCmd.Flags().StringVarP(&uploadOptions.metadataFormat, "format", "m", core.DataLoadingConfig_JSON.String(), fmt.Sprintf("What should be the output format for the primitive and structured types. Options [%v]", GetFormatVals()))
	uploadCmd.Flags().StringVarP(&uploadOptions.uploadMode, "upload-mode", "u", core.IOStrategy_UPLOAD_ON_EXIT.String(), fmt.Sprintf("When should upload start/upload mode. Options [%v]", GetUploadModeVals()))
	uploadCmd.Flags().StringVarP(&uploadOptions.metaOutputName, "meta-output-name", "", "outputs.pb", "The key name under the remoteOutputPrefix that should be return to provide meta information about the outputs on successful execution")
	uploadCmd.Flags().DurationVarP(&uploadOptions.timeout, "timeout", "t", time.Hour*1, "Max time to allow for uploads to complete, default is 1H")
	uploadCmd.Flags().BytesBase64VarP(&uploadOptions.typedInterface, "interface", "i", nil, "Typed Interface - core.TypedInterface, base64 encoded string of the serialized protobuf")
	uploadCmd.Flags().DurationVarP(&uploadOptions.containerStartTimeout, "start-timeout", "", 0, "Max time to allow for container to startup. 0 indicates wait for ever.")
	uploadCmd.Flags().StringVarP(&uploadOptions.startWatcherType, "start-watcher-type", "", containerwatcher.WatcherTypeSharedProcessNS, fmt.Sprintf("Sidecar will wait for container before starting upload process. Watcher type makes the type configurable. Available Type %+v", containerwatcher.AllWatcherTypes))
	uploadCmd.Flags().StringVarP(&uploadOptions.exitWatcherType, "exit-watcher-type", "", containerwatcher.WatcherTypeSharedProcessNS, fmt.Sprintf("Sidecar will wait for completion of the container before starting upload process. Watcher type makes the type configurable. Available Type %+v", containerwatcher.AllWatcherTypes))
	uploadCmd.Flags().StringVarP(&uploadOptions.containerInfo.Name, "watch-container", "", "", "For KubeAPI watcher, Wait for this container to exit.")
	uploadCmd.Flags().StringVarP(&uploadOptions.containerInfo.Namespace, "namespace", "", "", "For KubeAPI watcher, Namespace of the pod [optional]")
	uploadCmd.Flags().StringVarP(&uploadOptions.containerInfo.PodName, "pod-name", "", "", "For KubeAPI watcher, Name of the pod [optional].")
	return uploadCmd
}
