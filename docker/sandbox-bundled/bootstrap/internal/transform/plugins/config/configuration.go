package config

import (
	"encoding/hex"
	"fmt"
	"os"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/kustomize/api/types"

	"github.com/flyteorg/flyte/docker/sandbox-bundled/bootstrap/internal/utils"
)

type ConfigurationNotFound struct {
	path string
}

func (e *ConfigurationNotFound) Error() string {
	return fmt.Sprintf("configuration file not found or is empty: %s", e.path)
}

type Configuration struct {
	ConfigMapName  string
	DeploymentName string
	Namespace      string
	Path           string
}

func NewConfiguration(
	configMapName, deploymentName, namespace, path string,
) (*Configuration, error) {
	// Check that the file is accessible and not empty
	info, err := os.Stat(path)
	if os.IsNotExist(err) || info.Size() == 0 {
		return nil, &ConfigurationNotFound{path}
	}
	if err != nil {
		return nil, err
	}

	return &Configuration{
		ConfigMapName:  configMapName,
		DeploymentName: deploymentName,
		Namespace:      namespace,
		Path:           path,
	}, nil
}

func (c *Configuration) checksum() ([]byte, error) {
	checksum, err := utils.FileChecksum(c.Path)
	if err != nil {
		return nil, err
	}
	return checksum, nil
}

func (c *Configuration) Update(k *types.Kustomization) error {
	// Add configmap args
	if k.ConfigMapGenerator == nil {
		k.ConfigMapGenerator = []types.ConfigMapArgs{}
	}
	k.ConfigMapGenerator = append(k.ConfigMapGenerator, types.ConfigMapArgs{
		GeneratorArgs: types.GeneratorArgs{
			Namespace: c.Namespace,
			Name:      c.ConfigMapName,
			Behavior:  "replace",
			KvPairSources: types.KvPairSources{
				FileSources: []string{fmt.Sprintf("999-extra-config.yaml=%s", c.Path)},
			},
		},
	})

	// Patch deployment to add annotation
	checksum, err := c.checksum()
	if err != nil {
		return err
	}
	patch := appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      c.DeploymentName,
			Namespace: c.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"checksum/extra-configuration": hex.EncodeToString(checksum),
					},
				},
			},
		},
	}
	patchYaml, err := utils.MarshalPatch(&patch, true, true)
	if err != nil {
		return err
	}
	if k.Patches == nil {
		k.Patches = []types.Patch{}
	}
	k.Patches = append(k.Patches, types.Patch{Patch: string(patchYaml)})

	return nil
}
