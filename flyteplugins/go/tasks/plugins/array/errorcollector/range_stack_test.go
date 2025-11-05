/*
 * Copyright (c) 2018 Lyft. All rights reserved.
 */

package errorcollector

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRangeStack_Add(t *testing.T) {
	s := rangeStack{}
	assert.Nil(t, s.Top())

	s.Push(&indexRange{start: 0, end: 1})
	second := &indexRange{start: 2, end: 5}
	s.Push(second)
	assert.Equal(t, second, s.Top())
}
