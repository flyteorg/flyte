/*
 * Copyright (c) 2018 Lyft. All rights reserved.
 */

package errorcollector

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorMessageCollector_Collect(t *testing.T) {
	c := NewErrorMessageCollector()
	dupeMsg := "duplicate message"
	c.Collect(0, dupeMsg)
	c.Collect(1, dupeMsg)
	c.Collect(2, "unique message")
	c.Collect(3, dupeMsg)

	assert.Len(t, c.messages, 2)
	assert.Len(t, *c.messages[dupeMsg], 2)
}

func TestErrorMessageCollector_Summary(t *testing.T) {
	c := NewErrorMessageCollector()
	dupeMsg := "duplicate message"
	c.Collect(0, dupeMsg)
	c.Collect(1, dupeMsg)
	c.Collect(2, "unique message")
	c.Collect(3, dupeMsg)

	assert.Equal(t, "[0-1][3]: duplicate message\n[2]: unique message\n", c.Summary(1000))
	assert.Equal(t, "[0-1][3]: duplicate message\n... and many more.", c.Summary(30))
}
