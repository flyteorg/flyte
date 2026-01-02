package config

import (
	"github.com/flyteorg/flyte/v2/flytestdlib/config"
	"k8s.io/apimachinery/pkg/api/resource"
)

type DataProxyConfig struct {
	Upload   DataProxyUploadConfig   `json:"upload" pflag:",Defines data proxy upload configuration."`
	Download DataProxyDownloadConfig `json:"download" pflag:",Defines data proxy download configuration."`
}

type DataProxyDownloadConfig struct {
	MaxExpiresIn config.Duration `json:"maxExpiresIn" pflag:",Maximum allowed expiration duration."`
}

type DataProxyUploadConfig struct {
	MaxSize               resource.Quantity `json:"maxSize" pflag:",Maximum allowed upload size."`
	MaxExpiresIn          config.Duration   `json:"maxExpiresIn" pflag:",Maximum allowed expiration duration."`
	DefaultFileNameLength int               `json:"defaultFileNameLength" pflag:",Default length for the generated file name if file name not provided in the request."`
	StoragePrefix         string            `json:"storagePrefix" pflag:",Storage prefix to use for all upload requests."`
}
