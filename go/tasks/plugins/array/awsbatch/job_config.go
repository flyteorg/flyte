/*
 * Copyright (c) 2018 Lyft. All rights reserved.
 */

package awsbatch

import (
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	v1 "k8s.io/api/core/v1"
)

const (
	// Keep these in-sync with flyteAdmin @
	// https://github.com/flyteorg/flyteadmin/blob/d1c61c34f62d8ee51964f47877802d070dfa9e98/pkg/manager/impl/execution_manager.go#L42-L43

	PrimaryTaskQueueKey = "primary_queue"
	DynamicTaskQueueKey = "dynamic_queue"
	ChildTaskQueueKey   = "child_queue"
)

type JobConfig struct {
	PrimaryTaskQueue string `json:"primary_queue"`
	DynamicTaskQueue string `json:"dynamic_queue"`
}

func (j *JobConfig) setKeyIfKnown(key, value string) bool {
	switch key {
	case PrimaryTaskQueueKey:
		j.PrimaryTaskQueue = value
		return true
	case ChildTaskQueueKey:
		fallthrough
	case DynamicTaskQueueKey:
		j.DynamicTaskQueue = value
		return true
	default:
		return false
	}
}

func (j *JobConfig) MergeFromKeyValuePairs(pairs []*core.KeyValuePair) *JobConfig {
	for _, entry := range pairs {
		j.setKeyIfKnown(entry.Key, entry.Value)
	}

	return j
}

func (j *JobConfig) MergeFromConfigMap(configMap *v1.ConfigMap) *JobConfig {
	if configMap == nil {
		return j
	}

	for key, value := range configMap.Data {
		j.setKeyIfKnown(key, value)
	}

	return j
}

func newJobConfig() (cfg *JobConfig) {
	return &JobConfig{}
}
