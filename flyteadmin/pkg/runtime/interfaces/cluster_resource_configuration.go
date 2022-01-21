package interfaces

import (
	"time"

	"github.com/flyteorg/flytestdlib/config"
)

type DataSourceValueFrom struct {
	EnvVar   string `json:"env"`
	FilePath string `json:"filePath"`
}

type DataSource struct {
	Value     string              `json:"value"`
	ValueFrom DataSourceValueFrom `json:"valueFrom"`
}

type DomainName = string

type TemplateData = map[string]DataSource

type ClusterResourceConfig struct {
	TemplatePath string `json:"templatePath"`
	// TemplateData maps template keys e.g. my_super_secret_password to a data source
	// which is then substituted in cluster resource templated config files wherever
	// {{ my_super_secret_password }} appears.
	TemplateData    TemplateData    `json:"templateData"`
	RefreshInterval config.Duration `json:"refreshInterval"`
	// Like TemplateData above, this also specifies template values as defaults to be substituted for specific domains
	// or for all domains.
	// For example:
	/*
		defaultData:
		  production:
		    foo:
		      value: "bar"
		    foofoo:
		      valueFrom:
		        env: SHELL
		  staging:
		    foo:
		      value: "baz"
	*/
	CustomData           map[DomainName]TemplateData `json:"customData"`
	StandaloneDeployment bool                        `json:"standaloneDeployment" pflag:", Whether the cluster resource sync is running in a standalone deployment and should call flyteadmin service endpoints"`
}

type ClusterResourceConfiguration interface {
	GetTemplatePath() string
	GetTemplateData() map[string]DataSource
	GetRefreshInterval() time.Duration
	GetCustomTemplateData() map[DomainName]TemplateData
	IsStandaloneDeployment() bool
}
