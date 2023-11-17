package processor

import (
	"context"
	"fmt"
	event2 "github.com/cloudevents/sdk-go/v2/event"
	"github.com/flyteorg/flyte/flyteartifacts/pkg/lib"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/artifact"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/event"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/golang/protobuf/proto"
)

// ServiceCallHandler will take events and call the grpc endpoints directly. The service should most likely be local.
type ServiceCallHandler struct {
	service artifact.ArtifactRegistryServer
	created chan<- artifact.Artifact
}

func (s *ServiceCallHandler) HandleEvent(ctx context.Context, cloudEvent *event2.Event, msg proto.Message) error {
	source := cloudEvent.Source()

	switch msgType := msg.(type) {
	case *event.CloudEventExecutionStart:
		logger.Debugf(ctx, "Handling CloudEventExecutionStart [%v]", msgType.ExecutionId)
		return s.HandleEventExecStart(ctx, msgType)
	case *event.CloudEventWorkflowExecution:
		logger.Debugf(ctx, "Handling CloudEventWorkflowExecution [%v]", msgType.RawEvent.ExecutionId)
		return s.HandleEventWorkflowExec(ctx, source, msgType)
	case *event.CloudEventTaskExecution:
		logger.Debugf(ctx, "Handling CloudEventTaskExecution [%v]", msgType.RawEvent.ParentNodeExecutionId)
		return s.HandleEventTaskExec(ctx, source, msgType)
	case *event.CloudEventNodeExecution:
		logger.Debugf(ctx, "Handling CloudEventNodeExecution [%v]", msgType.RawEvent.Id)
		return s.HandleEventNodeExec(ctx, source, msgType)
	default:
		return fmt.Errorf("HandleEvent found unknown message type [%T]", msgType)
	}
}

func (s *ServiceCallHandler) HandleEventExecStart(_ context.Context, _ *event.CloudEventExecutionStart) error {
	// metric
	return nil
}

// HandleEventWorkflowExec and the task one below are very similar. Can be combined in the future.
func (s *ServiceCallHandler) HandleEventWorkflowExec(ctx context.Context, source string, evt *event.CloudEventWorkflowExecution) error {

	if evt.RawEvent.Phase != core.WorkflowExecution_SUCCEEDED {
		logger.Debug(ctx, "Skipping non-successful workflow execution event")
		return nil
	}

	execID := evt.RawEvent.ExecutionId
	if evt.GetOutputData().GetLiterals() == nil || len(evt.OutputData.Literals) == 0 {
		logger.Debugf(ctx, "No output data to process for workflow event from [%v]", execID)
	}

	for varName, variable := range evt.OutputInterface.Outputs.Variables {
		if variable.GetArtifactPartialId() != nil {
			logger.Debugf(ctx, "Processing workflow output for %s, artifact name %s, from %v", varName, variable.GetArtifactPartialId().ArtifactKey.Name, execID)

			output := evt.OutputData.Literals[varName]

			// Add a tracking tag to the Literal before saving.
			version := fmt.Sprintf("%s/%s", source, varName)
			trackingTag := fmt.Sprintf("%s/%s/%s", execID.Project, execID.Domain, version)
			if output.Metadata == nil {
				output.Metadata = make(map[string]string, 1)
			}
			output.Metadata[lib.ArtifactKey] = trackingTag

			aSrc := &artifact.ArtifactSource{
				WorkflowExecution: execID,
				NodeId:            "end-node",
				Principal:         evt.Principal,
			}

			spec := artifact.ArtifactSpec{
				Value: output,
				Type:  evt.OutputInterface.Outputs.Variables[varName].Type,
			}

			partitions, tag, err := getPartitionsAndTag(
				ctx,
				*variable.GetArtifactPartialId(),
				variable,
				evt.InputData,
			)
			if err != nil {
				logger.Errorf(ctx, "failed processing [%s] variable [%v] with error: %v", varName, variable, err)
				return err
			}
			ak := core.ArtifactKey{
				Project: execID.Project,
				Domain:  execID.Domain,
				Name:    variable.GetArtifactPartialId().ArtifactKey.Name,
			}

			req := artifact.CreateArtifactRequest{
				ArtifactKey: &ak,
				Version:     version,
				Spec:        &spec,
				Partitions:  partitions,
				Tag:         tag,
				Source:      aSrc,
			}

			resp, err := s.service.CreateArtifact(ctx, &req)
			if err != nil {
				logger.Errorf(ctx, "failed to create artifact for [%s] with error: %v", varName, err)
				return err
			}
			// metric
			select {
			case s.created <- *resp.Artifact:
				logger.Debugf(ctx, "Sent %v from handle workflow", resp.Artifact.ArtifactId)
			default:
				// metric
				logger.Debugf(ctx, "Channel is full. didn't send %v", resp.Artifact.ArtifactId)
			}
			logger.Debugf(ctx, "Created wf artifact id [%+v] for key %s", resp.Artifact.ArtifactId, varName)
		}
	}

	return nil
}

