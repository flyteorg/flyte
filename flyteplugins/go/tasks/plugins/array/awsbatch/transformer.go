package awsbatch

import (
	"context"
	"sort"
	"time"

	"github.com/aws/aws-sdk-go/service/batch"
	"github.com/golang/protobuf/ptypes/duration"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	idlCore "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/errors"
	pluginCore "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core/template"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/flytek8s"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/utils"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/plugins/array"
	config2 "github.com/flyteorg/flyte/flyteplugins/go/tasks/plugins/array/awsbatch/config"
)

const (
	ArrayJobIndex       = "BATCH_JOB_ARRAY_INDEX_VAR_NAME"
	arrayJobIDFormatter = "%v:%v"
	failOnError         = "FLYTE_FAIL_ON_ERROR"
)

const assignResources = true

// Note that Name is not set on the result object.
// It's up to the caller to set the Name before creating the object in K8s.
func FlyteTaskToBatchInput(ctx context.Context, tCtx pluginCore.TaskExecutionContext, jobDefinition string, cfg *config2.Config) (
	batchInput *batch.SubmitJobInput, err error) {

	// Check that the taskTemplate is valid
	taskTemplate, err := tCtx.TaskReader().Read(ctx)
	if err != nil {
		return nil, err
	} else if taskTemplate == nil {
		return nil, errors.Errorf(errors.BadTaskSpecification, "Required value not set, taskTemplate is nil")
	}

	if taskTemplate.GetContainer() == nil {
		return nil, errors.Errorf(errors.BadTaskSpecification,
			"Required value not set, taskTemplate Container")
	}

	jobConfig := newJobConfig().
		MergeFromKeyValuePairs(taskTemplate.GetContainer().GetConfig()).
		MergeFromConfigMap(tCtx.TaskExecutionMetadata().GetOverrides().GetConfig())
	if len(jobConfig.DynamicTaskQueue) == 0 {
		return nil, errors.Errorf(errors.BadTaskSpecification, "config[%v] is missing", DynamicTaskQueueKey)
	}

	inputReader := array.GetInputReader(tCtx, taskTemplate)
	cmd, err := template.Render(
		ctx,
		taskTemplate.GetContainer().GetCommand(),
		template.Parameters{
			TaskExecMetadata: tCtx.TaskExecutionMetadata(),
			Inputs:           inputReader,
			OutputPath:       tCtx.OutputWriter(),
			Task:             tCtx.TaskReader(),
		})
	if err != nil {
		return nil, err
	}
	args, err := template.Render(ctx, taskTemplate.GetContainer().GetArgs(),
		template.Parameters{
			TaskExecMetadata: tCtx.TaskExecutionMetadata(),
			Inputs:           inputReader,
			OutputPath:       tCtx.OutputWriter(),
			Task:             tCtx.TaskReader(),
		})
	taskTemplate.GetContainer().GetEnv()
	if err != nil {
		return nil, err
	}

	envVars := getEnvVarsForTask(ctx, tCtx.TaskExecutionMetadata().GetTaskExecutionID(), taskTemplate.GetContainer().GetEnv(), cfg.DefaultEnvVars)

	// compile resources
	res, err := flytek8s.ToK8sResourceRequirements(taskTemplate.GetContainer().GetResources())
	if err != nil {
		return nil, err
	}

	if overrideResources := tCtx.TaskExecutionMetadata().GetOverrides().GetResources(); overrideResources != nil {
		flytek8s.MergeResources(*overrideResources, res)
	}

	platformResources := tCtx.TaskExecutionMetadata().GetPlatformResources()
	if platformResources == nil {
		platformResources = &v1.ResourceRequirements{}
	}

	flytek8s.SanitizeGPUResourceRequirements(res)
	resources := flytek8s.ApplyResourceOverrides(*res, *platformResources, assignResources)

	submitJobInput := &batch.SubmitJobInput{}
	if taskTemplate.GetCustom() != nil {
		err = utils.UnmarshalStructToObj(taskTemplate.GetCustom(), &submitJobInput)
		if err != nil {
			return nil, errors.Errorf(errors.BadTaskSpecification,
				"invalid TaskSpecification [%v], Err: [%v]", taskTemplate.GetCustom(), err.Error())
		}
	}
	submitJobInput.SetJobName(tCtx.TaskExecutionMetadata().GetTaskExecutionID().GetGeneratedName()).
		SetJobDefinition(jobDefinition).SetJobQueue(jobConfig.DynamicTaskQueue).
		SetRetryStrategy(toRetryStrategy(ctx, toBackoffLimit(taskTemplate.Metadata), cfg.MinRetries, cfg.MaxRetries)).
		SetContainerOverrides(toContainerOverrides(ctx, append(cmd, args...), &resources, envVars)).
		SetTimeout(toTimeout(taskTemplate.Metadata.GetTimeout(), cfg.DefaultTimeOut.Duration))

	return submitJobInput, nil
}

