/*
 * Copyright (c) 2018 Lyft. All rights reserved.
 */

package errorcollector

import (
	"fmt"

	"k8s.io/apimachinery/pkg/util/sets"
)

// Defines a simple string collector that dedupes messages into affected ranges to save in storage.
type ErrorMessageCollector struct {
	messages map[string]*indexRangeCollection
}

func (c ErrorMessageCollector) Collect(idx int, msg string) {
	if existing, found := c.messages[msg]; found {
		existing.Add(idx)
		return
	}

	newRangeCollection := &indexRangeCollection{}
	newRangeCollection.Add(idx)
	c.messages[msg] = newRangeCollection
}

func (c ErrorMessageCollector) Summary(charLimit int) string {
	res := ""
	appendIfNotEnough := "... and many more."
	charLimit -= len(appendIfNotEnough)
	if charLimit < 0 {
		return ""
	}

	sortedKeys := sets.NewString()
	for msg := range c.messages {
		sortedKeys.Insert(msg)
	}

	for _, msg := range sortedKeys.List() {
		res += fmt.Sprintf("%v: %v\n", c.messages[msg].String(), msg)
		if len(res) >= charLimit {
			// TODO: If this is the last element, don't append this message
			res += appendIfNotEnough
			break
		}
	}

	return res
}

func (c ErrorMessageCollector) Length() int {
	return len(c.messages)
}

// Creates a new efficient ErrorMessageCollector that can dedupe errors for array sub-tasks for storage efficiency.
func NewErrorMessageCollector() ErrorMessageCollector {
	return ErrorMessageCollector{
		messages: map[string]*indexRangeCollection{},
	}
}
