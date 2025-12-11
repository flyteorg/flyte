package handler

import (
	"time"

	pluginCore "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyte/v2/flytestdlib/storage"
)

type TaskNodeState struct {
	PluginPhase                        pluginCore.Phase
	PluginPhaseVersion                 uint32
	PluginState                        []byte
	PluginStateVersion                 uint32
	LastPhaseUpdatedAt                 time.Time
	PreviousNodeExecutionCheckpointURI storage.DataReference
	CleanupOnFailure                   bool
}
