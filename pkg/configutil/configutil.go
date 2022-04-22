package configutil

import (
	"html/template"
	"os"

	f "github.com/flyteorg/flytectl/pkg/filesystemutils"
)

const (
	AdminConfigTemplate = `admin:
  # For GRPC endpoints you might want to use dns:///flyte.myexample.com
  endpoint: {{.Host}}
  authType: Pkce
  insecure: {{.Insecure}}
logger:
  show-source: true
  level: 0`
	StorageConfigTemplate = `
storage:
  connection:
    access-key: minio
    auth-type: accesskey
    disable-ssl: true
    endpoint: http://localhost:30084
    region: us-east-1
    secret-key: miniostorage
  type: minio
  container: "my-s3-bucket"
  enable-multicontainer: true`
	StorageS3ConfigTemplate = `
storage:
  type: stow	
  stow:
    kind: s3
    config:
      auth_type: iam
      region: <replace> # Example: us-east-2
  container: <replace> # Example my-bucket. Flyte k8s cluster / service account for execution should have read access to this bucket`
	StorageGCSConfigTemplate = `
storage:
  type: stow	
  stow:
    kind: google
    config:
      json: ""
      project_id: <replace-me> # TODO: replace <project-id> with the GCP project ID
      scopes: https://www.googleapis.com/auth/devstorage.read_write
  container: <replace> # Example my-bucket. Flyte k8s cluster / service account for execution should have read access to this bucket`
)

type ConfigTemplateSpec struct {
	Host     string
	Insecure bool
}

var (
	FlytectlConfig = f.FilePathJoin(f.UserHomeDir(), ".flyte", "config-sandbox.yaml")
	ConfigFile     = f.FilePathJoin(f.UserHomeDir(), ".flyte", "config.yaml")
	Kubeconfig     = f.FilePathJoin(f.UserHomeDir(), ".flyte", "k3s", "k3s.yaml")
)

// GetSandboxTemplate return sandbox cluster config
func GetSandboxTemplate() string {
	return AdminConfigTemplate + StorageConfigTemplate
}

// GetDemoTemplate return demo cluster config
func GetDemoTemplate() string {
	return AdminConfigTemplate
}

// GetAWSCloudTemplate return aws Flyte config with storage config
func GetAWSCloudTemplate() string {
	return AdminConfigTemplate + StorageS3ConfigTemplate
}

// GetGoogleCloudTemplate return google Flyte config with storage config
func GetGoogleCloudTemplate() string {
	return AdminConfigTemplate + StorageGCSConfigTemplate
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