func UpdateBatchInputForArray(_ context.Context, batchInput *batch.SubmitJobInput, arraySize int64) *batch.SubmitJobInput {
	var arrayProps *batch.ArrayProperties
	envVars := batchInput.ContainerOverrides.Environment
	if arraySize > 1 {
		envVars = append(envVars, &batch.KeyValuePair{Name: refStr(ArrayJobIndex), Value: refStr("AWS_BATCH_JOB_ARRAY_INDEX")})
		arrayProps = &batch.ArrayProperties{
			Size: refInt(arraySize),
		}
	} else {
		// AWS Batch doesn't allow arrays of size 1. Since the task template defines the job as an array, we substitute
		// these env vars to make it look like one.
		envVars = append(envVars, &batch.KeyValuePair{Name: refStr(ArrayJobIndex), Value: refStr("FAKE_JOB_ARRAY_INDEX")},
			&batch.KeyValuePair{Name: refStr("FAKE_JOB_ARRAY_INDEX"), Value: refStr("0")})
	}
	batchInput.ArrayProperties = arrayProps
	batchInput.ContainerOverrides.Environment = envVars

	return batchInput
}

func getEnvVarsForTask(ctx context.Context, execID pluginCore.TaskExecutionID, containerEnvVars []*core.KeyValuePair,
	defaultEnvVars map[string]string) []v1.EnvVar {
	envVars, _ := flytek8s.DecorateEnvVars(ctx, flytek8s.ToK8sEnvVar(containerEnvVars), nil, execID, "")
	m := make(map[string]string, len(envVars))
	for _, envVar := range envVars {
		m[envVar.Name] = envVar.Value
	}

	for key, value := range defaultEnvVars {
		m[key] = value
	}
	m[failOnError] = "true"
	finalEnvVars := make([]v1.EnvVar, 0, len(m))
	for key, val := range m {
		finalEnvVars = append(finalEnvVars, v1.EnvVar{
			Name:  key,
			Value: val,
		})
	}
	return finalEnvVars
}

func toTimeout(templateTimeout *duration.Duration, defaultTimeout time.Duration) *batch.JobTimeout {
	if templateTimeout != nil && templateTimeout.Seconds > 0 {
		return (&batch.JobTimeout{}).SetAttemptDurationSeconds(templateTimeout.GetSeconds())
	}

	if defaultTimeout.Seconds() > 0 {
		return (&batch.JobTimeout{}).SetAttemptDurationSeconds(int64(defaultTimeout.Seconds()))
	}

	return nil
}

func toEnvironmentVariables(_ context.Context, envVars []v1.EnvVar) []*batch.KeyValuePair {
	res := make([]*batch.KeyValuePair, 0, len(envVars))
	for _, pair := range envVars {
		res = append(res, &batch.KeyValuePair{
			Name:  refStr(pair.Name),
			Value: refStr(pair.Value),
		})
	}

	sort.Slice(res, func(i, j int) bool {
		return *res[i].Name < *res[j].Name
	})

	return res
}

func toContainerOverrides(ctx context.Context, command []string, overrides *v1.ResourceRequirements,
	envVars []v1.EnvVar) *batch.ContainerOverrides {

	return &batch.ContainerOverrides{
		// Batch expects memory override in megabytes.
		Memory: refInt(overrides.Limits.Memory().ScaledValue(resource.Mega)),
		// Batch expects a rounded number of whole CPUs.
		Vcpus:       refInt(overrides.Limits.Cpu().Value()),
		Environment: toEnvironmentVariables(ctx, envVars),
		Command:     refStrSlice(command),
	}
}

func refStr(s string) *string {
	res := (s + " ")[:len(s)]
	return &res
}

func refStrSlice(s []string) []*string {
	res := make([]*string, 0, len(s))
	for _, str := range s {
		res = append(res, refStr(str))
	}

	return res
}

func refInt(i int64) *int64 {
	return &i
}

func toRetryStrategy(_ context.Context, backoffLimit *int32, minRetryAttempts, maxRetryAttempts int32) *batch.RetryStrategy {
	if backoffLimit == nil || *backoffLimit < 0 {
		return nil
	}

	retries := *backoffLimit
	if retries > maxRetryAttempts {
		retries = maxRetryAttempts
	}

	if retries < minRetryAttempts {
		retries = minRetryAttempts
	}

	return &batch.RetryStrategy{
		// Attempts is the total number of attempts for a task, if retries is set to 1, attempts should be set to 2 to
		// account for the first try.
		Attempts: refInt(int64(retries) + 1),
	}
}

func toBackoffLimit(metadata *idlCore.TaskMetadata) *int32 {
	if metadata == nil || metadata.Retries == nil {
		return nil
	}

	i := int32(metadata.Retries.Retries)
	return &i
}

func jobPhaseToPluginsPhase(jobStatus string) pluginCore.Phase {
	switch jobStatus {
	case batch.JobStatusSubmitted:
		fallthrough
	case batch.JobStatusPending:
		fallthrough
	case batch.JobStatusRunnable:
		fallthrough
	case batch.JobStatusStarting:
		return pluginCore.PhaseQueued
	case batch.JobStatusRunning:
		return pluginCore.PhaseRunning
	case batch.JobStatusSucceeded:
		return pluginCore.PhaseSuccess
	case batch.JobStatusFailed:
		// Retryable failure vs Permanent can be overridden if the task writes an errors.pb in the output prefix.
		return pluginCore.PhaseRetryableFailure
	}

	return pluginCore.PhaseUndefined
}
