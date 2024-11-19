package datacatalog

import (
	"context"
	"encoding/base64"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/datacatalog"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/catalog"
	"github.com/flyteorg/flyte/flytepropeller/pkg/compiler/validators"
	"github.com/flyteorg/flyte/flytestdlib/pbhash"
)

const (
	cachedTaskTag = "flyte_cached"

	launchplanNamespace = "flyte_lp"
	taskNamespace       = "flyte_task"
	workflowNamespace   = "flyte_wf"

	maxParamHashLength = 8
)

// Declare the definition of empty literal and variable maps. This is important because we hash against
// the literal and variable maps. So Nil and empty literals and variable maps should translate to these definitions
// in order to have a consistent hash.
var emptyLiteralMap = core.LiteralMap{Literals: map[string]*core.Literal{}}
var emptyVariableMap = core.VariableMap{Variables: map[string]*core.Variable{}}

func getDatasetName(id core.Identifier) (string, error) {
	var namespace string
	switch id.ResourceType {
	case core.ResourceType_LAUNCH_PLAN:
		namespace = launchplanNamespace
	case core.ResourceType_TASK:
		namespace = taskNamespace
	case core.ResourceType_WORKFLOW:
		namespace = workflowNamespace
	default:
		return "", fmt.Errorf("unsupported resource type [%v] for id [%v]", id.ResourceType, id)
	}
	return fmt.Sprintf("%s-%s", namespace, id.Name), nil
}

// Transform the artifact Data into task execution outputs as a literal map
func GenerateTaskOutputsFromArtifact(id core.Identifier, taskInterface core.TypedInterface, artifact *datacatalog.Artifact) (*core.LiteralMap, error) {

	// if there are no outputs in the task, return empty map
	if taskInterface.Outputs == nil || len(taskInterface.Outputs.Variables) == 0 {
		return &emptyLiteralMap, nil
	}

	outputVariables := taskInterface.Outputs.Variables
	artifactDataList := artifact.Data

	// verify the task outputs matches what is stored in ArtifactData
	if len(outputVariables) != len(artifactDataList) {
		return nil, fmt.Errorf("the task %s with %d outputs, should have %d artifactData for artifact %s", id.String(), len(outputVariables), len(artifactDataList), artifact.Id)
	}

	outputs := make(map[string]*core.Literal, len(artifactDataList))
	for _, artifactData := range artifactDataList {
		// verify that the name and type of artifactData matches what is expected from the interface
		if _, ok := outputVariables[artifactData.Name]; !ok {
			return nil, fmt.Errorf("unexpected artifactData with name [%v] does not match any task output variables %v", artifactData.Name, reflect.ValueOf(outputVariables).MapKeys())
		}

		expectedVarType := outputVariables[artifactData.Name].GetType()
		inputType := validators.LiteralTypeForLiteral(artifactData.Value)
		err := validators.ValidateLiteralType(inputType)
		if err != nil {
			return nil, fmt.Errorf("failed to validate literal type for %s with err: %s", artifactData.Name, err)
		}
		if !validators.AreTypesCastable(inputType, expectedVarType) {
			return nil, fmt.Errorf("unexpected artifactData: [%v] type: [%v] does not match any task output type: [%v]", artifactData.Name, inputType, expectedVarType)
		}

		outputs[artifactData.Name] = artifactData.Value
	}

	return &core.LiteralMap{Literals: outputs}, nil
}

func generateDataSetVersion(ctx context.Context, taskInterface core.TypedInterface, cacheVersion string) (string, error) {
	signatureHash, err := GenerateInterfaceSignatureHash(ctx, taskInterface)
	if err != nil {
		return "", err
	}

	cacheVersion = strings.Trim(cacheVersion, " ")
	if len(cacheVersion) == 0 {
		return "", fmt.Errorf("task cannot have an empty discoveryVersion %v", cacheVersion)
	}

	return fmt.Sprintf("%s-%s", cacheVersion, signatureHash), nil
}

