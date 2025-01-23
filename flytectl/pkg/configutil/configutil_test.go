package configutil

import (
	"io"
	"io/ioutil"
	"os"
	"testing"

	f "github.com/flyteorg/flyte/flytectl/pkg/filesystemutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetupConfig(t *testing.T) {
	file, err := os.CreateTemp("", "*.yaml")
	require.NoError(t, err)

	templateValue := ConfigTemplateSpec{
		Host:     "dns:///localhost:30081",
		Insecure: true,
	}
	err = SetupConfig(file.Name(), AdminConfigTemplate, templateValue)
	assert.NoError(t, err)
	configBytes, err := io.ReadAll(file)
	assert.NoError(t, err)
	expected := `admin:
  # For GRPC endpoints you might want to use dns:///flyte.myexample.com
  endpoint: dns:///localhost:30081
  insecure: true
`
	assert.Equal(t, expected, string(configBytes))

	file, err = os.Create(file.Name())
	require.NoError(t, err)
	templateValue = ConfigTemplateSpec{
		Host:     "dns:///admin.example.com",
		Insecure: true,
		Console:  "https://console.example.com",
	}
	err = SetupConfig(file.Name(), AdminConfigTemplate, templateValue)
	assert.NoError(t, err)
	configBytes, err = ioutil.ReadAll(file)
	assert.NoError(t, err)
	expected = `admin:
  # For GRPC endpoints you might want to use dns:///flyte.myexample.com
  endpoint: dns:///admin.example.com
  insecure: true
console:
  endpoint: https://console.example.com
`
	assert.Equal(t, expected, string(configBytes))

	file, err = os.Create(file.Name())
	require.NoError(t, err)
	templateValue = ConfigTemplateSpec{
		Host:     "dns:///admin.example.com",
		Insecure: true,
		DataConfig: &DataConfig{
			Endpoint:  "http://localhost:9000",
			AccessKey: "my-access-key",
			SecretKey: "my-secret-key",
		},
	}
	err = SetupConfig(file.Name(), AdminConfigTemplate, templateValue)
	assert.NoError(t, err)
	configBytes, err = io.ReadAll(file)
	assert.NoError(t, err)
	expected = `admin:
  # For GRPC endpoints you might want to use dns:///flyte.myexample.com
  endpoint: dns:///admin.example.com
  insecure: true
`
	assert.Equal(t, expected, string(configBytes))

	// Cleanup
	if file != nil {
		_ = os.Remove(file.Name())
	}
}

func TestConfigCleanup(t *testing.T) {
	_, err := os.Stat(f.FilePathJoin(f.UserHomeDir(), ".flyte"))
	if os.IsNotExist(err) {
		_ = os.MkdirAll(f.FilePathJoin(f.UserHomeDir(), ".flyte"), 0755)
	}
	_ = os.WriteFile(FlytectlConfig, []byte("string"), 0600)
	_ = os.WriteFile(Kubeconfig, []byte("string"), 0600)

	err = ConfigCleanup()
	assert.Nil(t, err)

	_, err = os.Stat(FlytectlConfig)
	check := os.IsNotExist(err)
	assert.Equal(t, check, true)

	_, err = os.Stat(Kubeconfig)
	check = os.IsNotExist(err)
	assert.Equal(t, check, true)
	_ = ConfigCleanup()
}

func TestSetupFlytectlConfig(t *testing.T) {
	templateValue := ConfigTemplateSpec{
		Host:     "dns:///localhost:30081",
		Insecure: true,
	}
	_, err := os.Stat(f.FilePathJoin(f.UserHomeDir(), ".flyte"))
	if os.IsNotExist(err) {
		_ = os.MkdirAll(f.FilePathJoin(f.UserHomeDir(), ".flyte"), 0755)
	}
	err = SetupConfig("version.yaml", AdminConfigTemplate, templateValue)
	assert.Nil(t, err)
	_, err = os.Stat("version.yaml")
	assert.Nil(t, err)
	check := os.IsNotExist(err)
	assert.Equal(t, check, false)
	_ = ConfigCleanup()
}

func TestAwsConfig(t *testing.T) {
	assert.Equal(t, AdminConfigTemplate, GetTemplate())
}
