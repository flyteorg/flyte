package vars

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVars(t *testing.T) {
	v := NewVars(map[string]ValueGetter{
		"%(foo)%": func() (string, error) {
			return "there", nil
		},
		"%(bar)%": func() (string, error) {
			return "friend", nil
		},
	})
	result, err := v.Transform([]byte("hello %(foo)%, %(bar)%"))
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, result, []byte("hello there, friend"), "strings should match")
}
