package printer

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type Inner struct {
	X string     `json:"x"`
	Y *time.Time `json:"y"`
}

func TestJSONToTable(t *testing.T) {
	d := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	j := []struct {
		A string `json:"a"`
		B int    `json:"b"`
		S *Inner `json:"s"`
	}{
		{"hello", 0, &Inner{"x-hello", nil}},
		{"hello", 0, &Inner{"x-hello", &d}},
		{"hello", 0, nil},
	}

	b, err := json.Marshal(j)
	assert.NoError(t, err)
	assert.NoError(t, JSONToTable(b, []Column{
		{"A", "$.a"},
		{"S", "$.s.y"},
	}))
	// Output:
	// | A     | S                    |
	// ------- ----------------------
	// | hello | %!s(<nil>)           |
	// | hello | 2020-01-01T00:00:00Z |
	// | hello |                      |
	// 3 rows
}
