package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"

	"github.com/flyteorg/flyte/docker/devbox-bundled/bootstrap/internal/transform"
	"github.com/flyteorg/flyte/docker/devbox-bundled/bootstrap/internal/transform/plugins/config"
	"github.com/flyteorg/flyte/docker/devbox-bundled/bootstrap/internal/transform/plugins/vars"
)

const (
	configDirPath              = "/var/lib/flyte/config"
	configurationConfigMapName = "flyte-devbox-extra-config"
	deploymentName             = "flyte-binary"
	devModeEnvVar              = "FLYTE_DEV"
	dockerHost                 = "host.docker.internal"

	// Set by `flyte start devbox` when it runs the host-side Apple Silicon bridge (Apple Silicon only).
	appleSiliconBridgePortEnvVar        = "FLYTE_DEVBOX_APPLE_SILICON_PORT"
	appleSiliconBridgeTokenSecretEnvVar = "FLYTE_DEVBOX_APPLE_SILICON_TOKEN_SECRET"
	hostPlatformEnvVar                  = "FLYTE_DEVBOX_HOST_PLATFORM"
	namespace                           = "flyte"

	// Template paths
	devTemplatePath      = "/var/lib/rancher/k3s/server/manifests-staging/dev.yaml"
	fullTemplatePath     = "/var/lib/rancher/k3s/server/manifests-staging/complete.yaml"
	renderedManifestPath = "/var/lib/rancher/k3s/server/manifests/flyte.yaml"
)

// getNodeIP returns the preferred outbound IP address of the node.
// This is used to create Kubernetes Endpoints that point back to
// services running on the host (e.g., embedded PostgreSQL).
func getNodeIP() (string, error) {
	// Use a UDP dial to determine the preferred outbound IP.
	// No actual connection is made.
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		// Fallback: scan interfaces for a non-loopback address
		return getFirstNonLoopbackIP()
	}
	defer conn.Close() //nolint:errcheck
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String(), nil
}

func getFirstNonLoopbackIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
			return ipnet.IP.String(), nil
		}
	}
	return "", fmt.Errorf("no suitable IP address found")
}

func lookupHostGatewayIP() (string, error) {
	addrs, err := net.LookupHost(dockerHost)
	if err != nil {
		return "", err
	}
	return addrs[0], nil
}

// appleSiliconBridgeEndpoint is the host Apple Silicon bridge's URL as seen from pods, or "" when the devbox was
// started without one. Pods cannot resolve host.docker.internal (CoreDNS does not read the
// node's /etc/hosts), so the endpoint uses the host gateway IP directly.
func appleSiliconBridgeEndpoint(lookupIP func() (string, error)) (string, error) {
	port := os.Getenv(appleSiliconBridgePortEnvVar)
	if port == "" {
		return "", nil
	}
	ip, err := lookupIP()
	if err != nil {
		return "", err
	}
	return "http://" + net.JoinHostPort(ip, port), nil
}

func main() {
	var tmplPath string
	var tPlugins []transform.Plugin

	if os.Getenv(devModeEnvVar) == "True" {
		tmplPath = devTemplatePath
	} else {
		// If we are not running in dev mode, look for user-specified configuration
		// to load into the sandbox deployment
		tmplPath = fullTemplatePath

		cOpts := config.LoaderOpts{
			ConfigurationConfigMapName: configurationConfigMapName,
			DeploymentName:             deploymentName,
			Namespace:                  namespace,
			DirPath:                    configDirPath,
		}
		c, err := config.NewLoader(&cOpts)
		if err != nil {
			log.Fatalf("failed to initialize config loader: %s", err)
		}
		tPlugins = append(tPlugins, c)
	}

	// Replace template variables
	v := vars.NewVars(map[string]vars.ValueGetter{
		"%{HOST_GATEWAY_IP}%": lookupHostGatewayIP,
		"%{APPLE_SILICON_BRIDGE_ENDPOINT}%": func() (string, error) {
			return appleSiliconBridgeEndpoint(lookupHostGatewayIP)
		},
		"%{APPLE_SILICON_BRIDGE_TOKEN_SECRET}%": func() (string, error) {
			return os.Getenv(appleSiliconBridgeTokenSecretEnvVar), nil
		},
		"%{DEVBOX_HOST_PLATFORM}%": func() (string, error) {
			return os.Getenv(hostPlatformEnvVar), nil
		},
		"%{NODE_IP}%": func() (string, error) {
			return getNodeIP()
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
