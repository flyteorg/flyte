package config

import (
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"

	"github.com/flyteorg/flyte/docker/devbox-bundled/bootstrap/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoaderHappy(t *testing.T) {
	cOpts := LoaderOpts{
		ConfigurationConfigMapName:            "test-config",
		ClusterResourceTemplatesConfigMapName: "test-cluster-resource-templates",
		DeploymentName:                        "test-deployment",
		Namespace:                             "test",
		DirPath:                               filepath.Join("testdata", "happy"),
	}
	c, err := NewLoader(&cOpts)
	require.NoError(t, err)

	base, err := os.ReadFile(filepath.Join("testdata", "base.yaml"))
	require.NoError(t, err)

	rendered, err := c.Transform(base)
	require.NoError(t, err)

	configAbsPath, err := filepath.Abs(filepath.Join("testdata", "happy", "config.yaml"))
	require.NoError(t, err)
	configChecksum, err := utils.FileChecksum(configAbsPath)
	require.NoError(t, err)

	crtAbsPath, err := filepath.Abs(filepath.Join("testdata", "happy", "cluster-resource-templates", "resource.yaml"))
	require.NoError(t, err)
	crtChecksum, err := utils.FileCollectionChecksum([]string{crtAbsPath})
	require.NoError(t, err)

	expected := "apiVersion: v1\n" +
		"data:\n" +
		"  resource.yaml: |\n" +
		"    foo: bar\n" +
		"kind: ConfigMap\n" +
		"metadata:\n" +
		"  name: test-cluster-resource-templates\n" +
		"  namespace: test\n" +
		"---\n" +
		"apiVersion: v1\n" +
		"data:\n" +
		"  999-extra-config.yaml: |\n" +
		"    ham: spam\n" +
		"kind: ConfigMap\n" +
		"metadata:\n" +
		"  name: test-config\n" +
		"  namespace: test\n" +
		"---\n" +
		"apiVersion: apps/v1\n" +
		"kind: Deployment\n" +
		"metadata:\n" +
		"  name: test-deployment\n" +
		"  namespace: test\n" +
		"spec:\n" +
		"  replicas: 1\n" +
		"  selector:\n" +
		"    matchLabels:\n" +
		"      app: test\n" +
		"  template:\n" +
		"    metadata:\n" +
		"      annotations:\n" +
		"        checksum/extra-cluster-resource-templates: " + hex.EncodeToString(crtChecksum) + "\n" +
		"        checksum/extra-configuration: " + hex.EncodeToString(configChecksum) + "\n" +
		"      labels:\n" +
		"        app: test\n" +
		"    spec:\n" +
		"      containers:\n" +
		"      - image: test:test\n" +
		"        name: test\n"

	assert.Equal(t, expected, string(rendered))
}

func TestLoaderEmptyDir(t *testing.T) {
	cOpts := LoaderOpts{
		ConfigurationConfigMapName:            "test-config",
		ClusterResourceTemplatesConfigMapName: "test-cluster-resource-templates",
		DeploymentName:                        "test-deployment",
		Namespace:                             "test",
		DirPath:                               filepath.Join("testdata", "emptydir"),
	}
	c, err := NewLoader(&cOpts)
	require.NoError(t, err)
	assert.Nil(t, c.configuration)
	assert.Nil(t, c.clusterResourceTemplates)

	base, err := os.ReadFile(filepath.Join("testdata", "base.yaml"))
	require.NoError(t, err)

	rendered, err := c.Transform(base)
	require.NoError(t, err)
	assert.Equal(t, base, rendered)
}

func TestLoaderEmptyFiles(t *testing.T) {
	cOpts := LoaderOpts{
		ConfigurationConfigMapName:            "test-config",
		ClusterResourceTemplatesConfigMapName: "test-cluster-resource-templates",
		DeploymentName:                        "test-deployment",
		Namespace:                             "test",
		DirPath:                               filepath.Join("testdata", "emptyfile"),
	}
	c, err := NewLoader(&cOpts)
	require.NoError(t, err)
	assert.Nil(t, c.configuration)
	assert.Nil(t, c.clusterResourceTemplates)

	base, err := os.ReadFile(filepath.Join("testdata", "base.yaml"))
	require.NoError(t, err)

	rendered, err := c.Transform(base)
	require.NoError(t, err)
	assert.Equal(t, base, rendered)
}
