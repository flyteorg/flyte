package main

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetNodeIP(t *testing.T) {
	ip, err := getNodeIP()
	require.NoError(t, err)
	assert.NotEmpty(t, ip)

	// Should be a valid IPv4 address
	parsed := net.ParseIP(ip)
	assert.NotNil(t, parsed, "should be a valid IP address")
	assert.NotNil(t, parsed.To4(), "should be an IPv4 address")

	// Should not be loopback
	assert.False(t, parsed.IsLoopback(), "should not be a loopback address")
}

func TestGetFirstNonLoopbackIP(t *testing.T) {
	ip, err := getFirstNonLoopbackIP()
	require.NoError(t, err)
	assert.NotEmpty(t, ip)

	parsed := net.ParseIP(ip)
	assert.NotNil(t, parsed, "should be a valid IP address")
	assert.False(t, parsed.IsLoopback(), "should not be a loopback address")
}
