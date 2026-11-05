package sandbox

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetK8sEndpoint(t *testing.T) {
	cfg := Config{
		Port: "6443",
	}
	assert.Equal(t, "https://127.0.0.1:6443", cfg.GetK8sEndpoint())

	customCfg := Config{
		Port: "30086",
	}
	assert.Equal(t, "https://127.0.0.1:30086", customCfg.GetK8sEndpoint())
}
