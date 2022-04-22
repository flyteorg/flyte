package configutil

import (
	"io/ioutil"
	"os"
	"testing"

	f "github.com/flyteorg/flytectl/pkg/filesystemutils"
	"github.com/stretchr/testify/assert"
)

func TestConfigCleanup(t *testing.T) {
	_, err := os.Stat(f.FilePathJoin(f.UserHomeDir(), ".flyte"))
	if os.IsNotExist(err) {
		_ = os.MkdirAll(f.FilePathJoin(f.UserHomeDir(), ".flyte"), 0755)
	}
	_ = ioutil.WriteFile(FlytectlConfig, []byte("string"), 0600)
	_ = ioutil.WriteFile(Kubeconfig, []byte("string"), 0600)

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
	assert.Equal(t, AdminConfigTemplate+StorageS3ConfigTemplate, GetAWSCloudTemplate())
	assert.Equal(t, AdminConfigTemplate+StorageGCSConfigTemplate, GetGoogleCloudTemplate())
	assert.Equal(t, AdminConfigTemplate+StorageConfigTemplate, GetSandboxTemplate())
	assert.Equal(t, AdminConfigTemplate, GetDemoTemplate())
}
