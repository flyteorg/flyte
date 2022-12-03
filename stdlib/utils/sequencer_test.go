package utils

import (
	"sync"
	"testing"

	"fmt"

	"github.com/stretchr/testify/assert"
)

func TestSequencer(t *testing.T) {
	size := 3
	sequencer := GetSequencer()
	curVal := sequencer.GetCur() + 1
	// sum = n(a0 + aN) / 2
	expectedSum := uint64(size) * (curVal + curVal + uint64(size-1)) / 2
	numbers := make(chan uint64, size)

	var wg sync.WaitGroup
	wg.Add(size)

	iter := 0
	for iter < size {
		go func() {
			number := sequencer.GetNext()
			fmt.Printf("list value: %d", number)
			numbers <- number
			wg.Done()
		}()
		iter++
	}
	wg.Wait()
	close(numbers)

	unique, sum := uniqueAndSum(numbers)
	assert.True(t, unique, "sequencer generated duplicate numbers")
	assert.Equal(t, expectedSum, sum, "sequencer generated sequence numbers with gap %d %d", expectedSum, sum)
}

func uniqueAndSum(list chan uint64) (bool, uint64) {
	set := make(map[uint64]struct{})
	var sum uint64

	for elem := range list {
		fmt.Printf("list value: %d\n", elem)
		if _, ok := set[elem]; ok {
			return false, sum
		}
		set[elem] = struct{}{}
		sum += elem
	}
	return true, sum
}
