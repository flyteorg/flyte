package yunikorn

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateTaskGroupName(t *testing.T) {
	type inputFormat struct {
		isMaster bool
		index    int
	}
	var tests = []struct {
		input  inputFormat
		expect string
	}{
		{
			input:  inputFormat{isMaster: true, index: 0},
			expect: fmt.Sprintf("%s-%s", TaskGroupGenericName, "head"),
		},
		{
			input:  inputFormat{isMaster: true, index: 1},
			expect: fmt.Sprintf("%s-%s", TaskGroupGenericName, "head"),
		},
		{
			input:  inputFormat{isMaster: false, index: 0},
			expect: fmt.Sprintf("%s-%s-%d", TaskGroupGenericName, "worker", 0),
		},
		{
			input:  inputFormat{isMaster: false, index: 1},
			expect: fmt.Sprintf("%s-%s-%d", TaskGroupGenericName, "worker", 1),
		},
	}
	t.Run("Gernerate ray task group name", func(t *testing.T) {
		for _, tt := range tests {
			got := GenerateTaskGroupName(tt.input.isMaster, tt.input.index)
			assert.Equal(t, tt.expect, got)
		}
	})
}