func getPartitionsAndTag(ctx context.Context, partialID core.ArtifactID, variable *core.Variable, inputData *core.LiteralMap) (map[string]string, string, error) {
	if variable == nil || inputData == nil {
		return nil, "", fmt.Errorf("variable or input data is nil")
	}

	var partitions map[string]string
	// todo: consider updating idl to make CreateArtifactRequest just take a full Partitions
	// object rather than a mapstrstr @eapolinario @enghabu
	if partialID.GetPartitions().GetValue() != nil && len(partialID.GetPartitions().GetValue()) > 0 {
		partitions = make(map[string]string, len(partialID.GetPartitions().GetValue()))
		for k, lv := range partialID.GetPartitions().GetValue() {
			if lv.GetStaticValue() != "" {
				partitions[k] = lv.GetStaticValue()
			} else if lv.GetInputBinding() != nil {
				if lit, ok := inputData.Literals[lv.GetInputBinding().GetVar()]; ok {
					// todo: figure out formatting. Maybe we can add formatting directives to the input binding
					//   @enghabu @eapolinario
					renderedStr, err := lib.RenderLiteral(lit)
					if err != nil {
						logger.Errorf(ctx, "failed to render literal for input [%s] partition [%s] with error: %v", lv.GetInputBinding().GetVar(), k, err)
						return nil, "", err
					}
					partitions[k] = renderedStr
				} else {
					return nil, "", fmt.Errorf("input binding [%s] not found in input data", lv.GetInputBinding().GetVar())
				}
			} else {
				return nil, "", fmt.Errorf("unknown binding found in context of a materialized artifact")
			}
		}
	}

	var tag = ""
	var err error
	if lv := variable.GetArtifactTag().GetValue(); lv != nil {
		if lv.GetStaticValue() != "" {
			tag = lv.GetStaticValue()
		} else if lv.GetInputBinding() != nil {
			tag, err = lib.RenderLiteral(inputData.Literals[lv.GetInputBinding().GetVar()])
			if err != nil {
				logger.Errorf(ctx, "failed to render input [%s] for tag with error: %v", lv.GetInputBinding().GetVar(), err)
				return nil, "", err
			}
		} else {
			return nil, "", fmt.Errorf("triggered binding found in context of a materialized artifact when rendering tag")
		}
	}

	return partitions, tag, nil
}

func (s *ServiceCallHandler) HandleEventTaskExec(ctx context.Context, _ string, evt *event.CloudEventTaskExecution) error {

	if evt.RawEvent.Phase != core.TaskExecution_SUCCEEDED {
		logger.Debug(ctx, "Skipping non-successful task execution event")
		return nil
	}
	// metric

	return nil
}

