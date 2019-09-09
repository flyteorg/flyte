package transformer

import (
	"context"
	"fmt"
	"reflect"

	"encoding/base64"

	datacatalog "github.com/lyft/datacatalog/protos/gen"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
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

func getDatasetNameFromTask(task *core.TaskTemplate) string {
	return fmt.Sprintf("%s-%s", taskNamespace, task.Id.Name)
}

func GenerateTaskOutputsFromArtifact(task *core.TaskTemplate, artifact *datacatalog.Artifact) (*core.LiteralMap, error) {
	// if there are no outputs in the task, return empty map
	if task.Interface.Outputs == nil || len(task.Interface.Outputs.Variables) == 0 {
		return &emptyLiteralMap, nil
	}

	outputVariables := task.Interface.Outputs.Variables
	artifactDataList := artifact.Data

	// verify the task outputs matches what is stored in ArtifactData
	if len(outputVariables) != len(artifactDataList) {
		return nil, fmt.Errorf("The task %v with %d outputs, should have %d artifactData for artifact %v",
			task.Id, len(outputVariables), len(artifactDataList), artifact.Id)
	}

	outputs := make(map[string]*core.Literal, len(artifactDataList))
	for _, artifactData := range artifactDataList {
		// verify that the name and type of artifactData matches what is expected from the interface
		if _, ok := outputVariables[artifactData.Name]; !ok {
			return nil, fmt.Errorf("Unexpected artifactData with name [%v] does not match any task output variables %v", artifactData.Name, reflect.ValueOf(outputVariables).MapKeys())
		}

		expectedVarType := outputVariables[artifactData.Name].GetType()
		inputType := validators.LiteralTypeForLiteral(artifactData.Value)
		if !validators.AreTypesCastable(inputType, expectedVarType) {
			return nil, fmt.Errorf("Unexpected artifactData: [%v] type: [%v] does not match any task output type: [%v]", artifactData.Name, inputType, expectedVarType)
		}

		outputs[artifactData.Name] = artifactData.Value
	}

	return &core.LiteralMap{Literals: outputs}, nil
}

func generateDataSetVersionFromTask(ctx context.Context, task *core.TaskTemplate) (string, error) {
	signatureHash, err := generateTaskSignatureHash(ctx, task)
	if err != nil {
		return "", err
	}

	cacheVersion := task.Metadata.DiscoveryVersion
	if len(cacheVersion) == 0 {
		return "", fmt.Errorf("Task cannot have an empty discoveryVersion %v", cacheVersion)
	}
	return fmt.Sprintf("%s-%s", cacheVersion, signatureHash), nil
}

func generateTaskSignatureHash(ctx context.Context, task *core.TaskTemplate) (string, error) {
	taskInputs := &emptyVariableMap
	taskOutputs := &emptyVariableMap

	if task.Interface.Inputs != nil && len(task.Interface.Inputs.Variables) != 0 {
		taskInputs = task.Interface.Inputs
	}

	if task.Interface.Outputs != nil && len(task.Interface.Outputs.Variables) != 0 {
		taskOutputs = task.Interface.Outputs
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
func GenerateDatasetIDForTask(ctx context.Context, task *core.TaskTemplate) (*datacatalog.DatasetID, error) {
	datasetVersion, err := generateDataSetVersionFromTask(ctx, task)
	if err != nil {
		return nil, err
	}

	datasetID := &datacatalog.DatasetID{
		Project: task.Id.Project,
		Domain:  task.Id.Domain,
		Name:    getDatasetNameFromTask(task),
		Version: datasetVersion,
	}
	return datasetID, nil
}
