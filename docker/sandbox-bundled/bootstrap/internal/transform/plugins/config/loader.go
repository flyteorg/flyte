package config

import (
	"errors"
	"os"
	"path/filepath"

	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/filesys"
	"sigs.k8s.io/yaml"
)

type LoaderOpts struct {
	ConfigurationConfigMapName            string
	ClusterResourceTemplatesConfigMapName string
	DeploymentName                        string
	Namespace                             string
	DirPath                               string
}

type Loader struct {
	configuration            *Configuration
	clusterResourceTemplates *ClusterResourceTemplates
}

func NewLoader(opts *LoaderOpts) (*Loader, error) {
	var err error
	loader := Loader{}

	absDirPath, err := filepath.Abs(opts.DirPath)
	if err != nil {
		return nil, err
	}

	loader.configuration, err = NewConfiguration(
		opts.ConfigurationConfigMapName,
		opts.DeploymentName,
		opts.Namespace,
		filepath.Join(absDirPath, "config.yaml"),
	)
	var configurationNotFound *ConfigurationNotFound
	if err != nil && !errors.As(err, &configurationNotFound) {
		return nil, err
	}

	loader.clusterResourceTemplates, err = NewClusterResourceTemplates(
		opts.ClusterResourceTemplatesConfigMapName,
		opts.DeploymentName,
		opts.Namespace,
		filepath.Join(absDirPath, "cluster-resource-templates"),
	)
	var clusterResourceTemplatesNotFound *ClusterResourceTemplatesNotFound
	if err != nil && !errors.As(err, &clusterResourceTemplatesNotFound) {
		return nil, err
	}

	return &loader, nil
}

func (cl *Loader) Transform(data []byte) ([]byte, error) {
	if cl.configuration == nil && cl.clusterResourceTemplates == nil {
		return data, nil
	}

	workDir, err := os.MkdirTemp("", "")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(workDir)

	baseManifestPath := filepath.Join(workDir, "base.yaml")
	if err := os.WriteFile(baseManifestPath, data, 0644); err != nil {
		return nil, err
	}

	k := types.Kustomization{Resources: []string{baseManifestPath}}

	if cl.clusterResourceTemplates != nil {
		if err := cl.clusterResourceTemplates.Update(&k); err != nil {
			return nil, err
		}
	}

	if cl.configuration != nil {
		if err := cl.configuration.Update(&k); err != nil {
			return nil, err
		}
	}

	kyaml, err := yaml.Marshal(k)
	if err != nil {
		return nil, err
	}
	if err := os.WriteFile(filepath.Join(workDir, "kustomization.yaml"), kyaml, 0644); err != nil {
		return nil, err
	}

	opts := krusty.MakeDefaultOptions()
	opts.DoLegacyResourceSort = true
	opts.LoadRestrictions = types.LoadRestrictionsNone
	kustomizer := krusty.MakeKustomizer(opts)
	resMap, err := kustomizer.Run(filesys.MakeFsOnDisk(), workDir)
	if err != nil {
		return nil, err
	}
	out, err := resMap.AsYaml()
	if err != nil {
		return nil, err
	}
	return out, nil
}
