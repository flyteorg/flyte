package db

import (
	"github.com/flyteorg/flyte/flyteartifacts/pkg/models"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/jackc/pgx/v5/pgtype"
)

func PartitionsIdlToHstore(idlPartitions *core.Partitions) pgtype.Hstore {
	if idlPartitions == nil || idlPartitions.GetValue() == nil {
		return nil
	}
	var hstore = make(pgtype.Hstore)

	for k, v := range idlPartitions.GetValue() {
		sv := v.GetStaticValue()
		hstore[k] = &sv
	}
	return hstore
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
