package sagemaker

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"

	pluginErrors "github.com/flyteorg/flyteplugins/go/tasks/errors"

	"k8s.io/client-go/kubernetes/scheme"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery"

	commonv1 "github.com/aws/amazon-sagemaker-operator-for-k8s/api/v1/common"
	hpojobv1 "github.com/aws/amazon-sagemaker-operator-for-k8s/api/v1/hyperparametertuningjob"
	trainingjobv1 "github.com/aws/amazon-sagemaker-operator-for-k8s/api/v1/trainingjob"
	pluginsCore "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/k8s"
)

// Sanity test that the plugin implements method of k8s.Plugin
var _ k8s.Plugin = awsSagemakerPlugin{}

type awsSagemakerPlugin struct {
	TaskType pluginsCore.TaskType
}

func (awsSagemakerPlugin) GetProperties() k8s.PluginProperties {
	return k8s.PluginProperties{}
}

func (m awsSagemakerPlugin) BuildIdentityResource(_ context.Context, _ pluginsCore.TaskExecutionMetadata) (client.Object, error) {
	if m.TaskType == trainingJobTaskType || m.TaskType == customTrainingJobTaskType {
		return &trainingjobv1.TrainingJob{}, nil
	}
	if m.TaskType == hyperparameterTuningJobTaskType {
		return &hpojobv1.HyperparameterTuningJob{}, nil
	}
	return nil, pluginErrors.Errorf(pluginErrors.BadTaskSpecification, "The sagemaker plugin is unable to build identity resource for an unknown task type [%v]", m.TaskType)
}

func (m awsSagemakerPlugin) BuildResource(ctx context.Context, taskCtx pluginsCore.TaskExecutionContext) (client.Object, error) {

	// Unmarshal the custom field of the task template back into the HyperparameterTuningJob struct generated in flyteidl
	if m.TaskType == trainingJobTaskType {
		return m.buildResourceForTrainingJob(ctx, taskCtx)
	}
	if m.TaskType == customTrainingJobTaskType {
		return m.buildResourceForCustomTrainingJob(ctx, taskCtx)
	}
	if m.TaskType == hyperparameterTuningJobTaskType {
		return m.buildResourceForHyperparameterTuningJob(ctx, taskCtx)
	}
	return nil, pluginErrors.Errorf(pluginErrors.BadTaskSpecification, "The SageMaker plugin is unable to build resource for unknown task type [%s]", m.TaskType)
}

func (m awsSagemakerPlugin) GetTaskPhase(ctx context.Context, pluginContext k8s.PluginContext, resource client.Object) (pluginsCore.PhaseInfo, error) {
	if m.TaskType == trainingJobTaskType {
		job := resource.(*trainingjobv1.TrainingJob)
		return m.getTaskPhaseForTrainingJob(ctx, pluginContext, job)
	} else if m.TaskType == customTrainingJobTaskType {
		job := resource.(*trainingjobv1.TrainingJob)
		return m.getTaskPhaseForCustomTrainingJob(ctx, pluginContext, job)
	} else if m.TaskType == hyperparameterTuningJobTaskType {
		job := resource.(*hpojobv1.HyperparameterTuningJob)
		return m.getTaskPhaseForHyperparameterTuningJob(ctx, pluginContext, job)
	}
	return pluginsCore.PhaseInfoUndefined, pluginErrors.Errorf(pluginErrors.BadTaskSpecification, "cannot get task phase for unknown task type [%s]", m.TaskType)
}

func init() {
	if err := commonv1.AddToScheme(scheme.Scheme); err != nil {
		panic(err)
	}

	// Registering the plugin for HyperparameterTuningJob
	pluginmachinery.PluginRegistry().RegisterK8sPlugin(
		k8s.PluginEntry{
			ID:                  hyperparameterTuningJobTaskPluginID,
			RegisteredTaskTypes: []pluginsCore.TaskType{hyperparameterTuningJobTaskType},
			ResourceToWatch:     &hpojobv1.HyperparameterTuningJob{},
			Plugin:              awsSagemakerPlugin{TaskType: hyperparameterTuningJobTaskType},
			IsDefault:           false,
		})

	// Registering the plugin for standalone TrainingJob
	pluginmachinery.PluginRegistry().RegisterK8sPlugin(
		k8s.PluginEntry{
			ID:                  trainingJobTaskPluginID,
			RegisteredTaskTypes: []pluginsCore.TaskType{trainingJobTaskType},
			ResourceToWatch:     &trainingjobv1.TrainingJob{},
			Plugin:              awsSagemakerPlugin{TaskType: trainingJobTaskType},
			IsDefault:           false,
		})

	// Registering the plugin for custom TrainingJob
	pluginmachinery.PluginRegistry().RegisterK8sPlugin(
		k8s.PluginEntry{
			ID:                  customTrainingJobTaskPluginID,
			RegisteredTaskTypes: []pluginsCore.TaskType{customTrainingJobTaskType},
			ResourceToWatch:     &trainingjobv1.TrainingJob{},
			Plugin:              awsSagemakerPlugin{TaskType: customTrainingJobTaskType},
			IsDefault:           false,
		})
}
