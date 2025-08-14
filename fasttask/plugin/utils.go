package plugin

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"google.golang.org/protobuf/types/known/structpb"
	v1 "k8s.io/api/core/v1"

	idlcore "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/utils"

	"github.com/unionai/flyte/fasttask/plugin/interfaces"
	"github.com/unionai/flyte/fasttask/plugin/pb"
)

var nonLowerAlphanumericDashOrDotRegex = regexp.MustCompile(`[^a-z0-9\.\-]+`)

// buildExecutionEnvID creates an `ExecutionEnvID` from a task ID and an `ExecutionEnv`. This
// collection of attributes is used to uniquely identify an execution environment.
func buildExecutionEnvID(tCtx core.TaskExecutionContext, executionEnv *idlcore.ExecutionEnv) interfaces.ExecutionEnvID {
	taskExecutionID := tCtx.TaskExecutionMetadata().GetTaskExecutionID().GetID()
	labels := tCtx.TaskExecutionMetadata().GetLabels()

	// we attempt to retrieve this metadata from the injected labels before the taskExecutionID
	// because the fasttask plugin uses Pod labels to detect orphaned environments. in scenarios where
	// the execution metadata differs from the task identifier metadata (ex. reference launchplans to
	// a different project / domain) the can result in mistakenly identifying an orphaned environment
	// which leads to prematurely deleting fasttask replicas.
	organization := getLabelIfExists(labels, ORGANIZATION_LABEL, taskExecutionID.GetTaskId().GetOrg())
	project := getLabelIfExists(labels, PROJECT_LABEL, taskExecutionID.GetTaskId().GetProject())
	domain := getLabelIfExists(labels, DOMAIN_LABEL, taskExecutionID.GetTaskId().GetDomain())

	return interfaces.ExecutionEnvID{
		Org:     organization,
		Project: project,
		Domain:  domain,
		Name:    executionEnv.GetName(),
		Version: executionEnv.GetVersion(),
	}
}

func buildTaskInfo(executionEnvID interfaces.ExecutionEnvID, executionEnv *idlcore.ExecutionEnv, workerID string) (*core.TaskInfo, error) {
	assignmentInfo := &pb.FastTaskAssignment{
		EnvironmentOrg:     executionEnvID.Org,
		EnvironmentProject: executionEnvID.Project,
		EnvironmentDomain:  executionEnvID.Domain,
		EnvironmentName:    executionEnvID.Name,
		EnvironmentVersion: executionEnvID.Version,
		AssignedWorker:     workerID,
	}
	customInfo := structpb.Struct{}
	err := utils.MarshalStruct(assignmentInfo, &customInfo)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	taskInfo := core.TaskInfo{
		OccurredAt: &now,
		ExternalResources: []*core.ExternalResource{
			{CustomInfo: &customInfo},
		},
	}

	return &taskInfo, nil
}

// getLabelIfExists attempts to return the value from provided label, if it does not exist it
// returns the provided default value
func getLabelIfExists(labels map[string]string, label, value string) string {
	if value, exists := labels[label]; exists {
		return value
	}

	return value
}

// isValidEnvironmentSpec validates the FastTaskEnvironmentSpec
func isValidEnvironmentSpec(executionEnvironmentID interfaces.ExecutionEnvID, fastTaskEnvironmentSpec *pb.FastTaskEnvironmentSpec) error {
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

// getEnvironmentTTLOrDefault returns the TTL for the given FastTaskEnvironmentSpec. If the termination criteria is not
// set, the default TTL is returned.
func getEnvironmentTTLOrDefault(fastTaskEnvironmentSpec *pb.FastTaskEnvironmentSpec) float64 {
	if fastTaskEnvironmentSpec == nil || fastTaskEnvironmentSpec.GetTerminationCriteria() == nil {
		return GetConfig().DefaultEnvironmentTTL.Duration.Seconds()
	}

	return float64(fastTaskEnvironmentSpec.GetTtlSeconds())
}

// getReplicaTTLOrDefault returns the TTL for the given FastTaskEnvironmentSpec.
// If not set, then the default replica TTL is returned.
func getReplicaTTLOrDefault(fastTaskEnvironmentSpec *pb.FastTaskEnvironmentSpec) float64 {
	if fastTaskEnvironmentSpec.GetScaledownTtlSeconds() == nil {
		return GetConfig().DefaultWorkerTTL.Duration.Seconds()
	}

	return float64(fastTaskEnvironmentSpec.GetScaledownTtlSeconds().GetValue())
}

// parseExecutionEnvID builds an `ExecutionEnvID` from the k8s labels provided.
func parseExectionEnvID(labels map[string]string) (interfaces.ExecutionEnvID, error) {
	project, exists := labels[PROJECT_LABEL]
	if !exists {
		return interfaces.ExecutionEnvID{}, fmt.Errorf("project label not found")
	}

	domain, exists := labels[DOMAIN_LABEL]
	if !exists {
		return interfaces.ExecutionEnvID{}, fmt.Errorf("domain label not found")
	}

	name, exists := labels[EXECUTION_ENV_NAME]
	if !exists {
		return interfaces.ExecutionEnvID{}, fmt.Errorf("execution environment name label not found")
	}

	version, exists := labels[EXECUTION_ENV_VERSION]
	if !exists {
		return interfaces.ExecutionEnvID{}, fmt.Errorf("execution environment version label not found")
	}

	return interfaces.ExecutionEnvID{
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

func getMinReplicaCount(fastTaskEnvironmentSpec *pb.FastTaskEnvironmentSpec) int {
	if fastTaskEnvironmentSpec.GetMinReplicaCount() != nil && fastTaskEnvironmentSpec.GetMinReplicaCount().GetValue() > 0 {
		return int(fastTaskEnvironmentSpec.GetMinReplicaCount().GetValue())
	}

	return int(fastTaskEnvironmentSpec.GetReplicaCount())
}

func hashMapValues(m map[string]string) string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var sb strings.Builder
	for _, k := range keys {
		sb.WriteString(k)
		sb.WriteString("=")
		sb.WriteString(m[k])
		sb.WriteString(";")
	}
	return sb.String()
}
