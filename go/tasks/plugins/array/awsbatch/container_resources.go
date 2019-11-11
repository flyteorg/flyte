/*
 * Copyright (c) 2018 Lyft. All rights reserved.
 */

package awsbatch

import (
	"context"

	"strings"

	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/lyft/flytestdlib/logger"
	"k8s.io/apimachinery/pkg/api/resource"
)

const (
	CpusKey     = "cpu"
	GpusKey     = "gpu"
	MemoryMbKey = "memory"
)

const (
	NoScale resource.Scale = 0
)

const (
	DefaultCPU    = 2
	DefaultMemory = 700
)

type resourceOverrides struct {
	Cpus     int64 `json:"cpu"`
	Gpus     int64 `json:"gpu"`
	MemoryMB int64 `json:"memory"`
}

func resourceEntriesIndex(ctx context.Context, resources []*core.Resources_ResourceEntry) map[string]*core.Resources_ResourceEntry {
	res := make(map[string]*core.Resources_ResourceEntry, len(resources))
	for _, resourceEntry := range resources {
		if resourceName, found := core.Resources_ResourceName_name[int32(resourceEntry.Name)]; found {
			res[strings.ToLower(resourceName)] = resourceEntry
		} else {
			logger.Warnf(ctx, "Resource Name value's not recognized [%v]", resourceEntry.Name)
		}
	}

	return res
}

func getInt64ResourceValue(_ context.Context, key string, resources []map[string]*core.Resources_ResourceEntry,
	scale resource.Scale, defaultValue int64) int64 {

	for _, resourcesMap := range resources {
		if val, found := resourcesMap[key]; found {
			if q, err := resource.ParseQuantity(val.Value); err == nil {
				return q.ScaledValue(scale)
			}

			break
		}
	}

	return defaultValue
}

// Creates a ContainerResources after parsing out known keys from resources. The first map that contains the given key
// wins. So to get the overriding behavior, you should pass the highest precedence override first.
func newContainerResources(ctx context.Context, resources ...[]*core.Resources_ResourceEntry) resourceOverrides {
	resourcesIndics := make([]map[string]*core.Resources_ResourceEntry, 0, len(resources))
	for _, resourceList := range resources {
		resourcesIndics = append(resourcesIndics, resourceEntriesIndex(ctx, resourceList))
	}

	overrides := resourceOverrides{
		Cpus:     getInt64ResourceValue(ctx, CpusKey, resourcesIndics, NoScale, DefaultCPU),
		Gpus:     getInt64ResourceValue(ctx, GpusKey, resourcesIndics, NoScale, 0),
		MemoryMB: getInt64ResourceValue(ctx, MemoryMbKey, resourcesIndics, resource.Mega, DefaultMemory),
	}

	return overrides
}

func newContainerResourcesFromContainerTask(ctx context.Context, container *core.Container) resourceOverrides {
	return newContainerResources(ctx, container.GetResources().GetRequests())
}
