/*
 * Copyright (c) 2018 Lyft. All rights reserved.
 */

package awsbatch

import (
	"context"
	"fmt"
	"testing"

	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/stretchr/testify/assert"
)

func TestNewContainerResources_Memory(t *testing.T) {
	testMemory := func(t *testing.T, input string, expected int64) {
		res := newContainerResources(context.TODO(), nil, []*core.Resources_ResourceEntry{
			{
				Name:  core.Resources_MEMORY,
				Value: input,
			},
		})

		assert.Equal(t, expected, res.MemoryMB)
	}

	t.Run("Mb", func(t *testing.T) {
		testMemory(t, "500M", 500)
	})

	t.Run("Gb", func(t *testing.T) {
		testMemory(t, "5G", 5000)
	})

	t.Run("Absolute", func(t *testing.T) {
		testMemory(t, "5000000000", 5000)
	})
}

func TestNewContainerResources_Override(t *testing.T) {
	t.Run("Take first", func(t *testing.T) {
		defaultValues := []*core.Resources_ResourceEntry{
			{
				Name:  core.Resources_CPU,
				Value: "1",
			},
		}

		overrideValues := []*core.Resources_ResourceEntry{
			{
				Name:  core.Resources_CPU,
				Value: "3",
			},
		}

		res := newContainerResources(context.TODO(), nil, overrideValues, defaultValues)
		assert.Equal(t, int64(3), res.Cpus)
	})

	t.Run("Fallback", func(t *testing.T) {
		defaultValues := []*core.Resources_ResourceEntry{
			{
				Name:  core.Resources_CPU,
				Value: "1",
			},
		}

		var overrideValues []*core.Resources_ResourceEntry

		res := newContainerResources(context.TODO(), nil, overrideValues, defaultValues)
		assert.Equal(t, int64(1), res.Cpus)
	})
}

func Example_newContainerResources() {
	defaultValues := []*core.Resources_ResourceEntry{
		{
			Name:  core.Resources_CPU,
			Value: "1",
		},
	}

	overrideValues := []*core.Resources_ResourceEntry{
		{
			Name:  core.Resources_CPU,
			Value: "3",
		},
	}

	res := newContainerResources(context.TODO(), nil, overrideValues, defaultValues)
	fmt.Printf("Computed Cpu: %v\n", res.Cpus)

	// Output:
	// Computed Cpu: 3
}
