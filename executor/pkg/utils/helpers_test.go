package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCopyMap(t *testing.T) {
	m := map[string]string{
		"k1": "v1",
		"k2": "v2",
	}
	co := CopyMap(m)
	assert.NotNil(t, co)
	assert.Equal(t, m, co)

	assert.Nil(t, CopyMap(nil))
}

func TestGetSanitizedPrometheusKey(t *testing.T) {
	output, err := GetSanitizedPrometheusKey("Metric:a_mc")
	assert.Nil(t, err)
	assert.Equal(t, "Metric:a_mc", output)
}

func TestGetSanitizedPrometheusKeyDash(t *testing.T) {
	output, err := GetSanitizedPrometheusKey("Metric--amc")
	assert.Nil(t, err)
	assert.Equal(t, "Metricamc", output)
}

func TestGetSanitizedPrometheusKeySlach(t *testing.T) {
	output, err := GetSanitizedPrometheusKey("Metric:amc\\")
	assert.Nil(t, err)
	assert.Equal(t, "Metric:amc", output)
}
