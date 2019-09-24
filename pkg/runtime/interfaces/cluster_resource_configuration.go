package interfaces

import (
	"time"

	"github.com/lyft/flytestdlib/config"
)

type DataSourceValueFrom struct {
	EnvVar   string `json:"env"`
	FilePath string `json:"filePath"`
}

type DataSource struct {
	Value     string              `json:"value"`
	ValueFrom DataSourceValueFrom `json:"valueFrom"`
}

type ClusterResourceConfig struct {
	TemplatePath string `json:"templatePath"`
	// TemplateData maps template keys e.g. my_super_secret_password to a data source
	// which is then substituted in cluster resource templated config files wherever
	// {{ my_super_secret_password }} appears.
	TemplateData    map[string]DataSource `json:"templateData"`
	RefreshInterval config.Duration       `json:"refreshInterval"`
}

type ClusterResourceConfiguration interface {
	GetTemplatePath() string
	GetTemplateData() map[string]DataSource
	GetRefreshInterval() time.Duration
}
