package config

import (
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/kustomize/api/types"

	"github.com/flyteorg/flyte/docker/sandbox-bundled/bootstrap/internal/utils"
)

type ClusterResourceTemplatesNotFound struct {
	dirPath string
}

func (e *ClusterResourceTemplatesNotFound) Error() string {
	return fmt.Sprintf(
		"cluster resource templates directory not found or is empty: %s",
		e.dirPath,
	)
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
	// Check that the directory is accessible
	_, err := os.Stat(dirPath)
	if os.IsNotExist(err) {
		return nil, &ClusterResourceTemplatesNotFound{dirPath}
	} else if err != nil {
		return nil, err
	}

	// Walk directory and collect templates
	var paths []string
	err = filepath.WalkDir(dirPath, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		info, err := os.Stat(p)
		if err != nil {
			return err
		}
		if info.Size() > 0 && filepath.Ext(p) == ".yaml" {
			paths = append(paths, p)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Handle case where no templates were found
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

func (crt *ClusterResourceTemplates) checksum() ([]byte, error) {
	checksum, err := utils.FileCollectionChecksum(crt.Paths)
	if err != nil {
		return nil, err
	}
	return checksum, nil
}

func (crt *ClusterResourceTemplates) Update(k *types.Kustomization) error {
	// Add configmap args
	if k.ConfigMapGenerator == nil {
		k.ConfigMapGenerator = []types.ConfigMapArgs{}
	}
	k.ConfigMapGenerator = append(k.ConfigMapGenerator, types.ConfigMapArgs{
		GeneratorArgs: types.GeneratorArgs{
			Namespace: crt.Namespace,
			Name:      crt.ConfigMapName,
			Behavior:  "replace",
			KvPairSources: types.KvPairSources{
				FileSources: crt.Paths,
			},
		},
	})

	// Patch deployment to add annotation
	checksum, err := crt.checksum()
	if err != nil {
		return err
	}
	patch := appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      crt.DeploymentName,
			Namespace: crt.Namespace,
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
