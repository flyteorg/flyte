package plugin

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core"

	"github.com/unionai/flyte/fasttask/plugin/pb"
)

var nonLowerAlphanumericDashOrDotRegex = regexp.MustCompile(`[^a-z0-9\.\-]+`)

// isValidEnvironmentSpec validates the FastTaskEnvironmentSpec
func isValidEnvironmentSpec(executionEnvironmentID core.ExecutionEnvID, fastTaskEnvironmentSpec *pb.FastTaskEnvironmentSpec) error {
	if len(executionEnvironmentID.Name) == 0 {
		return fmt.Errorf("execution environment name is required")
	}

	if len(sanitizeEnvName(executionEnvironmentID.Name)) == 0 {
		return fmt.Errorf("sanitizing execution environment name '%s' results in an invalid k8s pod name '%s', it must adhere to RFC 1123.",
			executionEnvironmentID.Name, sanitizeEnvName(executionEnvironmentID.Name))
	}

	if len(executionEnvironmentID.Version) == 0 {
		return fmt.Errorf("execution environment version is required")
	}

	if fastTaskEnvironmentSpec.GetBacklogLength() < 0 {
		return fmt.Errorf("backlog length must be greater than or equal to 0")
	}

	if fastTaskEnvironmentSpec.GetParallelism() <= 0 {
		return fmt.Errorf("parallelism must be greater than 0")
	}

	// currently the only supported termination criteria is ttlSeconds. if more are added, this
	// logic will need to be updated. we expect either `nil` termination criteria or a non-zero
	// ttlSeconds.
	if fastTaskEnvironmentSpec.GetTerminationCriteria() != nil && fastTaskEnvironmentSpec.GetTtlSeconds() == 0 {
		return fmt.Errorf("ttlSeconds must be greater than 0 if terminationCriteria is set")
	}

	podTemplateSpec := &v1.PodTemplateSpec{}
	if err := json.Unmarshal(fastTaskEnvironmentSpec.GetPodTemplateSpec(), podTemplateSpec); err != nil {
		return fmt.Errorf("unable to unmarshal PodTemplateSpec [%v], Err: [%v]", fastTaskEnvironmentSpec.GetPodTemplateSpec(), err.Error())
	}

	if fastTaskEnvironmentSpec.GetReplicaCount() <= 0 {
		return fmt.Errorf("replica count must be greater than 0")
	}

	return nil
}

// getTTLOrDefault returns the TTL for the given FastTaskEnvironmentSpec. If the termination criteria is not
// set, the default TTL is returned.
func getTTLOrDefault(fastTaskEnvironmentSpec *pb.FastTaskEnvironmentSpec) time.Duration {
	if fastTaskEnvironmentSpec.GetTerminationCriteria() == nil {
		return GetConfig().DefaultTTL.Duration
	}

	return time.Second * time.Duration(fastTaskEnvironmentSpec.GetTtlSeconds())
}

func isPodNotFoundErr(err error) bool {
	return k8serrors.IsNotFound(err) || k8serrors.IsGone(err)
}

// parseExecutionEnvID builds an `ExecutionEnvID` from the k8s labels provided.
func parseExectionEnvID(labels map[string]string) (core.ExecutionEnvID, error) {
	project, exists := labels[PROJECT_LABEL]
	if !exists {
		return core.ExecutionEnvID{}, fmt.Errorf("project label not found")
	}

	domain, exists := labels[DOMAIN_LABEL]
	if !exists {
		return core.ExecutionEnvID{}, fmt.Errorf("domain label not found")
	}

	name, exists := labels[EXECUTION_ENV_NAME]
	if !exists {
		return core.ExecutionEnvID{}, fmt.Errorf("execution environment name label not found")
	}

	version, exists := labels[EXECUTION_ENV_VERSION]
	if !exists {
		return core.ExecutionEnvID{}, fmt.Errorf("execution environment version label not found")
	}

	return core.ExecutionEnvID{
		Org:     labels[ORGANIZATION_LABEL],
		Project: project,
		Domain:  domain,
		Name:    name,
		Version: version,
	}, nil
}

// sanitizeEnvName sanitizes the environment name to be a valid k8s pod name. This means it converts
// the name to lowercase, replaces all underscores with dashes, removes all non-alphanumeric or dash
// characters, strips leading dashes, and truncates the name to be 63 characters with the appended
// nonce value.
func sanitizeEnvName(name string) string {
	// convert to lowercase
	lowerName := strings.ToLower(name)

	// replace all underscores with dashes
	dashedName := strings.ReplaceAll(lowerName, "_", "-")

	// remove all non-alphanumeric or dash characters
	cleanedName := nonLowerAlphanumericDashOrDotRegex.ReplaceAllString(dashedName, "")

	// strip leading dashes or periods
	for strings.HasPrefix(cleanedName, "-") || strings.HasPrefix(cleanedName, ".") {
		cleanedName = cleanedName[1:]
	}

	// return the first `N` characters where `N` is 63 characters minus the nonce length and dash
	return cleanedName[:min(len(cleanedName), 63-(GetConfig().NonceLength+1))]
}
