package k8s

import (
	"fmt"
	"io/ioutil"

	"github.com/pkg/errors"
	restclient "k8s.io/client-go/rest"
)

type ClusterConfig struct {
	Name     string `json:"name" pflag:",Friendly name of the remote cluster"`
	Endpoint string `json:"endpoint" pflag:", Remote K8s cluster endpoint"`
	Auth     Auth   `json:"auth" pflag:"-, Auth setting for the cluster"`
	Enabled  bool   `json:"enabled" pflag:", Boolean flag to enable or disable"`
}

type Auth struct {
	TokenPath  string `json:"tokenPath" pflag:", Token path"`
	CaCertPath string `json:"caCertPath" pflag:", Certificate path"`
}

func (auth Auth) GetCA() ([]byte, error) {
	cert, err := ioutil.ReadFile(auth.CaCertPath)
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

// KubeClientConfig ...
func KubeClientConfig(host string, auth Auth) (*restclient.Config, error) {
	tokenString, err := auth.GetToken()
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to get auth token: %+v", err))
	}

	caCert, err := auth.GetCA()
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to get auth CA: %+v", err))
	}

	tlsClientConfig := restclient.TLSClientConfig{}
	tlsClientConfig.CAData = caCert
	return &restclient.Config{
		Host:            host,
		TLSClientConfig: tlsClientConfig,
		BearerToken:     tokenString,
	}, nil
}