func (s *ServiceCallHandler) HandleEventNodeExec(ctx context.Context, source string, evt *event.CloudEventNodeExecution) error {
	if evt.RawEvent.Phase != core.NodeExecution_SUCCEEDED {
		logger.Debug(ctx, "Skipping non-successful task execution event")
		return nil
	}
	if evt.RawEvent.Id.NodeId == "end-node" {
		logger.Debug(ctx, "Skipping end node for %s", evt.RawEvent.Id.ExecutionId.Name)
		return nil
	}
	// metric

	execID := evt.RawEvent.Id.ExecutionId
	if evt.GetOutputData().GetLiterals() == nil || len(evt.OutputData.Literals) == 0 {
		logger.Debugf(ctx, "No output data to process for task event from [%s] node %s", execID, evt.RawEvent.Id.NodeId)
	}

	if evt.OutputInterface == nil {
		if evt.GetOutputData() != nil {
			// metric this as error
			logger.Errorf(ctx, "No output interface to process for task event from [%s] node %s, but output data is not nil", execID, evt.RawEvent.Id.NodeId)
		}
		logger.Debugf(ctx, "No output interface to process for task event from [%s] node %s", execID, evt.RawEvent.Id.NodeId)
		return nil
	}

	if evt.RawEvent.GetTaskNodeMetadata() != nil {
		if evt.RawEvent.GetTaskNodeMetadata().CacheStatus == core.CatalogCacheStatus_CACHE_HIT {
			logger.Debugf(ctx, "Skipping cache hit for %s", evt.RawEvent.Id)
			return nil
		}
	}
	var taskExecID *core.TaskExecutionIdentifier
	if taskExecID = evt.GetTaskExecId(); taskExecID == nil {
		logger.Debugf(ctx, "No task execution id to process for task event from [%s] node %s", execID, evt.RawEvent.Id.NodeId)
	}

	// See note on the cloudevent_publisher side, we'll have to call one of the get data endpoints to get the actual data
	// rather than reading them here. But read here for now.

	// Iterate through the output interface. For any outputs that have an artifact ID specified, grab the
	// output Literal and construct a Create request and call the service.
	for varName, variable := range evt.OutputInterface.Outputs.Variables {
		if variable.GetArtifactPartialId() != nil {
			logger.Debugf(ctx, "Processing output for %s, artifact name %s, from %v", varName, variable.GetArtifactPartialId().ArtifactKey.Name, execID)

			output := evt.OutputData.Literals[varName]

			// Add a tracking tag to the Literal before saving.
			version := fmt.Sprintf("%s/%d/%s", source, taskExecID.RetryAttempt, varName)
			trackingTag := fmt.Sprintf("%s/%s/%s", execID.Project, execID.Domain, version)
			if output.Metadata == nil {
				output.Metadata = make(map[string]string, 1)
			}
			output.Metadata[lib.ArtifactKey] = trackingTag

			aSrc := &artifact.ArtifactSource{
				WorkflowExecution: execID,
				NodeId:            evt.RawEvent.Id.NodeId,
				Principal:         evt.Principal,
			}

			if taskExecID != nil {
				aSrc.RetryAttempt = taskExecID.RetryAttempt
				aSrc.TaskId = taskExecID.TaskId
			}

			spec := artifact.ArtifactSpec{
				Value: output,
				Type:  evt.OutputInterface.Outputs.Variables[varName].Type,
			}

			partitions, tag, err := getPartitionsAndTag(
				ctx,
				*variable.GetArtifactPartialId(),
				variable,
				evt.InputData,
			)
			if err != nil {
				logger.Errorf(ctx, "failed processing [%s] variable [%v] with error: %v", varName, variable, err)
				return err
			}
			ak := core.ArtifactKey{
				Project: execID.Project,
				Domain:  execID.Domain,
				Name:    variable.GetArtifactPartialId().ArtifactKey.Name,
			}

			req := artifact.CreateArtifactRequest{
				ArtifactKey: &ak,
				Version:     version,
				Spec:        &spec,
				Partitions:  partitions,
				Tag:         tag,
				Source:      aSrc,
			}

			resp, err := s.service.CreateArtifact(ctx, &req)
			if err != nil {
				logger.Errorf(ctx, "failed to create artifact for [%s] with error: %v", varName, err)
				return err
			}
			// metric
			select {
			case s.created <- *resp.Artifact:
				logger.Debugf(ctx, "Sent %v from handle task", resp.Artifact.ArtifactId)
			default:
				// metric
				logger.Debugf(ctx, "Channel is full. task handler didn't send %v", resp.Artifact.ArtifactId)
			}

			logger.Debugf(ctx, "Created artifact id [%+v] for key %s", resp.Artifact.ArtifactId, varName)
		}
	}
	return nil
}

func NewServiceCallHandler(ctx context.Context, svc artifact.ArtifactRegistryServer, created chan<- artifact.Artifact) EventsHandlerInterface {
	logger.Infof(ctx, "Creating new service call handler")
	return &ServiceCallHandler{
		service: svc,
		created: created,
	}
}
