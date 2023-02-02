package config

import (
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/flyteorg/flyte/docker/sandbox-bundled/bootstrap/internal/utils"
	"github.com/stretchr/testify/assert"
)

func TestLoaderHappy(t *testing.T) {
	// Initialize config loader
	cOpts := LoaderOpts{
		ConfigurationConfigMapName:            "test-config",
		ClusterResourceTemplatesConfigMapName: "test-cluster-resource-templates",
		DeploymentName:                        "test-deployment",
		Namespace:                             "test",
		DirPath:                               filepath.Join("testdata", "happy"),
	}
	c, err := NewLoader(&cOpts)
	if err != nil {
		t.Fatal(err)
	}

	// Read in base manifest
	base, err := os.ReadFile(filepath.Join("testdata", "base.yaml"))
	if err != nil {
		t.Fatal(err)
	}

	// Apply transform
	rendered, err := c.Transform(base)
	if err != nil {
		t.Fatal(err)
	}

	// Expected values
	configurationAbsPath, err := filepath.Abs(filepath.Join(cOpts.DirPath, "config.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	configurationChecksum, err := utils.FileChecksum(configurationAbsPath)
	if err != nil {
		t.Fatal(err)
	}
	clusterResourceTemplatesAbsPath, err := filepath.Abs(filepath.Join(
		cOpts.DirPath,
		"cluster-resource-templates",
		"resource.yaml",
	))
	if err != nil {
		t.Fatal(err)
	}
	clusterResourceTemplatesChecksum, err := utils.FileCollectionChecksum(
		[]string{clusterResourceTemplatesAbsPath},
	)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(
		t,
		string(rendered),
		fmt.Sprintf(`apiVersion: v1
data:
  resource.yaml: |
    foo: bar
kind: ConfigMap
metadata:
  name: test-cluster-resource-templates
  namespace: test
---
apiVersion: v1
data:
  999-extra-config.yaml: |
    ham: spam
kind: ConfigMap
metadata:
  name: test-config
  namespace: test
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
  namespace: test
spec:
  replicas: 1
  selector:
    matchLabels:
      app: test
  template:
    metadata:
      annotations:
        checksum/extra-cluster-resource-templates: %s
        checksum/extra-configuration: %s
      labels:
        app: test
    spec:
      containers:
      - image: test:test
        name: test
`, hex.EncodeToString(clusterResourceTemplatesChecksum), hex.EncodeToString(configurationChecksum)),
		"YAML strings should match",
	)
}

func TestLoaderEmptyDir(t *testing.T) {
	// Initialize config loader
	cOpts := LoaderOpts{
		ConfigurationConfigMapName:            "test-config",
		ClusterResourceTemplatesConfigMapName: "test-cluster-resource-templates",
		DeploymentName:                        "test-deployment",
		Namespace:                             "test",
		DirPath:                               filepath.Join("testdata", "emptydir"),
	}
	c, err := NewLoader(&cOpts)
	if err != nil {
		t.Fatal(err)
	}
	assert.Nil(t, c.configuration)
	assert.Nil(t, c.clusterResourceTemplates)

	// Read in base manifest
	base, err := os.ReadFile(filepath.Join("testdata", "base.yaml"))
	if err != nil {
		t.Fatal(err)
	}

	// Apply transform
	rendered, err := c.Transform(base)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, base, rendered, "YAML strings should match")
}

func TestLoaderEmptyFiles(t *testing.T) {
	// Initialize config loader
	cOpts := LoaderOpts{
		ConfigurationConfigMapName:            "test-config",
		ClusterResourceTemplatesConfigMapName: "test-cluster-resource-templates",
		DeploymentName:                        "test-deployment",
		Namespace:                             "test",
		DirPath:                               filepath.Join("testdata", "emptyfile"),
	}
	c, err := NewLoader(&cOpts)
	if err != nil {
		t.Fatal(err)
	}
	assert.Nil(t, c.configuration)
	assert.Nil(t, c.clusterResourceTemplates)

	// Read in base manifest
	base, err := os.ReadFile(filepath.Join("testdata", "base.yaml"))
	if err != nil {
		t.Fatal(err)
	}

	// Apply transform
	rendered, err := c.Transform(base)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, base, rendered, "YAML strings should match")
}
