package datacatalog

import (
	"context"
	"encoding/base64"
	"fmt"
	"reflect"
	"strings"

	datacatalog "github.com/lyft/datacatalog/protos/gen"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/lyft/flyteplugins/go/tasks/pluginmachinery/catalog"

	"github.com/lyft/flytepropeller/pkg/compiler/validators"

	"github.com/lyft/flytestdlib/pbhash"
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
	if inputs == nil || len(inputs.Literals) == 0 {
		inputs = &emptyLiteralMap
	}

	inputsHash, err := pbhash.ComputeHash(ctx, inputs)
	if err != nil {
		return "", err
	}

	hashString := base64.RawURLEncoding.EncodeToString(inputsHash)
	tag := fmt.Sprintf("%s-%s", cachedTaskTag, hashString)
	return tag, nil
}

// Get the DataSetID for a task.
// NOTE: the version of the task is a combination of both the discoverable_version and the task signature.
// This is because the interfact may of changed even if the discoverable_version hadn't.
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
