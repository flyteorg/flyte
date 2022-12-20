/*
 * Copyright (c) 2018 Lyft. All rights reserved.
 */

package errorcollector

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIndexRangeCollection_simplify(t *testing.T) {
	c := &indexRangeCollection{}
	c.Add(0)
	c.Add(1)
	c.Add(5)
	c.Add(10)
	c.Add(4)
	c.Add(3)
	c.Add(2)
	c.Add(8)

	c.simplify()

	arr := make([]indexRange, 0, len(*c))
	for _, item := range *c {
		arr = append(arr, *item)
	}

	assert.Equal(t, []indexRange{
		{start: 0, end: 5},
		{start: 8, end: 8},
		{start: 10, end: 10},
	}, arr)
}
