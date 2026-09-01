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

	t.Run("Reconverging", func(t *testing.T) {
		reconverging := map[string][]string{
			"root":     {"left", "right"},
			"left":     {"shared-a", "shared-b"},
			"right":    {"shared-a", "shared-b"},
			"shared-a": {"leaf"},
			"shared-b": {"leaf"},
		}
		visits := make(map[string]int)

		cycle, visited, detected := detectCycle("root", func(nodeID string) sets.String {
			visits[nodeID]++
			return neighbors(reconverging)(nodeID)
		})

		assert.False(t, detected)
		assert.Empty(t, cycle)
		assert.Equal(t, uniqueNodesCount(reconverging), len(visited))
		assert.Equal(t, map[string]int{
			"root":     1,
			"left":     1,
			"right":    1,
			"shared-a": 1,
			"shared-b": 1,
			"leaf":     1,
		}, visits)
	})
}
