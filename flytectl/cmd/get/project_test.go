package get

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_transformProject(t *testing.T) {
	t.Run("happy", func(t *testing.T) {
		v, err := json.Marshal(map[string]string{
			"Id":          "id",
			"Name":        "name",
			"Description": "description",
		})
		assert.NoError(t, err)
		row, err := transformProject(v)
		assert.NoError(t, err)
		typedRow, ok := row.(PrintableProject)
		assert.True(t, ok)
		assert.NotNil(t, typedRow)
		assert.Equal(t, "id", typedRow.Id)
	})
}
