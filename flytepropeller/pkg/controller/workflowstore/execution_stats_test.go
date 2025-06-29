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

func createDefaultExecutionStatsHolder() (*ExecutionStatsHolder, error) {
	esh, err := NewExecutionStatsHolder()
	if err != nil {
		return nil, err
	}
	err = esh.AddOrUpdateEntry("exec1", SingleExecutionStats{ActiveNodeCount: 5, ActiveTaskCount: 10})
	if err != nil {
		return nil, err
	}
	err = esh.AddOrUpdateEntry("exec2", SingleExecutionStats{ActiveNodeCount: 3, ActiveTaskCount: 6})
	if err != nil {
		return nil, err
	}
	return esh, nil
}

func TestAggregateActiveValues(t *testing.T) {
	esh, err := createDefaultExecutionStatsHolder()
	assert.NoError(t, err)

	flows, nodes, tasks, err := esh.AggregateActiveValues()
	assert.NoError(t, err)
	assert.Equal(t, 2, flows)
	assert.Equal(t, uint32(8), nodes)
	assert.Equal(t, uint32(16), tasks)
}

// Test removal on an empty ExecutionStatsHolder
func TestRemoveTerminatedExecutionsEmpty(t *testing.T) {
	esh, err := NewExecutionStatsHolder()
	assert.NoError(t, err)

	err = esh.RemoveTerminatedExecutions(context.TODO(), map[string]bool{})
	assert.NoError(t, err)

	err = esh.RemoveTerminatedExecutions(context.TODO(), map[string]bool{"exec1": true})
	assert.NoError(t, err)
}

// Test removal of a subset of entries from ExcutionStatsHolder
func TestRemoveTerminatedExecutionsSubset(t *testing.T) {
	esh, err := createDefaultExecutionStatsHolder()
	assert.NoError(t, err)

	err = esh.RemoveTerminatedExecutions(context.TODO(), map[string]bool{"exec2": true})
	assert.NoError(t, err)

	assert.Equal(t, 1, len(esh.executions))
	assert.Equal(t, uint32(3), esh.executions["exec2"].ActiveNodeCount)
	assert.Equal(t, uint32(6), esh.executions["exec2"].ActiveTaskCount)
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
			execID := fmt.Sprintf("exec%d", id)
			err := esh.AddOrUpdateEntry(execID, SingleExecutionStats{ActiveNodeCount: uint32(id), ActiveTaskCount: uint32(id * 2)}) // #nosec G115
			assert.NoError(t, err)
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
