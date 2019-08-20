package resourcemanager

import (
	"github.com/magiconair/properties/assert"
	"testing"
)

func TestCreateKey(t *testing.T) {
	assert.Equal(t, "asdf:fdsa", createKey("asdf", "fdsa"))
}
