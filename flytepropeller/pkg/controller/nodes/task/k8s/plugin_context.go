package k8s

import (
	"context"

	pluginsCore "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/io"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/ioutils"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/k8s"
	"github.com/flyteorg/flytestdlib/logger"
)

var _ k8s.PluginContext = &pluginContext{}

type pluginContext struct {
	pluginsCore.TaskExecutionContext
	// Lazily creates a buffered outputWriter, overriding the input outputWriter.
	ow *ioutils.BufferedOutputWriter
}

// Provides an output sync of type io.OutputWriter
func (p *pluginContext) OutputWriter() io.OutputWriter {
	logger.Debugf(context.TODO(), "K8s plugin is requesting output writer, creating a buffer.")
	buf := ioutils.NewBufferedOutputWriter(context.TODO(), p.TaskExecutionContext.OutputWriter())
	p.ow = buf
	return buf
}

func newPluginContext(tCtx pluginsCore.TaskExecutionContext) *pluginContext {
	return &pluginContext{
		TaskExecutionContext: tCtx,
		ow:                   nil,
	}
}