func GenerateInterfaceSignatureHash(ctx context.Context, resourceInterface core.TypedInterface) (string, error) {
	taskInputs := &emptyVariableMap
	taskOutputs := &emptyVariableMap

	if resourceInterface.Inputs != nil && len(resourceInterface.Inputs.Variables) != 0 {
		taskInputs = resourceInterface.Inputs
	}

	if resourceInterface.Outputs != nil && len(resourceInterface.Outputs.Variables) != 0 {
		taskOutputs = resourceInterface.Outputs
	}

	inputHash, err := pbhash.ComputeHash(ctx, taskInputs)
	if err != nil {
		return "", err
	}

	outputHash, err := pbhash.ComputeHash(ctx, taskOutputs)
	if err != nil {
		return "", err
	}

	inputHashString := base64.RawURLEncoding.EncodeToString(inputHash)

	if len(inputHashString) > maxParamHashLength {
		inputHashString = inputHashString[0:maxParamHashLength]
	}

	outputHashString := base64.RawURLEncoding.EncodeToString(outputHash)
	if len(outputHashString) > maxParamHashLength {
		outputHashString = outputHashString[0:maxParamHashLength]
	}

	return fmt.Sprintf("%v-%v", inputHashString, outputHashString), nil
}

// Generate a tag by hashing the input values which are not in cacheIgnoreInputVars
func GenerateArtifactTagName(ctx context.Context, inputs *core.LiteralMap, cacheIgnoreInputVars []string) (string, error) {
	hashString, err := catalog.HashLiteralMap(ctx, inputs, cacheIgnoreInputVars)
	if err != nil {
		return "", err
	}
	tag := fmt.Sprintf("%s-%s", cachedTaskTag, hashString)
	return tag, nil
}

// Get the DataSetID for a task.
// NOTE: the version of the task is a combination of both the discoverable_version and the task signature.
// This is because the interface may of changed even if the discoverable_version hadn't.
func GenerateDatasetID(ctx context.Context, k catalog.Key) (*datacatalog.DatasetID, error) {
	datasetVersion, err := generateDataSetVersion(ctx, k.TypedInterface, k.CacheVersion)
	if err != nil {
		return nil, err
	}

	name, err := getDatasetName(k.Identifier)
	if err != nil {
		return nil, err
	}

	datasetID := &datacatalog.DatasetID{
		Project: k.Identifier.Project,
		Domain:  k.Identifier.Domain,
		Name:    name,
		Version: datasetVersion,
		Org:     k.Identifier.Org,
	}
	return datasetID, nil
}

func DatasetIDToIdentifier(id *datacatalog.DatasetID) *core.Identifier {
	if id == nil {
		return nil
	}
	return &core.Identifier{ResourceType: core.ResourceType_DATASET, Name: id.Name, Project: id.Project, Domain: id.Domain, Version: id.Version, Org: id.Org}
}

// With Node-Node relationship this is bound to change. So lets keep it extensible
const (
	TaskVersionKey     = "task-version"
	ExecNameKey        = "execution-name"
	ExecDomainKey      = "exec-domain"
	ExecProjectKey     = "exec-project"
	ExecNodeIDKey      = "exec-node"
	ExecTaskAttemptKey = "exec-attempt"
	ExecOrgKey         = "exec-rog"
)

// Understanding Catalog Identifiers
// DatasetID represents the ID of the dataset. For Flyte this represents the ID of the generating task and the version calculated as the hash of the interface & cache version. refer to `GenerateDatasetIDForTask`
// TaskID is the same as the DatasetID + name: (DataSetID - namespace) + task version which is stored in the metadata
// ExecutionID is stored only in the metadata (project and domain available after Jul-2020)
// NodeExecID = Execution ID + Node ID (available after Jul-2020)
// TaskExecID is the same as the NodeExecutionID + attempt (attempt is available in Metadata) after Jul-2020
func GetDatasetMetadataForSource(taskExecutionID *core.TaskExecutionIdentifier) *datacatalog.Metadata {
	if taskExecutionID == nil {
		return &datacatalog.Metadata{}
	}
	return &datacatalog.Metadata{
		KeyMap: map[string]string{
			TaskVersionKey: taskExecutionID.TaskId.Version,
		},
	}
}

