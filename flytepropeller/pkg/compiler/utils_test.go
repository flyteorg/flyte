package compiler

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/util/sets"
)

func neighbors(adjList map[string][]string) func(nodeId string) sets.String {
	return func(nodeId string) sets.String {
		if lst, found := adjList[nodeId]; found {
			return sets.NewString(lst...)
		}

		return sets.NewString()
	}
}

func uniqueNodesCount(adjList map[string][]string) int {
	uniqueNodeIds := sets.NewString()
	for key, value := range adjList {
		uniqueNodeIds.Insert(key)
		uniqueNodeIds.Insert(value...)
	}

	return uniqueNodeIds.Len()
}

func assertNoCycle(t *testing.T, startNode string, adjList map[string][]string) {
	cycle, visited, detected := detectCycle(startNode, neighbors(adjList))
	assert.False(t, detected)
	assert.Equal(t, uniqueNodesCount(adjList), len(visited))
	assert.Equal(t, 0, len(cycle))
}

func assertCycle(t *testing.T, startNode string, adjList map[string][]string) {
	cycle, _, detected := detectCycle(startNode, neighbors(adjList))
	assert.True(t, detected)
	assert.NotEqual(t, 0, len(cycle))
	t.Logf("Cycle: %v", strings.Join(cycle, ","))
}

func TestDetectCycle(t *testing.T) {
	t.Run("Linear", func(t *testing.T) {
		linear := map[string][]string{
			"1": {"2"},
			"2": {"3"},
			"3": {"4"},
		}

		assertNoCycle(t, "1", linear)
	})

	t.Run("Cycle", func(t *testing.T) {
		cyclic := map[string][]string{
			"1": {"2", "3"},
			"2": {"3"},
			"3": {"1"},
		}

		assertCycle(t, "1", cyclic)
	})
}
