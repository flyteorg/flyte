package datacatalog

import (
	"context"
	"encoding/base64"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/datacatalog"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/catalog"
	"github.com/flyteorg/flytepropeller/pkg/compiler/validators"
	"github.com/flyteorg/flytestdlib/pbhash"
)

const cachedTaskTag = "flyte_cached"
const taskNamespace = "flyte_task"
const maxParamHashLength = 8

// Declare the definition of empty literal and variable maps. This is important because we hash against
// the literal and variable maps. So Nil and empty literals and variable maps should translate to these defintions
// in order to have a consistent hash.
var emptyLiteralMap = core.LiteralMap{Literals: map[string]*core.Literal{}}
var emptyVariableMap = core.VariableMap{Variables: map[string]*core.Variable{}}

func getDatasetNameFromTask(taskID core.Identifier) string {
	return fmt.Sprintf("%s-%s", taskNamespace, taskID.Name)
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
		if !validators.AreTypesCastable(inputType, expectedVarType) {
			return nil, fmt.Errorf("unexpected artifactData: [%v] type: [%v] does not match any task output type: [%v]", artifactData.Name, inputType, expectedVarType)
		}

		outputs[artifactData.Name] = artifactData.Value
	}

	return &core.LiteralMap{Literals: outputs}, nil
}

func generateDataSetVersionFromTask(ctx context.Context, taskInterface core.TypedInterface, cacheVersion string) (string, error) {
	signatureHash, err := generateTaskSignatureHash(ctx, taskInterface)
	if err != nil {
		return "", err
	}

	cacheVersion = strings.Trim(cacheVersion, " ")
	if len(cacheVersion) == 0 {
		return "", fmt.Errorf("task cannot have an empty discoveryVersion %v", cacheVersion)
	}

	return fmt.Sprintf("%s-%s", cacheVersion, signatureHash), nil
}

func generateTaskSignatureHash(ctx context.Context, taskInterface core.TypedInterface) (string, error) {
	taskInputs := &emptyVariableMap
	taskOutputs := &emptyVariableMap

	if taskInterface.Inputs != nil && len(taskInterface.Inputs.Variables) != 0 {
		taskInputs = taskInterface.Inputs
	}

	if taskInterface.Outputs != nil && len(taskInterface.Outputs.Variables) != 0 {
		taskOutputs = taskInterface.Outputs
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

// Generate a tag by hashing the input values
func GenerateArtifactTagName(ctx context.Context, inputs *core.LiteralMap) (string, error) {
	hashString, err := catalog.HashLiteralMap(ctx, inputs)
	if err != nil {
		return "", err
	}
	tag := fmt.Sprintf("%s-%s", cachedTaskTag, hashString)
	return tag, nil
}

// Get the DataSetID for a task.
// NOTE: the version of the task is a combination of both the discoverable_version and the task signature.
// This is because the interface may of changed even if the discoverable_version hadn't.
func GenerateDatasetIDForTask(ctx context.Context, k catalog.Key) (*datacatalog.DatasetID, error) {
	datasetVersion, err := generateDataSetVersionFromTask(ctx, k.TypedInterface, k.CacheVersion)
	if err != nil {
		return nil, err
	}

	datasetID := &datacatalog.DatasetID{
		Project: k.Identifier.Project,
		Domain:  k.Identifier.Domain,
		Name:    getDatasetNameFromTask(k.Identifier),
		Version: datasetVersion,
	}
	return datasetID, nil
}

func DatasetIDToIdentifier(id *datacatalog.DatasetID) *core.Identifier {
	if id == nil {
		return nil
	}
	return &core.Identifier{ResourceType: core.ResourceType_DATASET, Name: id.Name, Project: id.Project, Domain: id.Domain, Version: id.Version}
}

// With Node-Node relationship this is bound to change. So lets keep it extensible
const (
	taskVersionKey     = "task-version"
	execNameKey        = "execution-name"
	execDomainKey      = "exec-domain"
	execProjectKey     = "exec-project"
	execNodeIDKey      = "exec-node"
	execTaskAttemptKey = "exec-attempt"
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
			taskVersionKey: taskExecutionID.TaskId.Version,
		},
	}
}

func GetArtifactMetadataForSource(taskExecutionID *core.TaskExecutionIdentifier) *datacatalog.Metadata {
	if taskExecutionID == nil {
		return &datacatalog.Metadata{}
	}
	return &datacatalog.Metadata{
		KeyMap: map[string]string{
			execProjectKey:     taskExecutionID.NodeExecutionId.GetExecutionId().GetProject(),
			execDomainKey:      taskExecutionID.NodeExecutionId.GetExecutionId().GetDomain(),
			execNameKey:        taskExecutionID.NodeExecutionId.GetExecutionId().GetName(),
			execNodeIDKey:      taskExecutionID.NodeExecutionId.GetNodeId(),
			execTaskAttemptKey: strconv.Itoa(int(taskExecutionID.GetRetryAttempt())),
		},
	}
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

	// Jul-06-2020 DataCatalog stores only wfExecutionKey & taskVersionKey So we will default the project / domain to the current dataset's project domain
	val := GetOrDefault(artifactMd.KeyMap, execTaskAttemptKey, "0")
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
			Version:      GetOrDefault(datasetMd.KeyMap, taskVersionKey, "unknown"),
		},
		RetryAttempt: uint32(attempt),
		NodeExecutionId: &core.NodeExecutionIdentifier{
			NodeId: GetOrDefault(artifactMd.KeyMap, execNodeIDKey, "unknown"),
			ExecutionId: &core.WorkflowExecutionIdentifier{
				Project: GetOrDefault(artifactMd.KeyMap, execProjectKey, currentID.GetProject()),
				Domain:  GetOrDefault(artifactMd.KeyMap, execDomainKey, currentID.GetDomain()),
				Name:    GetOrDefault(artifactMd.KeyMap, execNameKey, "unknown"),
			},
		},
	}, nil
}

// Given the Catalog Information (returned from a Catalog call), returns the CatalogMetadata that is populated in the event.
func EventCatalogMetadata(datasetID *datacatalog.DatasetID, tag *datacatalog.Tag, sourceID *core.TaskExecutionIdentifier) *core.CatalogMetadata {
	md := &core.CatalogMetadata{
		DatasetId: DatasetIDToIdentifier(datasetID),
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
