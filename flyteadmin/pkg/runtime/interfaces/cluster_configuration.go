package interfaces

import (
	"io/ioutil"

	"github.com/flyteorg/flyteadmin/pkg/config"

	"github.com/pkg/errors"
)

// Holds details about a cluster used for workflow execution.
type ClusterConfig struct {
	Name             string                   `json:"name"`
	Endpoint         string                   `json:"endpoint"`
	Auth             Auth                     `json:"auth"`
	Enabled          bool                     `json:"enabled"`
	KubeClientConfig *config.KubeClientConfig `json:"kubeClientConfig,omitempty"`
}

type Auth struct {
	Type      string `json:"type"`
	TokenPath string `json:"tokenPath"`
	CertPath  string `json:"certPath"`
}

type ClusterEntity struct {
	ID     string  `json:"id"`
	Weight float32 `json:"weight"`
}

func (auth Auth) GetCA() ([]byte, error) {
	cert, err := ioutil.ReadFile(auth.CertPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read k8s CA cert from configured path")
	}
	return cert, nil
}

func (auth Auth) GetToken() (string, error) {
	token, err := ioutil.ReadFile(auth.TokenPath)
	if err != nil {
		return "", errors.Wrap(err, "failed to read k8s bearer token from configured path")
	}
	return string(token), nil
}

type Clusters struct {
	ClusterConfigs        []ClusterConfig            `json:"clusterConfigs"`
	LabelClusterMap       map[string][]ClusterEntity `json:"labelClusterMap"`
	DefaultExecutionLabel string                     `json:"defaultExecutionLabel"`
}

//go:generate mockery -name ClusterConfiguration -case=underscore -output=../mocks -case=underscore

// Provides values set in runtime configuration files.
// These files can be changed without requiring a full server restart.
type ClusterConfiguration interface {
	// Returns clusters defined in runtime configuration files.
	GetClusterConfigs() []ClusterConfig

	// Returns label cluster map for routing
	GetLabelClusterMap() map[string][]ClusterEntity

	// Returns default execution label used as fallback if no execution cluster was explicitly defined.
	GetDefaultExecutionLabel() string
}
