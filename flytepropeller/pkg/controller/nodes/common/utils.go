package common

import (
	pluginCore "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core"
	"strconv"
	"sync"

	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/encoding"
	"github.com/flyteorg/flyte/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/executors"
)

const maxUniqueIDLength = 20

// GenerateUniqueID is the UniqueId of a node is unique within a given workflow execution.
// In order to achieve that we track the lineage of the node.
// To compute the uniqueID of a node, we use the uniqueID and retry attempt of the parent node
// For nodes in level 0, there is no parent, and parentInfo is nil
func GenerateUniqueID(parentInfo executors.ImmutableParentInfo, nodeID string) (string, error) {
	var parentUniqueID v1alpha1.NodeID
	var parentRetryAttempt string

	if parentInfo != nil {
		parentUniqueID = parentInfo.GetUniqueID()
		parentRetryAttempt = strconv.Itoa(int(parentInfo.CurrentAttempt()))
	}

	return encoding.FixedLengthUniqueIDForParts(maxUniqueIDLength, []string{parentUniqueID, parentRetryAttempt, nodeID})
}

// CreateParentInfo creates a unique parent id, the unique id of parent is dependent on the unique id and the current
// attempt of the grandparent to track the lineage.
func CreateParentInfo(grandParentInfo executors.ImmutableParentInfo, nodeID string, parentAttempt uint32) (executors.ImmutableParentInfo, error) {
	uniqueID, err := GenerateUniqueID(grandParentInfo, nodeID)
	if err != nil {
		return nil, err
	}
	return executors.NewParentInfo(uniqueID, parentAttempt), nil

}

type SafeDefaultPlugins struct {
	mu      sync.RWMutex
	plugins map[pluginCore.TaskType]pluginCore.Plugin
}

func NewSafeDefaultPlugins() SafeDefaultPlugins {
	return SafeDefaultPlugins{
		plugins: make(map[pluginCore.TaskType]pluginCore.Plugin),
	}
}

func (p *SafeDefaultPlugins) GetPlugin(taskType pluginCore.TaskType) (pluginCore.Plugin, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	corePlugin, exists := p.plugins[taskType]
	return corePlugin, exists
}

func (p *SafeDefaultPlugins) GetAllPlugins() map[pluginCore.TaskType]pluginCore.Plugin {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.plugins
}

func (p *SafeDefaultPlugins) SetPlugin(taskType pluginCore.TaskType, plugin pluginCore.Plugin) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.plugins[taskType] = plugin
}
