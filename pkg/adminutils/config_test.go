package adminutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetConfig(t *testing.T) {
	defaultConfig = &Config{
		MaxRecords: 500,
		BatchSize:  100,
	}
	c := GetConfig()
	assert.Equal(t, defaultConfig.BatchSize, c.BatchSize)
	assert.Equal(t, defaultConfig.MaxRecords, c.MaxRecords)
}
