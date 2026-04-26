package config

import (
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"

	"github.com/flyteorg/flyte/docker/devbox-bundled/bootstrap/internal/utils"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/kustomize/api/types"
)

type ClusterResourceTemplatesNotFound struct {
	path string
}

func (e *ClusterResourceTemplatesNotFound) Error() string {
	return fmt.Sprintf("cluster resource templates directory not found or is empty: %s", e.path)
}

type ClusterResourceTemplates struct {
	ConfigMapName  string
	DeploymentName string
	Namespace      string
	Paths          []string
}

func NewClusterResourceTemplates(
	configMapName, deploymentName, namespace, dirPath string,
) (*ClusterResourceTemplates, error) {
	entries, err := os.ReadDir(dirPath)
	if os.IsNotExist(err) {
		return nil, &ClusterResourceTemplatesNotFound{dirPath}
	}
	if err != nil {
		return nil, err
	}

	var paths []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			return nil, err
		}
		if info.Size() > 0 {
			paths = append(paths, filepath.Join(dirPath, entry.Name()))
		}
	}

	if len(paths) == 0 {
		return nil, &ClusterResourceTemplatesNotFound{dirPath}
	}

	return &ClusterResourceTemplates{
		ConfigMapName:  configMapName,
		DeploymentName: deploymentName,
		Namespace:      namespace,
		Paths:          paths,
	}, nil
}

func (c *ClusterResourceTemplates) Update(k *types.Kustomization) error {
	// Build ConfigMap data from files
	data := map[string]string{}
	for _, path := range c.Paths {
		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("reading cluster resource template %s: %w", path, err)
		}
		data[filepath.Base(path)] = string(content)
	}

	// Patch the ConfigMap directly with file contents
	cm := corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ConfigMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      c.ConfigMapName,
			Namespace: c.Namespace,
		},
		Data: data,
	}
	cmPatchYaml, err := utils.MarshalPatch(&cm, true, true)
	if err != nil {
		return err
	}

	// Patch the Deployment with a checksum annotation
	checksum, err := utils.FileCollectionChecksum(c.Paths)
	if err != nil {
		return err
	}
	deployPatch := appsv1.Deployment{
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
						"checksum/extra-cluster-resource-templates": hex.EncodeToString(checksum),
					},
				},
			},
		},
	}
	deployPatchYaml, err := utils.MarshalPatch(&deployPatch, true, true)
	if err != nil {
		return err
	}

	if k.Patches == nil {
		k.Patches = []types.Patch{}
	}
	k.Patches = append(k.Patches,
		types.Patch{Patch: string(cmPatchYaml)},
		types.Patch{Patch: string(deployPatchYaml)},
	)

	return nil
}
