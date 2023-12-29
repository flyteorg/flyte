package db

import (
	"context"
	"fmt"
	"github.com/flyteorg/flyte/flyteartifacts/pkg/models"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/artifact"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
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
	if len(hs) == 0 {
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
	}

	if serviceModel.Artifact.GetSource().GetWorkflowExecution() != nil {
		// artifact and execution project/domains are always the same.
		// Note the service model will not have workflow execution if it was an upload
		wfExec := WorkflowExecution{
			ExecutionProject: serviceModel.Artifact.ArtifactId.ArtifactKey.Project,
			ExecutionDomain:  serviceModel.Artifact.ArtifactId.ArtifactKey.Domain,
			ExecutionName:    serviceModel.Source.WorkflowExecution.Name,
		}
		ga.WorkflowExecution = wfExec
		ga.NodeID = serviceModel.Source.NodeId
	}
	if serviceModel.GetSource() != nil {
		ga.Principal = serviceModel.GetSource().GetPrincipal()
	}

	if serviceModel.GetSource().GetTaskId() != nil {
		// If task id is there, so should the retry attempt
		retry := serviceModel.GetSource().GetRetryAttempt()
		ga.RetryAttempt = &retry
		ga.TaskProject = serviceModel.GetSource().GetTaskId().Project
		ga.TaskDomain = serviceModel.GetSource().GetTaskId().Domain
		ga.TaskName = serviceModel.GetSource().GetTaskId().Name
		ga.TaskVersion = serviceModel.GetSource().GetTaskId().Version
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
			Value:            lit,
			Type:             lt,
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
	aSrc := artifact.ArtifactSource{
		NodeId:    ga.NodeID,
		Principal: ga.Principal,
	}
	if ga.RetryAttempt != nil {
		aSrc.RetryAttempt = *ga.RetryAttempt
	}
	if ga.WorkflowExecutionID != 0 {
		execID := &core.WorkflowExecutionIdentifier{
			Project: ga.ArtifactKey.Project,
			Domain:  ga.ArtifactKey.Domain,
			Name:    ga.WorkflowExecution.ExecutionName,
		}
		aSrc.WorkflowExecution = execID
	}
	if ga.TaskProject != "" {
		aSrc.TaskId = &core.Identifier{
			ResourceType: core.ResourceType_TASK,
			Project:      ga.TaskProject,
			Domain:       ga.TaskDomain,
			Name:         ga.TaskName,
			Version:      ga.TaskVersion,
		}
	}
	a.Source = &aSrc

	return models.Artifact{
		Artifact:          a,
		OffloadedMetadata: "",
		LiteralTypeBytes:  ga.LiteralType,
		LiteralValueBytes: ga.LiteralValue,
	}, nil
}

func ServiceToGormTrigger(serviceTrigger models.Trigger) Trigger {

	t := Trigger{
		TriggerKey: TriggerKey{
			Project: serviceTrigger.Project,
			Domain:  serviceTrigger.Domain,
			Name:    serviceTrigger.Name,
		},
		Version: serviceTrigger.Version,
		Active:  serviceTrigger.Active,
		LaunchPlanID: LaunchPlanID{
			Name:    serviceTrigger.LaunchPlan.Id.Name,
			Version: serviceTrigger.LaunchPlan.Id.Version,
		},
		LaunchPlanSpec:    serviceTrigger.SpecBytes,
		LaunchPlanClosure: serviceTrigger.ClosureBytes,
	}

	var runsOn = make([]ArtifactKey, len(serviceTrigger.RunsOn))
	for i, a := range serviceTrigger.RunsOn {
		runsOn[i] = ArtifactKey{
			Project: a.ArtifactKey.Project,
			Domain:  a.ArtifactKey.Domain,
			Name:    a.ArtifactKey.Name,
		}
	}
	t.RunsOn = runsOn

	return t
}

func GormToServiceTrigger(gormTrigger Trigger) (models.Trigger, error) {
	spec := &admin.LaunchPlanSpec{}
	closure := &admin.LaunchPlanClosure{}
	if err := proto.Unmarshal(gormTrigger.LaunchPlanSpec, spec); err != nil {
		return models.Trigger{}, err
	}
	if err := proto.Unmarshal(gormTrigger.LaunchPlanClosure, closure); err != nil {
		return models.Trigger{}, err
	}
	lpID := core.Identifier{
		ResourceType: core.ResourceType_LAUNCH_PLAN,
		Project:      gormTrigger.TriggerKey.Project,
		Domain:       gormTrigger.TriggerKey.Domain,
		Name:         gormTrigger.LaunchPlanID.Name,
		Version:      gormTrigger.Version, // gormTrigger.LaunchPlanID.Version,
	}
	t := models.Trigger{
		Project:      gormTrigger.TriggerKey.Project,
		Domain:       gormTrigger.TriggerKey.Domain,
		Name:         gormTrigger.TriggerKey.Name,
		Version:      gormTrigger.Version,
		Active:       gormTrigger.Active,
		LaunchPlanID: lpID,
		LaunchPlan: &admin.LaunchPlan{
			Id:      &lpID,
			Spec:    spec,
			Closure: closure,
		},
		SpecBytes:    gormTrigger.LaunchPlanSpec,
		ClosureBytes: gormTrigger.LaunchPlanClosure,
	}

	// TODO: This is a copy/paste of the code in transformers.go. Refactor.
	// Basically the DB model only has artifact keys, not whole artifact IDs including partitions
	// so pull the artifact IDs again from the spec.
	lc := spec.GetEntityMetadata().GetLaunchConditions()

	var err error
	idlTrigger := core.Trigger{}
	err = ptypes.UnmarshalAny(lc, &idlTrigger)
	if err != nil {
		logger.Errorf(context.TODO(), "Failed to unmarshal launch conditions to idl, metadata: [%+v]", spec.GetEntityMetadata())
		return models.Trigger{}, err
	}
	if len(idlTrigger.Triggers) == 0 {
		return models.Trigger{}, fmt.Errorf("invalid request to CreateTrigger, launch conditions cannot be empty")
	}
	var runsOnArtifactIDs = make([]core.ArtifactID, len(idlTrigger.Triggers))
	for i, t := range idlTrigger.Triggers {
		runsOnArtifactIDs[i] = *t
		runsOnArtifactIDs[i].ArtifactKey.Project = lpID.Project
		runsOnArtifactIDs[i].ArtifactKey.Domain = lpID.Domain
	}

	t.RunsOn = runsOnArtifactIDs

	return t, nil
}
