package register

import (
	"context"
	"errors"
	"fmt"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/lyft/flytectl/cmd/config"
	cmdCore "github.com/lyft/flytectl/cmd/core"
	"github.com/lyft/flytectl/pkg/printer"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/lyft/flytestdlib/logger"
)

//go:generate pflags RegisterFilesConfig

var (
	filesConfig = &RegisterFilesConfig{
		version:     "v1",
		skipOnError: false,
	}
)

const registrationProjectPattern = "{{ registration.project }}"
const registrationDomainPattern = "{{ registration.domain }}"
const registrationVersionPattern = "{{ registration.version }}"

type RegisterFilesConfig struct {
	version     string `json:"version" pflag:",version of the entity to be registered with flyte."`
	skipOnError bool   `json:"skipOnError" pflag:",fail fast when registering files."`
}

type RegisterResult struct {
	Name   string
	Status string
	Info   string
}

var projectColumns = []printer.Column{
	{"Name", "$.Name"},
	{"Status", "$.Status"},
	{"Additional Info", "$.Info"},
}

func unMarshalContents(ctx context.Context, fileContents []byte, fname string) (proto.Message, error) {
	workflowSpec := &admin.WorkflowSpec{}
	if err := proto.Unmarshal(fileContents, workflowSpec); err == nil {
		return workflowSpec, nil
	}
	logger.Debugf(ctx, "Failed to unmarshal file %v for workflow type", fname)
	taskSpec := &admin.TaskSpec{}
	if err := proto.Unmarshal(fileContents, taskSpec); err == nil {
		return taskSpec, nil
	}
	logger.Debugf(ctx, "Failed to unmarshal  file %v for task type", fname)
	launchPlan := &admin.LaunchPlan{}
	if err := proto.Unmarshal(fileContents, launchPlan); err == nil {
		return launchPlan, nil
	}
	logger.Debugf(ctx, "Failed to unmarshal file %v for launch plan type", fname)
	return nil, errors.New(fmt.Sprintf("Failed unmarshalling file %v", fname))

}

func register(ctx context.Context, message proto.Message, cmdCtx cmdCore.CommandContext) error {
	switch message.(type) {
	case *admin.LaunchPlan:
		launchPlan := message.(*admin.LaunchPlan)
		_, err := cmdCtx.AdminClient().CreateLaunchPlan(ctx, &admin.LaunchPlanCreateRequest{
			Id: &core.Identifier{
				ResourceType: core.ResourceType_LAUNCH_PLAN,
				Project:      config.GetConfig().Project,
				Domain:       config.GetConfig().Domain,
				Name:         launchPlan.Id.Name,
				Version:      filesConfig.version,
			},
			Spec: launchPlan.Spec,
		})
		return err
	case *admin.WorkflowSpec:
		workflowSpec := message.(*admin.WorkflowSpec)
		_, err := cmdCtx.AdminClient().CreateWorkflow(ctx, &admin.WorkflowCreateRequest{
			Id: &core.Identifier{
				ResourceType: core.ResourceType_WORKFLOW,
				Project:      config.GetConfig().Project,
				Domain:       config.GetConfig().Domain,
				Name:         workflowSpec.Template.Id.Name,
				Version:      filesConfig.version,
			},
			Spec: workflowSpec,
		})
		return err
	case *admin.TaskSpec:
		taskSpec := message.(*admin.TaskSpec)
		_, err := cmdCtx.AdminClient().CreateTask(ctx, &admin.TaskCreateRequest{
			Id: &core.Identifier{
				ResourceType: core.ResourceType_TASK,
				Project:      config.GetConfig().Project,
				Domain:       config.GetConfig().Domain,
				Name:         taskSpec.Template.Id.Name,
				Version:      filesConfig.version,
			},
			Spec: taskSpec,
		})
		return err
	default:
		return errors.New(fmt.Sprintf("Failed registering unknown entity  %v", message))
	}
}

