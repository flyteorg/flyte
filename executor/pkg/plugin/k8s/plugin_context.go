package k8s

import (
	"context"
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"

	pluginsCore "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/io"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/ioutils"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/k8s"
	"github.com/flyteorg/flyte/v2/flytestdlib/storage"
)

var _ k8s.PluginContext = &pluginContext{}

// pluginContext adapts a TaskExecutionContext to k8s.PluginContext, adding the K8sReader
// and a buffered OutputWriter that the k8s plugin ecosystem expects.
type pluginContext struct {
	pluginsCore.TaskExecutionContext
	ow             *ioutils.BufferedOutputWriter
	k8sPluginState *k8s.PluginState
	k8sReader      client.Reader
}

func (p *pluginContext) OutputWriter() io.OutputWriter {
	buf := ioutils.NewBufferedOutputWriter(context.TODO(), p.TaskExecutionContext.OutputWriter())
	p.ow = buf
	return buf
}

func (p *pluginContext) PluginStateReader() pluginsCore.PluginStateReader {
	return &pluginStateReader{k8sPluginState: p.k8sPluginState}
}

func (p *pluginContext) K8sReader() client.Reader {
	return p.k8sReader
}

func (p *pluginContext) DataStore() *storage.DataStore {
	return p.TaskExecutionContext.DataStore()
}

func newPluginContext(tCtx pluginsCore.TaskExecutionContext, k8sPluginState *k8s.PluginState, k8sReader client.Reader) *pluginContext {
	return &pluginContext{
		TaskExecutionContext: tCtx,
		k8sPluginState:       k8sPluginState,
		k8sReader:            k8sReader,
	}
}

// pluginStateReader is a specialized PluginStateReader that returns a pre-assigned k8s.PluginState.
// This allows the PluginManager to encapsulate state persistence and only expose Phase/PhaseVersion/Reason
// to k8s plugins.
type pluginStateReader struct {
	k8sPluginState *k8s.PluginState
}

func (p pluginStateReader) GetStateVersion() uint8 {
	return 0
}

func (p pluginStateReader) Get(t interface{}) (uint8, error) {
	if pointer, ok := t.(*k8s.PluginState); ok {
		*pointer = *p.k8sPluginState
	} else {
		return 0, fmt.Errorf("unexpected type when reading plugin state: expected *k8s.PluginState, got %T", t)
	}
	return 0, nil
}
