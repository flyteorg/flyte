package interfaces

import (
	"io/ioutil"

	"github.com/pkg/errors"
)

// Holds details about a cluster used for workflow execution.
type ClusterConfig struct {
	Name     string `json:"name"`
	Endpoint string `json:"endpoint"`
	Auth     Auth   `json:"auth"`
	Enabled  bool   `json:"enabled"`
}

type Auth struct {
	Type      string `json:"type"`
	TokenPath string `json:"tokenPath"`
	CertPath  string `json:"certPath"`
}

type ClusterSelectionStrategy string

var (
	ClusterSelectionRandom ClusterSelectionStrategy
)

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
	ClusterConfigs   []ClusterConfig          `json:"clusterConfigs"`
	ClusterSelection ClusterSelectionStrategy `json:"clusterSelectionStrategy"`
}

// Provides values set in runtime configuration files.
// These files can be changed without requiring a full server restart.
type ClusterConfiguration interface {
	// Returns clusters defined in runtime configuration files.
	GetClusterConfigs() []ClusterConfig

	// The cluster selection strategy setting
	GetClusterSelectionStrategy() ClusterSelectionStrategy
}
