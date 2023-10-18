package server

import (
	"fmt"
	"github.com/flyteorg/flyte/flyteartifacts/pkg/models"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/artifact"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
)

func CreateArtifactModelFromRequest(key *core.ArtifactKey, spec *artifact.ArtifactSpec, version string, partitions map[string]string, tag string, principal string) (models.Artifact, error) {
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

	a := artifact.Artifact{
		ArtifactId: &core.ArtifactID{
			ArtifactKey: &core.ArtifactKey{
				Project: key.Project,
				Domain:  key.Domain,
				Name:    key.Name,
			},
			Version: version,
		},
		Spec: spec,
		Tags: []string{tag},
	}

	if partitions != nil {
		cp := PartitionsToIdl(partitions)
		a.ArtifactId.Dimensions = &core.ArtifactID_Partitions{
			Partitions: cp,
		}
	}
	if principal != "" {
		a.Spec.Principal = principal
	}
	return models.Artifact{Artifact: a}, nil
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
