package configutil

import (
	"html/template"
	"os"

	f "github.com/flyteorg/flyte/flytectl/pkg/filesystemutils"
)

const (
	AdminConfigTemplate = `admin:
  # For GRPC endpoints you might want to use dns:///flyte.myexample.com
  endpoint: {{.Host}}
  insecure: {{.Insecure}}
{{- if .Console}}
console:
  endpoint: {{.Console}}
{{- end}}
`
)

type DataConfig struct {
	Endpoint  string
	AccessKey string
	SecretKey string
}

type ConfigTemplateSpec struct {
	Host       string
	Insecure   bool
	Console    string
	DataConfig *DataConfig
}

var (
	FlytectlConfig = f.FilePathJoin(f.UserHomeDir(), ".flyte", "config-sandbox.yaml")
	ConfigFile     = f.FilePathJoin(f.UserHomeDir(), ".flyte", "config.yaml")
	Kubeconfig     = f.FilePathJoin(f.UserHomeDir(), ".flyte", "k3s", "k3s.yaml")
)

// GetTemplate returns cluster config
func GetTemplate() string {
	return AdminConfigTemplate
}

// SetupConfig download the Flyte sandbox config
func SetupConfig(filename, templateStr string, templateSpec ConfigTemplateSpec) error {
	tmpl := template.New("config")
	tmpl, err := tmpl.Parse(templateStr)
	if err != nil {
		return err
	}
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	return tmpl.Execute(file, templateSpec)
}

// ConfigCleanup will remove the sandbox config from Flyte dir
func ConfigCleanup() error {
	err := os.Remove(FlytectlConfig)
	if err != nil {
		return err
	}
	err = os.RemoveAll(f.FilePathJoin(f.UserHomeDir(), ".flyte", "k3s"))
	if err != nil {
		return err
	}
	return nil
}
