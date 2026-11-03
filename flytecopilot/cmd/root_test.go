package cmd

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/ioutils"
	"github.com/flyteorg/flyte/v2/flytestdlib/promutils"
	"github.com/flyteorg/flyte/v2/flytestdlib/storage"
)

// The error document co-pilot writes on an upload/download failure is only honoured if the
// plugin's output reader finds it. The reader resolves the error path through
// ioutils.ErrorsSuffix, so the flag default has to match it -- otherwise a RECOVERABLE
// failure is written to a file nobody reads and gets mis-classified as a system error.
func TestErrOutputNameDefaultMatchesReader(t *testing.T) {
	cmd := NewDataCommand()
	f := cmd.PersistentFlags().Lookup("err-output-name")
	if assert.NotNil(t, f, "err-output-name flag should be registered") {
		assert.Equal(t, ioutils.ErrorsSuffix, f.DefValue)
	}
}

// End to end across the two components that disagreed: co-pilot writes the error document
// with its default name, and the plugin-side reader must pick it up as a recoverable task
// error carrying the real cause -- not fall through to a missing outputs.pb.
func TestUploadErrorIsReadBackAsRecoverable(t *testing.T) {
	ctx := context.TODO()
	scope := promutils.NewTestScope()
	store, err := storage.NewDataStore(&storage.Config{Type: storage.TypeMemory}, scope.NewSubScope("storage"))
	require.NoError(t, err)

	cmd := NewDataCommand()
	errOutputName := cmd.PersistentFlags().Lookup("err-output-name")
	require.NotNil(t, errOutputName)

	prefix := storage.DataReference("/output")
	opts := RootOptions{Store: store, Scope: scope, errorOutputName: errOutputName.DefValue}
	require.NoError(t, opts.UploadError(ctx, "OutputUploadFailed", fmt.Errorf("connection reset by peer"), prefix))

	reader := ioutils.NewRemoteFileOutputReader(ctx, store, ioutils.NewReadOnlyOutputFilePaths(ctx, store, prefix), 1024*1024)

	isErr, err := reader.IsError(ctx)
	require.NoError(t, err)
	assert.True(t, isErr, "reader should see the error document co-pilot wrote")

	ee, err := reader.ReadError(ctx)
	require.NoError(t, err)
	assert.True(t, ee.IsRecoverable, "an upload failure is retryable against the task's retry budget")
	assert.Equal(t, "OutputUploadFailed", ee.ExecutionError.GetCode())
	assert.Contains(t, ee.ExecutionError.GetMessage(), "connection reset by peer")
}