func GetArtifactMetadataForSource(metadata catalog.Metadata) *datacatalog.Metadata {
	if metadata.TaskExecutionIdentifier != nil {
		return &datacatalog.Metadata{
			KeyMap: map[string]string{
				ExecProjectKey:     metadata.TaskExecutionIdentifier.NodeExecutionId.GetExecutionId().GetProject(),
				ExecDomainKey:      metadata.TaskExecutionIdentifier.NodeExecutionId.GetExecutionId().GetDomain(),
				ExecNameKey:        metadata.TaskExecutionIdentifier.NodeExecutionId.GetExecutionId().GetName(),
				ExecNodeIDKey:      metadata.TaskExecutionIdentifier.NodeExecutionId.GetNodeId(),
				ExecTaskAttemptKey: strconv.Itoa(int(metadata.TaskExecutionIdentifier.GetRetryAttempt())),
				ExecOrgKey:         metadata.TaskExecutionIdentifier.GetNodeExecutionId().GetExecutionId().GetOrg(),
			},
		}
	} else if metadata.NodeExecutionIdentifier != nil {
		return &datacatalog.Metadata{
			KeyMap: map[string]string{
				ExecProjectKey: metadata.NodeExecutionIdentifier.GetExecutionId().GetProject(),
				ExecDomainKey:  metadata.NodeExecutionIdentifier.GetExecutionId().GetDomain(),
				ExecNameKey:    metadata.NodeExecutionIdentifier.GetExecutionId().GetName(),
				ExecNodeIDKey:  metadata.NodeExecutionIdentifier.GetNodeId(),
				ExecOrgKey:     metadata.NodeExecutionIdentifier.GetExecutionId().GetOrg(),
			},
		}
	}
	return &datacatalog.Metadata{}
}

// GetSourceFromMetadata returns the Source TaskExecutionIdentifier from the catalog metadata
// For all the information not available it returns Unknown. This is because as of July-2020 Catalog does not have all
// the information. After the first deployment of this code, it will have this and the "unknown's" can be phased out
func GetSourceFromMetadata(datasetMd, artifactMd *datacatalog.Metadata, currentID core.Identifier) (*core.TaskExecutionIdentifier, error) {
	if datasetMd == nil || datasetMd.KeyMap == nil {
		datasetMd = &datacatalog.Metadata{KeyMap: map[string]string{}}
	}
	if artifactMd == nil || artifactMd.KeyMap == nil {
		artifactMd = &datacatalog.Metadata{KeyMap: map[string]string{}}
	}

	// Jul-06-2020 DataCatalog stores only wfExecutionKey & TaskVersionKey So we will default the project / domain to the current dataset's project domain
	val := GetOrDefault(artifactMd.KeyMap, ExecTaskAttemptKey, "0")
	attempt, err := strconv.ParseUint(val, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("failed to parse [%v] to integer. Error: %w", val, err)
	}

	return &core.TaskExecutionIdentifier{
		TaskId: &core.Identifier{
			ResourceType: currentID.ResourceType,
			Project:      currentID.Project,
			Domain:       currentID.Domain,
			Name:         currentID.Name,
			Version:      GetOrDefault(datasetMd.KeyMap, TaskVersionKey, "unknown"),
			Org:          currentID.Org,
		},
		RetryAttempt: uint32(attempt),
		NodeExecutionId: &core.NodeExecutionIdentifier{
			NodeId: GetOrDefault(artifactMd.KeyMap, ExecNodeIDKey, "unknown"),
			ExecutionId: &core.WorkflowExecutionIdentifier{
				Project: GetOrDefault(artifactMd.KeyMap, ExecProjectKey, currentID.GetProject()),
				Domain:  GetOrDefault(artifactMd.KeyMap, ExecDomainKey, currentID.GetDomain()),
				Name:    GetOrDefault(artifactMd.KeyMap, ExecNameKey, "unknown"),
				Org:     GetOrDefault(artifactMd.KeyMap, ExecOrgKey, currentID.GetOrg()),
			},
		},
	}, nil
}

// Given the Catalog Information (returned from a Catalog call), returns the CatalogMetadata that is populated in the event.
func EventCatalogMetadata(datasetID *datacatalog.DatasetID, tag *datacatalog.Tag, sourceID *core.TaskExecutionIdentifier, createdAt *timestamppb.Timestamp) *core.CatalogMetadata {
	md := &core.CatalogMetadata{
		DatasetId: DatasetIDToIdentifier(datasetID),
		CreatedAt: createdAt,
	}

	if tag != nil {
		md.ArtifactTag = &core.CatalogArtifactTag{
			ArtifactId: tag.ArtifactId,
			Name:       tag.Name,
		}
	}

	if sourceID != nil {
		md.SourceExecution = &core.CatalogMetadata_SourceTaskExecution{
			SourceTaskExecution: sourceID,
		}
	}

	return md
}

// Returns a default value, if the given key is not found in the map, else returns the value in the map
func GetOrDefault(m map[string]string, key, defaultValue string) string {
	v, ok := m[key]
	if !ok {
		return defaultValue
	}
	return v
}
