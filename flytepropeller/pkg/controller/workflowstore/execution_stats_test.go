package workflowstore

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddOrUpdateEntry(t *testing.T) {
	esh, err := NewExecutionStatsHolder()
	assert.NoError(t, err)

	err = esh.AddOrUpdateEntry("exec1", SingleExecutionStats{ActiveNodeCount: 5, ActiveTaskCount: 10})
	assert.NoError(t, err)
	assert.Equal(t, 1, len(esh.executions))
	assert.Equal(t, uint32(5), esh.executions["exec1"].ActiveNodeCount)
	assert.Equal(t, uint32(10), esh.executions["exec1"].ActiveTaskCount)
}

func TestAggregateActiveValues(t *testing.T) {
	esh, _ := NewExecutionStatsHolder()
	esh.AddOrUpdateEntry("exec1", SingleExecutionStats{ActiveNodeCount: 5, ActiveTaskCount: 10})
	esh.AddOrUpdateEntry("exec2", SingleExecutionStats{ActiveNodeCount: 3, ActiveTaskCount: 6})

	flows, nodes, tasks, err := esh.AggregateActiveValues()
	assert.NoError(t, err)
	assert.Equal(t, 2, flows)
	assert.Equal(t, uint32(8), nodes)
	assert.Equal(t, uint32(16), tasks)
}

func TestConcurrentAccess(t *testing.T) {
	esh, err := NewExecutionStatsHolder()
	assert.NoError(t, err)

	var wg sync.WaitGroup
	// Number of concurrent operations
	concurrentOps := 100

	// Concurrently add or update entries
	for i := 0; i < concurrentOps; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			execId := fmt.Sprintf("exec%d", id)
			esh.AddOrUpdateEntry(execId, SingleExecutionStats{ActiveNodeCount: uint32(id), ActiveTaskCount: uint32(id * 2)})
		}(i)
	}

	// Concurrently sum active values
	for i := 0; i < concurrentOps/2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _, _, err := esh.AggregateActiveValues()
			assert.NoError(t, err)
		}()
	}

	wg.Wait()

	// Remove all entries
	err = esh.RemoveTerminatedExecutions(context.TODO(), map[string]bool{})
	assert.NoError(t, err)

	// After all operations, sum should be predictable as all entries should be deleted
	flows, nodes, tasks, err := esh.AggregateActiveValues()
	assert.NoError(t, err)
	assert.Zero(t, flows)
	assert.Zero(t, nodes)
	assert.Zero(t, tasks)
}
