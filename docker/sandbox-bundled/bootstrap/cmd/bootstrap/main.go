package main

import (
	"log"
	"net"
	"os"
	"path/filepath"

	"github.com/flyteorg/flyte/docker/sandbox-bundled/bootstrap/internal/transform"
	"github.com/flyteorg/flyte/docker/sandbox-bundled/bootstrap/internal/transform/plugins/config"
	"github.com/flyteorg/flyte/docker/sandbox-bundled/bootstrap/internal/transform/plugins/vars"
)

const (
	configDirPath                         = "/var/lib/flyte/config"
	configurationConfigMapName            = "flyte-sandbox-extra-config"
	clusterResourceTemplatesConfigMapName = "flyte-sandbox-extra-cluster-resource-templates"
	deploymentName                        = "flyte-sandbox"
	devModeEnvVar                         = "FLYTE_DEV"
	disableConnectorModeEnvVar            = "DISABLE_CONNECTOR"
	dockerHost                            = "host.docker.internal"
	namespace                             = "flyte"

	// Template paths
	devTemplatePath           = "/var/lib/rancher/k3s/server/manifests-staging/dev.yaml"
	fullTemplatePath          = "/var/lib/rancher/k3s/server/manifests-staging/complete.yaml"
	fullConnectorTemplatePath = "/var/lib/rancher/k3s/server/manifests-staging/complete-connector.yaml"
	renderedManifestPath      = "/var/lib/rancher/k3s/server/manifests/flyte.yaml"
)

func main() {
	var tmplPath string
	var tPlugins []transform.Plugin

	if os.Getenv(devModeEnvVar) == "True" {
		tmplPath = devTemplatePath
	} else {
		// If we are not running in dev mode, look for user-specified configuration
		// to load into the sandbox deployment
		tmplPath = fullConnectorTemplatePath
		if os.Getenv(disableConnectorModeEnvVar) == "True" {
			tmplPath = fullTemplatePath
		}

		cOpts := config.LoaderOpts{
			ConfigurationConfigMapName:            configurationConfigMapName,
			ClusterResourceTemplatesConfigMapName: clusterResourceTemplatesConfigMapName,
			DeploymentName:                        deploymentName,
			Namespace:                             namespace,
			DirPath:                               configDirPath,
		}
		c, err := config.NewLoader(&cOpts)
		if err != nil {
			log.Fatalf("failed to initialize config loader: %s", err)
		}
		tPlugins = append(tPlugins, c)
	}

	// Replace template variables
	v := vars.NewVars(map[string]vars.ValueGetter{
		"%{HOST_GATEWAY_IP}%": func() (string, error) {
			addrs, err := net.LookupHost(dockerHost)
			if err != nil {
				return "", err
			}
			return addrs[0], nil
		},
	})
	tPlugins = append(tPlugins, v)

	// Render final manifest and write out
	tmpl, err := os.ReadFile(tmplPath)
	if err != nil {
		log.Fatalf("failed to read manifest: %s", err)
	}
	t := transform.NewTransformer(tPlugins...)
	rendered, err := t.Transform(tmpl)
	if err != nil {
		log.Fatalf("failed to apply transformations: %s", err)
	}
	if err := os.MkdirAll(filepath.Dir(renderedManifestPath), 0755); err != nil {
		log.Fatalf("failed to create destination directory: %s", err)
	}
	if err := os.WriteFile(renderedManifestPath, rendered, 0644); err != nil {
		log.Fatalf("failed to write rendered manifest: %s", err)
	}
}
