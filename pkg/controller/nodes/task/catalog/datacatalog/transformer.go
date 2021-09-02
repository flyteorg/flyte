// Code generated

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
	"github.com/golang/protobuf/proto"
)

type LegacyVariableMap struct {
	// Legacy VariableMap copied from flyteidl@v0.19.19. It's used in place of updated VariableMap to generate dataset version hash to avoid cache bust.
	// Defines a map of variable names to variables.
	Variables            map[string]*core.Variable `protobuf:"bytes,1,rep,name=variables,proto3" json:"variables,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	XXX_NoUnkeyedLiteral struct{}                  `json:"-"`
	XXX_unrecognized     []byte                    `json:"-"`
	XXX_sizecache        int32                     `json:"-"`
}

func (m *LegacyVariableMap) Reset()         { *m = LegacyVariableMap{} }
func (m *LegacyVariableMap) String() string { return proto.CompactTextString(m) }
func (*LegacyVariableMap) ProtoMessage()    {}
func (*LegacyVariableMap) Descriptor() ([]byte, []int) {
	return fileDescriptor_cd7be6cbe85c3def, []int{1}
}

var fileDescriptor_cd7be6cbe85c3def = []byte{
	// 425 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x84, 0x53, 0x5d, 0x6b, 0xd4, 0x40,
	0x14, 0xed, 0xec, 0x6a, 0xbb, 0x7b, 0xa3, 0x55, 0xe6, 0x41, 0x63, 0xa8, 0x10, 0xf2, 0xb4, 0x45,
	0x9a, 0x40, 0x2c, 0x22, 0x3e, 0x16, 0xc4, 0x8a, 0x0a, 0x32, 0xf8, 0x01, 0xe2, 0xcb, 0x24, 0xb9,
	0x9b, 0x0e, 0xa6, 0x99, 0x71, 0x32, 0x09, 0xc4, 0xdf, 0xe1, 0xdf, 0xf0, 0xcd, 0x1f, 0x28, 0xc9,
	0x26, 0x69, 0xb2, 0x34, 0xf8, 0x36, 0x77, 0xce, 0x39, 0x39, 0x27, 0x73, 0xb8, 0xf0, 0x74, 0x9b,
	0xd5, 0x06, 0x45, 0x92, 0x05, 0xb1, 0xd4, 0x18, 0x88, 0xdc, 0xa0, 0xde, 0xf2, 0x18, 0x7d, 0xa5,
	0xa5, 0x91, 0xf4, 0x7e, 0x0f, 0xfb, 0x0d, 0xec, 0x3c, 0x99, 0xb2, 0x4d, 0xad, 0xb0, 0xd8, 0x31,
	0x9d, 0x93, 0x29, 0x94, 0x09, 0x83, 0x9a, 0x67, 0x1d, 0xea, 0x7d, 0x87, 0xd5, 0x17, 0xae, 0x05,
	0x8f, 0x32, 0xa4, 0x3e, 0xdc, 0x69, 0x84, 0x36, 0x71, 0xc9, 0xc6, 0x0a, 0x1d, 0x7f, 0x62, 0xe1,
	0xbf, 0xdf, 0x09, 0x3f, 0xd5, 0x0a, 0x59, 0xcb, 0xa3, 0x2e, 0x58, 0x09, 0x16, 0xb1, 0x16, 0xca,
	0x08, 0x99, 0xdb, 0x0b, 0x97, 0x6c, 0xd6, 0x6c, 0x7c, 0xe5, 0xfd, 0x21, 0x60, 0xf5, 0x9f, 0xff,
	0xc0, 0x15, 0x7d, 0x03, 0xeb, 0xaa, 0x1b, 0x0b, 0x9b, 0xb8, 0xcb, 0x8d, 0x15, 0x9e, 0xee, 0xd9,
	0x8c, 0xe8, 0xc3, 0xb9, 0x78, 0x9d, 0x1b, 0x5d, 0xb3, 0x1b, 0xad, 0xf3, 0x19, 0x8e, 0xa7, 0x20,
	0x7d, 0x08, 0xcb, 0x1f, 0x58, 0xb7, 0xd9, 0xd7, 0xac, 0x39, 0xd2, 0x33, 0xb8, 0x5b, 0xf1, 0xac,
	0xc4, 0x36, 0x98, 0x15, 0x3e, 0x9e, 0x31, 0x62, 0x3b, 0xd6, 0xab, 0xc5, 0x4b, 0xe2, 0xfd, 0x82,
	0xe3, 0xe6, 0xff, 0x92, 0xb7, 0xfd, 0x6b, 0xd3, 0x10, 0x0e, 0x45, 0xae, 0x4a, 0x53, 0xcc, 0xbc,
	0xca, 0x28, 0x2e, 0xeb, 0x98, 0xf4, 0x1c, 0x8e, 0x64, 0x69, 0x5a, 0xd1, 0xe2, 0xbf, 0xa2, 0x9e,
	0xea, 0xfd, 0x26, 0xb0, 0xfe, 0xc8, 0x35, 0xbf, 0x46, 0x83, 0x9a, 0x9e, 0xc2, 0xb2, 0xe2, 0xba,
	0x33, 0x9d, 0x8d, 0xde, 0x70, 0x68, 0x08, 0x47, 0x09, 0x6e, 0x79, 0x99, 0x99, 0xce, 0xee, 0xd1,
	0xed, 0xcd, 0x5d, 0x1e, 0xb0, 0x9e, 0x48, 0x4f, 0x60, 0xa5, 0xf1, 0x67, 0x29, 0x34, 0x26, 0xf6,
	0xd2, 0x25, 0x9b, 0xd5, 0xe5, 0x01, 0x1b, 0x6e, 0x2e, 0x00, 0x56, 0x11, 0x5e, 0xf1, 0x4a, 0x48,
	0xed, 0xfd, 0x25, 0x70, 0x6f, 0x88, 0xd5, 0x74, 0xf8, 0x0e, 0x40, 0xf5, 0x73, 0x5f, 0xe2, 0xb3,
	0x3d, 0xc7, 0xb1, 0xe0, 0x66, 0xe8, 0x6a, 0x1c, 0xc9, 0x9d, 0xaf, 0xf0, 0x60, 0x0f, 0xbe, 0xa5,
	0x48, 0x7f, 0x5a, 0xa4, 0x3d, 0x67, 0x36, 0x6a, 0xf2, 0xe2, 0xc5, 0xb7, 0xf3, 0x54, 0x98, 0xab,
	0x32, 0xf2, 0x63, 0x79, 0x1d, 0xb4, 0x02, 0xa9, 0xd3, 0x60, 0xd8, 0x85, 0x14, 0xf3, 0x40, 0x45,
	0x67, 0xa9, 0x0c, 0x26, 0xeb, 0x11, 0x1d, 0xb6, 0x6b, 0xf1, 0xfc, 0x5f, 0x00, 0x00, 0x00, 0xff,
	0xff, 0x61, 0xf3, 0xc7, 0x07, 0x7f, 0x03, 0x00, 0x00,
}

const cachedTaskTag = "flyte_cached"
const taskNamespace = "flyte_task"
const maxParamHashLength = 8

// Declare the definition of empty literal and variable maps. This is important because we hash against
// the literal and variable maps. So Nil and empty literals and variable maps should translate to these defintions
// in order to have a consistent hash.
var emptyLiteralMap = core.LiteralMap{Literals: map[string]*core.Literal{}}

func getDatasetNameFromTask(taskID core.Identifier) string {
	return fmt.Sprintf("%s-%s", taskNamespace, taskID.Name)
}

// Transform the artifact Data into task execution outputs as a literal map
func GenerateTaskOutputsFromArtifact(id core.Identifier, taskInterface core.TypedInterface, artifact *datacatalog.Artifact) (*core.LiteralMap, error) {

	// if there are no outputs in the task, return empty map
	if taskInterface.Outputs == nil || len(taskInterface.Outputs.Variables) == 0 {
		return &emptyLiteralMap, nil
	}

	outputVariables := validators.VariableMapEntriesToMap(taskInterface.Outputs.Variables)
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
	taskInputs := &LegacyVariableMap{Variables: map[string]*core.Variable{}}
	taskOutputs := &LegacyVariableMap{Variables: map[string]*core.Variable{}}

	if taskInterface.Inputs != nil && len(taskInterface.Inputs.Variables) != 0 {
		taskInputs.Variables = validators.VariableMapEntriesToMap(taskInterface.Inputs.Variables)
	}

	if taskInterface.Outputs != nil && len(taskInterface.Outputs.Variables) != 0 {
		taskOutputs.Variables = validators.VariableMapEntriesToMap(taskInterface.Outputs.Variables)
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

// Returns the Source TaskExecutionIdentifier from the catalog metadata
// For all the information not available it returns Unknown. This is because as of July-2020 Catalog does not have all
// the information. After the first deployment of this code, it will have this and the "unknown's" can be phased out
func GetSourceFromMetadata(datasetMd, artifactMd *datacatalog.Metadata, currentID core.Identifier) *core.TaskExecutionIdentifier {
	if datasetMd == nil || datasetMd.KeyMap == nil {
		datasetMd = &datacatalog.Metadata{KeyMap: map[string]string{}}
	}
	if artifactMd == nil || artifactMd.KeyMap == nil {
		artifactMd = &datacatalog.Metadata{KeyMap: map[string]string{}}
	}
	// Jul-06-2020 DataCatalog stores only wfExecutionKey & taskVersionKey So we will default the project / domain to the current dataset's project domain
	attempt, _ := strconv.Atoi(GetOrDefault(artifactMd.KeyMap, execTaskAttemptKey, "0"))
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
	}
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
