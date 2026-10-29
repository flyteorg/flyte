package flytek8s

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	v1 "k8s.io/api/core/v1"

	pluginserrors "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/errors"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/flytek8s/config"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
)

// Environment variables handed to a workload that requests the Apple Silicon accelerator.
// FLYTE_APPLE_SILICON_* is what the Flyte runtime in the container uses to attach to the host daemon; the
// OPENAI_* pair lets an unmodified OpenAI / MLX client reach the same daemon with no setup.
const (
	AppleSiliconEndpointEnvVar = "FLYTE_APPLE_SILICON_ENDPOINT"
	AppleSiliconWorkloadEnvVar = "FLYTE_APPLE_SILICON_WORKLOAD"
	AppleSiliconTokenEnvVar    = "FLYTE_APPLE_SILICON_TOKEN"
	OpenAIBaseURLEnvVar        = "OPENAI_BASE_URL"
	OpenAIAPIKeyEnvVar         = "OPENAI_API_KEY"

	// appleSiliconTokenVersion prefixes every workload token so the bridge can reject tokens minted by an
	// incompatible scheme instead of failing the HMAC check with a less useful message.
	appleSiliconTokenVersion = "as1"

	// appleSiliconPlatform is how the devbox reports an Apple Silicon host (flyte-sdk `_host.host_platform`).
	appleSiliconPlatform = "darwin/arm64"
)

// IsAppleSilicon reports whether the accelerator asks for the devbox-local Apple Silicon GPU.
func IsAppleSilicon(accelerator *core.GPUAccelerator) bool {
	return accelerator != nil && accelerator.GetDeviceClass() == core.GPUAccelerator_APPLE_SILICON
}

// MintAppleSiliconToken returns the credential a workload presents to the host Apple Silicon bridge. It binds the
// workload id to the bridge's shared secret, so a pod can only reach the daemon started for it.
// The bridge (flyte-sdk, flyte._internal.apple_silicon) verifies the same format.
func MintAppleSiliconToken(secret, workloadID string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(workloadID))
	return fmt.Sprintf("%s.%s.%s", appleSiliconTokenVersion, workloadID, hex.EncodeToString(mac.Sum(nil)))
}

// appleSiliconNotAvailable builds the APPLE_SILICON_NOT_AVAILABLE error, naming the host when the devbox told us what
// it is running on. A backend that is not a devbox at all (any real cluster) has no host to name.
func appleSiliconNotAvailable(hostPlatform string) error {
	if hostPlatform == appleSiliconPlatform {
		return pluginserrors.Errorf(pluginserrors.AppleSiliconNotAvailable,
			"gpu='AppleSilicon' needs the devbox's Apple Silicon bridge, but this devbox was started without it; "+
				"restart it with `flyte start devbox --apple-silicon`.")
	}
	if hostPlatform != "" {
		return pluginserrors.Errorf(pluginserrors.AppleSiliconNotAvailable,
			"gpu='AppleSilicon' requires an Apple Silicon devbox; this host is %s.", hostPlatform)
	}
	return pluginserrors.Errorf(pluginserrors.AppleSiliconNotAvailable,
		"gpu='AppleSilicon' requires an Apple Silicon devbox; this backend has no Apple Silicon bridge. "+
			"Apple Silicon is a devbox-local accelerator and is never cluster-schedulable: "+
			"use a cluster accelerator such as gpu='A100:1' here.")
}

// ValidateAppleSiliconBridge fails fast with APPLE_SILICON_NOT_AVAILABLE when this backend cannot serve Apple Silicon.
func ValidateAppleSiliconBridge() error {
	cfg := config.GetK8sPluginConfig().AppleSiliconBridge
	if !configured(cfg.Endpoint) || !configured(cfg.TokenSecret) {
		return appleSiliconNotAvailable(renderedOrEmpty(cfg.HostPlatform))
	}
	return nil
}

// configured reports whether a bridge config value was actually set. The devbox chart ships
// `%{...}%` placeholders that its bootstrap renders; a config used without that pass still carries
// the literal placeholder, which must not count as a bridge.
func configured(value string) bool {
	return value != "" && !strings.Contains(value, "%{")
}

func renderedOrEmpty(value string) string {
	if configured(value) {
		return value
	}
	return ""
}

// stripAcceleratorResources drops every accelerator resource from the container. An Apple Silicon workload
// carries a GPU count of 1 like any other accelerator request, but there is no device plugin to
// satisfy it: left in place it would keep the pod Pending forever.
func stripAcceleratorResources(resources *v1.ResourceRequirements) {
	names := getAllAcceleratorResourceNames()
	names[resourceGPU] = struct{}{}
	for name := range names {
		delete(resources.Requests, name)
		delete(resources.Limits, name)
	}
}

func hasEnv(container *v1.Container, name string) bool {
	for _, e := range container.Env {
		if e.Name == name {
			return true
		}
	}
	return false
}

func setEnvIfUnset(container *v1.Container, name, value string) {
	if !hasEnv(container, name) {
		container.Env = append(container.Env, v1.EnvVar{Name: name, Value: value})
	}
}

// ApplyAppleSiliconBridge rewires a pod that requests the APPLE_SILICON accelerator onto the
// devbox's host Apple Silicon bridge: accelerator resource requests are removed from every container and
// the primary container gets the bridge endpoint plus a credential scoped to workloadID. Values
// the user set explicitly (e.g. their own OPENAI_BASE_URL) are left alone.
func ApplyAppleSiliconBridge(podSpec *v1.PodSpec, primaryContainerName, workloadID string) error {
	if err := ValidateAppleSiliconBridge(); err != nil {
		return err
	}
	cfg := config.GetK8sPluginConfig().AppleSiliconBridge
	endpoint := strings.TrimRight(cfg.Endpoint, "/")
	token := MintAppleSiliconToken(cfg.TokenSecret, workloadID)

	var primary *v1.Container
	for i := range podSpec.InitContainers {
		stripAcceleratorResources(&podSpec.InitContainers[i].Resources)
	}
	for i := range podSpec.Containers {
		stripAcceleratorResources(&podSpec.Containers[i].Resources)
		if podSpec.Containers[i].Name == primaryContainerName {
			primary = &podSpec.Containers[i]
		}
	}
	if primary == nil {
		return pluginserrors.Errorf(pluginserrors.BadTaskSpecification,
			"invalid TaskSpecification, primary container [%s] not defined", primaryContainerName)
	}

	setEnvIfUnset(primary, AppleSiliconEndpointEnvVar, endpoint)
	setEnvIfUnset(primary, AppleSiliconWorkloadEnvVar, workloadID)
	setEnvIfUnset(primary, AppleSiliconTokenEnvVar, token)
	// The OpenAI pair goes in together or not at all: setting only one would send the bridge token to the
	// user's own OPENAI_BASE_URL, or the user's own key to the bridge.
	if !hasEnv(primary, OpenAIBaseURLEnvVar) && !hasEnv(primary, OpenAIAPIKeyEnvVar) {
		primary.Env = append(primary.Env,
			v1.EnvVar{Name: OpenAIBaseURLEnvVar, Value: endpoint + "/v1"},
			v1.EnvVar{Name: OpenAIAPIKeyEnvVar, Value: token})
	}
	return nil
}
