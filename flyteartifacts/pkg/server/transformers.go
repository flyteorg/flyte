package server

import (
	"fmt"
	"github.com/flyteorg/flyte/flyteartifacts/pkg/models"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/artifact"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
)

// CreateArtifactModelFrom is used for converting from an incoming request to a model
func CreateArtifactModelFrom(key *core.ArtifactKey, spec *artifact.ArtifactSpec, version string, partitions map[string]string, tag string, principal string) (models.Artifact, error) {

	if key == nil || spec == nil {
		return models.Artifact{}, fmt.Errorf("key and spec cannot be nil")
	}
	if len(version) == 0 {
		return models.Artifact{}, fmt.Errorf("version cannot be empty")
	}
	if spec.Type == nil || spec.Value == nil {
		return models.Artifact{}, fmt.Errorf("spec type and value cannot be nil")
	}

	ex := spec.Execution
	if ex == nil {
		return models.Artifact{}, fmt.Errorf("spec execution cannot be nil")
	}
	if ex.Project != key.Project || ex.Domain != key.Domain || ex.Name != key.Name {
		return models.Artifact{}, fmt.Errorf("spec execution must match key")
	}

	// gatepr: Offload metadata
	a := models.Artifact{
		Principal:             principal,
		Project:               key.Project,
		Domain:                key.Domain,
		Name:                  key.Name,
		Version:               version,
		Partitions:            partitions,
		LiteralType:           *spec.Type,
		LiteralValue:          *spec.Value,
		Description:           spec.ShortDescription,
		MetadataType:          spec.MetadataType,
		OffloadedUserMetadata: "",
		UserMetadata:          nil,
		ExecutionName:         spec.Execution.Name,
	}

	// If the artifact was made by a Task output, then this should be filled in.
	if spec.TaskExecution != nil {
		taskID := spec.TaskExecution.TaskId
		nodeExecID := spec.TaskExecution.NodeExecutionId
		a.TaskProject = taskID.Project
		a.TaskDomain = taskID.Domain
		a.TaskName = taskID.Name
		a.TaskVersion = taskID.Version
		a.NodeID = nodeExecID.NodeId
		a.RetryAttempt = spec.TaskExecution.RetryAttempt
	}
	// gatepr: Add workflow info
	return a, nil
}

func PartitionsToIdl(partitions map[string]string) *core.Partitions {
	if partitions == nil || len(partitions) == 0 {
		return nil
	}

	cp := core.Partitions{
		Value: make(map[string]*core.LabelValue),
	}

	for k, v := range partitions {
		cp.Value[k] = &core.LabelValue{
			Value: &core.LabelValue_StaticValue{
				StaticValue: v,
			},
		}
	}

	return &cp
}

// FromModelToIdl translates from the model to the IDL
func FromModelToIdl(model models.Artifact) artifact.Artifact {
	// gatepr: think about un-offloading

	a := artifact.Artifact{
		ArtifactId: &core.ArtifactID{
			ArtifactKey: &core.ArtifactKey{
				Project: model.Project,
				Domain:  model.Domain,
				Name:    model.Name,
			},
			Version: model.Version,
		},
		Spec: &artifact.ArtifactSpec{
			Type:  &model.LiteralType,
			Value: &model.LiteralValue,

			TaskExecution: &core.TaskExecutionIdentifier{
				TaskId: &core.Identifier{
					Project: model.TaskProject,
					Domain:  model.TaskDomain,
					Name:    model.TaskName,
					Version: model.TaskVersion,
				},
				NodeExecutionId: &core.NodeExecutionIdentifier{
					NodeId: model.NodeID,
				},
				RetryAttempt: model.RetryAttempt,
			},
			Execution: &core.WorkflowExecutionIdentifier{
				Project: model.Project,
				Domain:  model.Domain,
				Name:    model.ExecutionName,
			},
			Principal:        model.Principal,
			ShortDescription: model.Description,
			UserMetadata:     nil,
			MetadataType:     model.MetadataType,
		},
		Tags: nil,
	}

	cp := PartitionsToIdl(model.Partitions)
	if cp != nil {
		a.ArtifactId.Dimensions = &core.ArtifactID_Partitions{
			Partitions: cp,
		}
	}

	return a
}
