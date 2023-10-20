package db

import (
	"context"
	"github.com/flyteorg/flyte/flyteartifacts/pkg/models"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/artifact"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/golang/protobuf/proto"
	"github.com/jackc/pgx/v5/pgtype"
)

func PartitionsIdlToHstore(idlPartitions *core.Partitions) pgtype.Hstore {
	ctx := context.Background()
	if idlPartitions == nil || idlPartitions.GetValue() == nil {
		return nil
	}
	var hstore = make(pgtype.Hstore)

	for k, v := range idlPartitions.GetValue() {
		if len(v.GetStaticValue()) == 0 {
			logger.Warningf(ctx, "Partition key [%s] missing static value, [%+v]", k, v.GetValue())
			continue
		}
		sv := v.GetStaticValue()
		hstore[k] = &sv
	}
	return hstore
}

func HstoreToIdlPartitions(hs pgtype.Hstore) *core.Partitions {
	if hs == nil || len(hs) == 0 {
		return nil
	}
	m := make(map[string]*core.LabelValue, len(hs))
	for k, v := range hs {
		m[k] = &core.LabelValue{
			Value: &core.LabelValue_StaticValue{
				StaticValue: *v,
			},
		}
	}
	return &core.Partitions{
		Value: m,
	}
}

func ServiceToGormModel(serviceModel models.Artifact) (Artifact, error) {
	partitions := PartitionsIdlToHstore(serviceModel.Artifact.GetArtifactId().GetPartitions())

	ga := Artifact{
		ArtifactKey: ArtifactKey{
			Project: serviceModel.Artifact.ArtifactId.ArtifactKey.Project,
			Domain:  serviceModel.Artifact.ArtifactId.ArtifactKey.Domain,
			Name:    serviceModel.Artifact.ArtifactId.ArtifactKey.Name,
		},
		Version:    serviceModel.Artifact.ArtifactId.Version,
		Partitions: partitions,

		LiteralType:           serviceModel.LiteralTypeBytes,
		LiteralValue:          serviceModel.LiteralValueBytes,
		Description:           serviceModel.Artifact.Spec.ShortDescription,
		MetadataType:          serviceModel.Artifact.Spec.MetadataType,
		OffloadedUserMetadata: serviceModel.OffloadedMetadata,

		ExecutionName: serviceModel.Artifact.Spec.Execution.Name,
	}

	if serviceModel.Artifact.Spec.TaskExecution != nil {
		ga.TaskProject = serviceModel.Artifact.Spec.TaskExecution.TaskId.Project
		ga.TaskDomain = serviceModel.Artifact.Spec.TaskExecution.TaskId.Domain
		ga.TaskName = serviceModel.Artifact.Spec.TaskExecution.TaskId.Name
		ga.TaskVersion = serviceModel.Artifact.Spec.TaskExecution.TaskId.Version
		ga.RetryAttempt = &serviceModel.Artifact.Spec.TaskExecution.RetryAttempt
	}

	return ga, nil
}

func GormToServiceModel(ga Artifact) (models.Artifact, error) {
	lt := &core.LiteralType{}
	lit := &core.Literal{}
	if err := proto.Unmarshal(ga.LiteralType, lt); err != nil {
		return models.Artifact{}, err
	}
	if err := proto.Unmarshal(ga.LiteralValue, lit); err != nil {
		return models.Artifact{}, err
	}

	// gatepr: principal is missing still - can be added following discussion on source object.
	// taskexecution and additional source information to be added when resolved.
	// gatepr: implement tags
	a := artifact.Artifact{
		ArtifactId: &core.ArtifactID{
			ArtifactKey: &core.ArtifactKey{
				Project: ga.ArtifactKey.Project,
				Domain:  ga.ArtifactKey.Domain,
				Name:    ga.ArtifactKey.Name,
			},
			Version: ga.Version,
		},
		Spec: &artifact.ArtifactSpec{
			Value:         lit,
			Type:          lt,
			TaskExecution: nil,
			Execution: &core.WorkflowExecutionIdentifier{
				Project: ga.ArtifactKey.Project,
				Domain:  ga.ArtifactKey.Domain,
				Name:    ga.ExecutionName,
			},
			Principal:        "",
			ShortDescription: ga.Description,
			UserMetadata:     nil,
			MetadataType:     ga.MetadataType,
		},
		Tags: nil,
	}
	p := HstoreToIdlPartitions(ga.Partitions)
	if p != nil {
		a.ArtifactId.Dimensions = &core.ArtifactID_Partitions{Partitions: p}
	}

	return models.Artifact{
		Artifact:          a,
		OffloadedMetadata: "",
		LiteralTypeBytes:  nil,
		LiteralValueBytes: nil,
	}, nil
}
