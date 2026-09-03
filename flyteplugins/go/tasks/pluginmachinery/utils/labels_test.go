package utils

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/util/validation"
)

func TestSanitizeLabelValue(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected string
	}{
		{"empty stays empty", "", ""},
		{"already valid", "my-run.name_1", "my-run.name_1"},
		{"invalid characters", "my_module.my task/v2", "my_module.my-task-v2"},
		{"leading non alphanumeric", "-leading", "xleading"},
		{"trailing non alphanumeric", "trailing-", "trailingx"},
		{"unicode", "täsk", "t-sk"},
		{"truncated to 63", strings.Repeat("a", 70), strings.Repeat("a", 63)},
		{"truncation cannot leave a trailing dash", strings.Repeat("a", 62) + "-b", strings.Repeat("a", 62) + "x"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := SanitizeLabelValue(test.value)
			assert.Equal(t, test.expected, got)
			if got != "" {
				assert.Empty(t, validation.IsValidLabelValue(got))
			}
		})
	}
}
