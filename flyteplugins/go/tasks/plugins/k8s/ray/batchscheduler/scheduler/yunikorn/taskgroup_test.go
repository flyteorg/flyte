package yunikorn

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMarshal(t *testing.T) {
	t1 := TaskGroup{
		Name:                      "tg1",
		MinMember:                 int32(1),
		Labels:                    map[string]string{"attr": "value"},
		Annotations:               map[string]string{"attr": "value"},
		MinResource:               res,
		NodeSelector:              map[string]string{"node": "gpunode"},
		Tolerations:               nil,
		Affinity:                  nil,
		TopologySpreadConstraints: nil,
	}
	t2 := TaskGroup{
		Name:                      "tg2",
		MinMember:                 int32(1),
		Labels:                    map[string]string{"attr": "value"},
		Annotations:               map[string]string{"attr": "value"},
		MinResource:               res,
		NodeSelector:              map[string]string{"node": "gpunode"},
		Tolerations:               nil,
		Affinity:                  nil,
		TopologySpreadConstraints: nil,
	}
	var tests = []struct {
		input []TaskGroup
	}{
		{input: nil},
		{input: []TaskGroup{}},
		{input: []TaskGroup{t1}},
		{input: []TaskGroup{t1, t2}},
	}
	t.Run("Serialize task groups", func(t *testing.T) {
		for _, tt := range tests {
			_, err := Marshal(tt.input)
			assert.Nil(t, err)
		}
	})
}
