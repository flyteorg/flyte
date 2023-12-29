package models

import (
	"context"
	"fmt"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/artifact"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flytestdlib/logger"
)

func CreateArtifactModelFromRequest(ctx context.Context, key *core.ArtifactKey, spec *artifact.ArtifactSpec, version string, partitions map[string]string, tag string, source *artifact.ArtifactSource) (Artifact, error) {
	if key == nil || spec == nil {
		return Artifact{}, fmt.Errorf("key and spec cannot be nil")
	}
	if len(version) == 0 {
		return Artifact{}, fmt.Errorf("version cannot be empty")
	}
	if spec.Type == nil || spec.Value == nil {
		return Artifact{}, fmt.Errorf("spec type and value cannot be nil")
	}

	ex := source.GetWorkflowExecution()
	if ex == nil {
		return Artifact{}, fmt.Errorf("spec execution cannot be nil")
	}
	if ex.Project != key.Project || ex.Domain != key.Domain {
		return Artifact{}, fmt.Errorf("spec execution must match key")
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
		Spec:   spec,
		Tags:   []string{tag},
		Source: source,
	}

	if partitions != nil {
		cp := PartitionsToIdl(partitions)
		a.ArtifactId.Dimensions = &core.ArtifactID_Partitions{
			Partitions: cp,
		}
	}

	ltBytes, err := proto.Marshal(spec.Type)
	if err != nil {
		logger.Errorf(ctx, "Failed to marshal type for artifact: %+v@%s, err: %v", key, version, err)
		return Artifact{}, err
	}
	litBytes, err := proto.Marshal(spec.Value)
	if err != nil {
		logger.Errorf(ctx, "Failed to marshal literal value for artifact: %+v@%s, err: %v", key, version, err)
		return Artifact{}, err
	}

	return Artifact{
		Artifact:          a,
		LiteralTypeBytes:  ltBytes,
		LiteralValueBytes: litBytes,
	}, nil
}

func PartitionsToIdl(partitions map[string]string) *core.Partitions {
	if len(partitions) == 0 {
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

func PartitionsFromIdl(ctx context.Context, partitions *core.Partitions) map[string]string {
	if partitions == nil {
		return nil
	}

	p := make(map[string]string, len(partitions.Value))
	for k, v := range partitions.Value {
		if len(v.GetStaticValue()) == 0 {
			logger.Warningf(ctx, "Partition key [%s] missing static value, [%+v]", k, v.GetValue())
			continue
		}
		p[k] = v.GetStaticValue()
	}

	return p
}

func CreateTriggerModelFromRequest(ctx context.Context, request *artifact.CreateTriggerRequest) (Trigger, error) {
	if request.GetTriggerLaunchPlan().GetSpec() == nil || request.GetTriggerLaunchPlan().GetId() == nil || request.GetTriggerLaunchPlan().GetClosure() == nil {
		logger.Errorf(ctx, "Something nil in CreateTrigger, [%+v]", request)
		return Trigger{}, fmt.Errorf("invalid request to CreateTrigger, something is nil")
	}

	spec := request.GetTriggerLaunchPlan().GetSpec()
	if spec.GetEntityMetadata().GetLaunchConditions() == nil {
		logger.Errorf(ctx, "Launch conditions cannot be nil in CreateTrigger, [%+v]", request)
		return Trigger{}, fmt.Errorf("invalid request to CreateTrigger, launch conditions cannot be nil")
	}

	lpID := request.GetTriggerLaunchPlan().GetId()

	lc := spec.GetEntityMetadata().GetLaunchConditions()

	var err error
	idlTrigger := core.Trigger{}
	err = ptypes.UnmarshalAny(lc, &idlTrigger)
	if err != nil {
		logger.Errorf(ctx, "Failed to unmarshal launch conditions to idl, metadata: [%+v]", spec.GetEntityMetadata())
		return Trigger{}, err
	}
	if len(idlTrigger.Triggers) == 0 {
		return Trigger{}, fmt.Errorf("invalid request to CreateTrigger, launch conditions cannot be empty")
	}
	// Create a list of the Artifact IDs referenced by the trigger definition.
	// Keep in mind: these are not real IDs, they just contain partial information like the name.
	// Always set the referenced artifacts to the project/domain of the launch plan
	var runsOnArtifactIDs = make([]core.ArtifactID, len(idlTrigger.Triggers))
	for i, t := range idlTrigger.Triggers {
		runsOnArtifactIDs[i] = *t
		runsOnArtifactIDs[i].ArtifactKey.Project = lpID.Project
		runsOnArtifactIDs[i].ArtifactKey.Domain = lpID.Domain
	}

	specBytes, err := proto.Marshal(request.GetTriggerLaunchPlan().GetSpec())
	if err != nil {
		logger.Errorf(ctx, "Failed to marshal lp spec for CreateTrigger err: %v", err)
		return Trigger{}, err
	}
	closureBytes, err := proto.Marshal(request.GetTriggerLaunchPlan().GetClosure())
	if err != nil {
		logger.Errorf(ctx, "Failed to marshal lp closure for CreateTrigger err: %v", err)
		return Trigger{}, err
	}

	// Always set the project/domain of the trigger equal to the underlying launch plan
	t := Trigger{
		Project: lpID.Project,
		Domain:  lpID.Domain,
		Name:    idlTrigger.TriggerId.Name,
		// Use LP id for the version because the trigger doesn't come with its own
		// version for now... too difficult to update the version of the trigger
		// inside the launch conditions object during registration.
		Version:      lpID.Version,
		LaunchPlanID: *lpID,
		LaunchPlan:   request.GetTriggerLaunchPlan(),
		RunsOn:       runsOnArtifactIDs,
		Active:       true,
		SpecBytes:    specBytes,
		ClosureBytes: closureBytes,
	}

	return t, nil
}
