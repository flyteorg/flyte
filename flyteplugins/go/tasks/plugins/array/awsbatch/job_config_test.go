/*
 * Copyright (c) 2018 Lyft. All rights reserved.
 */

package awsbatch

import (
	"testing"

	"github.com/magiconair/properties/assert"
)

type jobConfigTestCase struct {
	name     string
	key      string
	value    string
	set      bool
	expected JobConfig
}

func TestSetKeyIfKnown(t *testing.T) {
	testCases := []jobConfigTestCase{
		{
			name:     "no match",
			key:      "foo",
			value:    "bar",
			set:      false,
			expected: JobConfig{},
		},
		{
			name:  "primary queue key",
			key:   "primary_queue",
			value: "foo",
			set:   true,
			expected: JobConfig{
				PrimaryTaskQueue: "foo",
			},
		},
		{
			name:  "dynamic queue key",
			key:   "dynamic_queue",
			value: "bar",
			set:   true,
			expected: JobConfig{
				DynamicTaskQueue: "bar",
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			jobConfig := JobConfig{}
			set := jobConfig.setKeyIfKnown(testCase.key, testCase.value)
			assert.Equal(t, testCase.set, set)
			assert.Equal(t, testCase.expected, jobConfig)
		})
	}
}