func hydrateNode(node *core.Node) error {
	targetNode := node.Target
	switch targetNode.(type) {
	case *core.Node_TaskNode:
		taskNodeWrapper := targetNode.(*core.Node_TaskNode)
		taskNodeReference := taskNodeWrapper.TaskNode.Reference.(*core.TaskNode_ReferenceId)
		hydrateIdentifier(taskNodeReference.ReferenceId)
	case *core.Node_WorkflowNode:
		workflowNodeWrapper := targetNode.(*core.Node_WorkflowNode)
		switch workflowNodeWrapper.WorkflowNode.Reference.(type) {
		case *core.WorkflowNode_SubWorkflowRef:
			subWorkflowNodeReference := workflowNodeWrapper.WorkflowNode.Reference.(*core.WorkflowNode_SubWorkflowRef)
			hydrateIdentifier(subWorkflowNodeReference.SubWorkflowRef)
		case *core.WorkflowNode_LaunchplanRef:
			launchPlanNodeReference := workflowNodeWrapper.WorkflowNode.Reference.(*core.WorkflowNode_LaunchplanRef)
			hydrateIdentifier(launchPlanNodeReference.LaunchplanRef)
		default:
			errors.New(fmt.Sprintf("Unknown type %T", workflowNodeWrapper.WorkflowNode.Reference))
		}
	case *core.Node_BranchNode:
		branchNodeWrapper := targetNode.(*core.Node_BranchNode)
		hydrateNode(branchNodeWrapper.BranchNode.IfElse.Case.ThenNode)
		if len(branchNodeWrapper.BranchNode.IfElse.Other) > 0 {
			for _, ifBlock := range branchNodeWrapper.BranchNode.IfElse.Other {
				hydrateNode(ifBlock.ThenNode)
			}
		}
		switch branchNodeWrapper.BranchNode.IfElse.Default.(type) {
		case *core.IfElseBlock_ElseNode:
			elseNodeReference := branchNodeWrapper.BranchNode.IfElse.Default.(*core.IfElseBlock_ElseNode)
			hydrateNode(elseNodeReference.ElseNode)
		case *core.IfElseBlock_Error:
			// Do nothing.
		default:
			return errors.New(fmt.Sprintf("Unknown type %T", branchNodeWrapper.BranchNode.IfElse.Default))
		}
	default:
		return errors.New(fmt.Sprintf("Unknown type %T", targetNode))
	}
	return nil
}

func hydrateIdentifier(identifier *core.Identifier) {
	if identifier.Project == "" || identifier.Project == registrationProjectPattern {
		identifier.Project = config.GetConfig().Project
	}
	if identifier.Domain == "" || identifier.Domain == registrationDomainPattern {
		identifier.Domain = config.GetConfig().Domain
	}
	if identifier.Version == "" || identifier.Version == registrationVersionPattern {
		identifier.Version = filesConfig.version
	}
}

func hydrateSpec(message proto.Message) error {
	switch message.(type) {
	case *admin.LaunchPlan:
		launchPlan := message.(*admin.LaunchPlan)
		hydrateIdentifier(launchPlan.Spec.WorkflowId)
	case *admin.WorkflowSpec:
		workflowSpec := message.(*admin.WorkflowSpec)
		for _, Noderef := range workflowSpec.Template.Nodes {
			if err := hydrateNode(Noderef); err != nil {
				return err
			}
		}
		hydrateIdentifier(workflowSpec.Template.Id)
		for _, subWorkflow := range workflowSpec.SubWorkflows {
			for _, Noderef := range subWorkflow.Nodes {
				if err := hydrateNode(Noderef); err != nil {
					return err
				}
			}
			hydrateIdentifier(subWorkflow.Id)
		}
	case *admin.TaskSpec:
		taskSpec := message.(*admin.TaskSpec)
		hydrateIdentifier(taskSpec.Template.Id)
	default:
		return errors.New(fmt.Sprintf("Unknown type %T", message))
	}
	return nil
}

func getJsonSpec(message proto.Message) string {
	marshaller := jsonpb.Marshaler{
		EnumsAsInts:  false,
		EmitDefaults: true,
		Indent:       "  ",
		OrigName:     true,
	}
	jsonSpec, _ := marshaller.MarshalToString(message)
	return jsonSpec
}
