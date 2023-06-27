package bigquery

import (
	"context"
	"encoding/gob"
	"fmt"
	"net/http"
	"time"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/ioutils"

	"golang.org/x/oauth2"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/flytek8s"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/google"
	structpb "github.com/golang/protobuf/ptypes/struct"
	"google.golang.org/api/bigquery/v2"
	"google.golang.org/api/googleapi"

	flyteIdlCore "github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	pluginErrors "github.com/flyteorg/flyteplugins/go/tasks/errors"
	pluginsCore "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
	"google.golang.org/api/option"

	"github.com/flyteorg/flytestdlib/logger"

	"github.com/flyteorg/flytestdlib/promutils"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/webapi"
)

const (
	bigqueryQueryJobTask  = "bigquery_query_job_task"
	bigqueryConsolePath   = "https://console.cloud.google.com/bigquery"
	bigqueryStatusRunning = "RUNNING"
	bigqueryStatusPending = "PENDING"
	bigqueryStatusDone    = "DONE"
)

type Plugin struct {
	metricScope       promutils.Scope
	cfg               *Config
	googleTokenSource google.TokenSourceFactory
}

type ResourceWrapper struct {
	Status         *bigquery.JobStatus
	CreateError    *googleapi.Error
	OutputLocation string
}

type ResourceMetaWrapper struct {
	K8sServiceAccount string
	Namespace         string
	JobReference      bigquery.JobReference
}

func (p Plugin) GetConfig() webapi.PluginConfig {
	return GetConfig().WebAPI
}

func (p Plugin) ResourceRequirements(_ context.Context, _ webapi.TaskExecutionContextReader) (
	namespace core.ResourceNamespace, constraints core.ResourceConstraintsSpec, err error) {

	// Resource requirements are assumed to be the same.
	return "default", p.cfg.ResourceConstraints, nil
}

func (p Plugin) Create(ctx context.Context, taskCtx webapi.TaskExecutionContextReader) (webapi.ResourceMeta,
	webapi.Resource, error) {
	return p.createImpl(ctx, taskCtx)
}

func (p Plugin) createImpl(ctx context.Context, taskCtx webapi.TaskExecutionContextReader) (*ResourceMetaWrapper,
	*ResourceWrapper, error) {

	taskTemplate, err := taskCtx.TaskReader().Read(ctx)
	jobID := taskCtx.TaskExecutionMetadata().GetTaskExecutionID().GetGeneratedName()

	if err != nil {
		return nil, nil, pluginErrors.Wrapf(pluginErrors.RuntimeFailure, err, "unable to fetch task specification")
	}

	inputs, err := taskCtx.InputReader().Get(ctx)

	if err != nil {
		return nil, nil, pluginErrors.Wrapf(pluginErrors.RuntimeFailure, err, "unable to fetch task inputs")
	}

	var job *bigquery.Job

	namespace := taskCtx.TaskExecutionMetadata().GetNamespace()
	k8sServiceAccount := flytek8s.GetServiceAccountNameFromTaskExecutionMetadata(taskCtx.TaskExecutionMetadata())
	identity := google.Identity{K8sNamespace: namespace, K8sServiceAccount: k8sServiceAccount}
	client, err := p.newBigQueryClient(ctx, identity)

	if err != nil {
		return nil, nil, pluginErrors.Wrapf(pluginErrors.RuntimeFailure, err, "unable to get bigquery client")
	}

	if taskTemplate.Type == bigqueryQueryJobTask {
		job, err = createQueryJob(jobID, taskTemplate.GetCustom(), inputs)
	} else {
		err = pluginErrors.Errorf(pluginErrors.BadTaskSpecification, "unexpected task type [%v]", taskTemplate.Type)
	}

	if err != nil {
		return nil, nil, err
	}

	job.Configuration.Query.Query = taskTemplate.GetSql().Statement
	job.Configuration.Labels = taskCtx.TaskExecutionMetadata().GetLabels()

	resp, err := client.Jobs.Insert(job.JobReference.ProjectId, job).Do()

	if err != nil {
		apiError, ok := err.(*googleapi.Error)
		resourceMeta := ResourceMetaWrapper{
			JobReference:      *job.JobReference,
			Namespace:         namespace,
			K8sServiceAccount: k8sServiceAccount,
		}

		if ok && apiError.Code == 409 {
			job, err := client.Jobs.Get(resourceMeta.JobReference.ProjectId, resourceMeta.JobReference.JobId).Do()

			if err != nil {
				err := pluginErrors.Wrapf(
					pluginErrors.RuntimeFailure,
					err,
					"failed to get job [%s]",
					formatJobReference(resourceMeta.JobReference))

				return nil, nil, err
			}

			resource := ResourceWrapper{Status: job.Status}

			return &resourceMeta, &resource, nil
		}

		if ok {
			resource := ResourceWrapper{CreateError: apiError}

			return &resourceMeta, &resource, nil
		}

		return nil, nil, pluginErrors.Wrapf(pluginErrors.RuntimeFailure, err, "failed to create query job")
	}

	var outputLocation string
	if resp.Status != nil && resp.Status.State == bigqueryStatusDone {
		getResp, err := client.Jobs.Get(job.JobReference.ProjectId, job.JobReference.JobId).Do()

		if err != nil {
			err := pluginErrors.Wrapf(
				pluginErrors.RuntimeFailure,
				err,
				"failed to get job [%s]",
				formatJobReference(*job.JobReference))

			return nil, nil, err
		}
		outputLocation = constructOutputLocation(ctx, getResp)
	}
	resource := ResourceWrapper{Status: resp.Status, OutputLocation: outputLocation}
	resourceMeta := ResourceMetaWrapper{
		JobReference:      *job.JobReference,
		Namespace:         namespace,
		K8sServiceAccount: k8sServiceAccount,
	}

	return &resourceMeta, &resource, nil
}

func createQueryJob(jobID string, custom *structpb.Struct, inputs *flyteIdlCore.LiteralMap) (*bigquery.Job, error) {
	queryJobConfig, err := unmarshalQueryJobConfig(custom)

	if err != nil {
		return nil, pluginErrors.Wrapf(pluginErrors.BadTaskSpecification, err, "can't unmarshall struct to QueryJobConfig")
	}

	jobConfigurationQuery, err := getJobConfigurationQuery(queryJobConfig, inputs)

	if err != nil {
		return nil, pluginErrors.Wrapf(pluginErrors.BadTaskSpecification, err, "unable to fetch task inputs")
	}

	jobReference := bigquery.JobReference{
		JobId:     jobID,
		Location:  queryJobConfig.Location,
		ProjectId: queryJobConfig.ProjectID,
	}

	return &bigquery.Job{
		Configuration: &bigquery.JobConfiguration{
			Query: jobConfigurationQuery,
		},
		JobReference: &jobReference,
	}, nil
}

func (p Plugin) Get(ctx context.Context, taskCtx webapi.GetContext) (latest webapi.Resource, err error) {
	return p.getImpl(ctx, taskCtx)
}

func (p Plugin) getImpl(ctx context.Context, taskCtx webapi.GetContext) (wrapper *ResourceWrapper, err error) {
	resourceMeta := taskCtx.ResourceMeta().(*ResourceMetaWrapper)

	identity := google.Identity{
		K8sNamespace:      resourceMeta.Namespace,
		K8sServiceAccount: resourceMeta.K8sServiceAccount,
	}
	client, err := p.newBigQueryClient(ctx, identity)

	if err != nil {
		return nil, pluginErrors.Wrapf(pluginErrors.RuntimeFailure, err, "unable to get client")
	}

	job, err := client.Jobs.Get(resourceMeta.JobReference.ProjectId, resourceMeta.JobReference.JobId).Location(resourceMeta.JobReference.Location).Do()

	if err != nil {
		err := pluginErrors.Wrapf(
			pluginErrors.RuntimeFailure,
			err,
			"failed to get job [%s]",
			formatJobReference(resourceMeta.JobReference))

		return nil, err
	}

	outputLocation := constructOutputLocation(ctx, job)
	return &ResourceWrapper{
		Status:         job.Status,
		OutputLocation: outputLocation,
	}, nil
}

func (p Plugin) Delete(ctx context.Context, taskCtx webapi.DeleteContext) error {
	if taskCtx.ResourceMeta() == nil {
		return nil
	}

	resourceMeta := taskCtx.ResourceMeta().(*ResourceMetaWrapper)

	identity := google.Identity{
		K8sNamespace:      resourceMeta.Namespace,
		K8sServiceAccount: resourceMeta.K8sServiceAccount,
	}
	client, err := p.newBigQueryClient(ctx, identity)

	if err != nil {
		return err
	}

	_, err = client.Jobs.Cancel(resourceMeta.JobReference.ProjectId, resourceMeta.JobReference.JobId).Location(resourceMeta.JobReference.Location).Do()

	if err != nil {
		return err
	}

	logger.Info(ctx, "Cancelled job [%s]", formatJobReference(resourceMeta.JobReference))

	return nil
}

func (p Plugin) Status(ctx context.Context, tCtx webapi.StatusContext) (phase core.PhaseInfo, err error) {
	resourceMeta := tCtx.ResourceMeta().(*ResourceMetaWrapper)
	resource := tCtx.Resource().(*ResourceWrapper)
	version := pluginsCore.DefaultPhaseVersion

	if resource == nil {
		return core.PhaseInfoUndefined, nil
	}

	taskInfo := createTaskInfo(resourceMeta)

	if resource.CreateError != nil {
		return handleCreateError(resource.CreateError, taskInfo), nil
	}

	switch resource.Status.State {
	case bigqueryStatusPending:
		return core.PhaseInfoQueuedWithTaskInfo(version, "Query is PENDING", taskInfo), nil

	case bigqueryStatusRunning:
		return core.PhaseInfoRunning(version, taskInfo), nil

	case bigqueryStatusDone:
		if resource.Status.ErrorResult != nil {
			return handleErrorResult(
				resource.Status.ErrorResult.Reason,
				resource.Status.ErrorResult.Message,
				taskInfo), nil
		}
		err = writeOutput(ctx, tCtx, resource.OutputLocation)
		if err != nil {
			logger.Warnf(ctx, "Failed to write output, uri [%s], err %s", resource.OutputLocation, err.Error())
			return core.PhaseInfoUndefined, err
		}
		return pluginsCore.PhaseInfoSuccess(taskInfo), nil
	}

	return core.PhaseInfoUndefined, pluginErrors.Errorf(pluginsCore.SystemErrorCode, "unknown execution phase [%v].", resource.Status.State)
}

func constructOutputLocation(ctx context.Context, job *bigquery.Job) string {
	if job == nil || job.Configuration == nil || job.Configuration.Query == nil || job.Configuration.Query.DestinationTable == nil {
		return ""
	}
	dst := job.Configuration.Query.DestinationTable
	outputLocation := fmt.Sprintf("bq://%v:%v.%v", dst.ProjectId, dst.DatasetId, dst.TableId)
	logger.Debugf(ctx, "BigQuery saves query results to [%v]", outputLocation)
	return outputLocation
}

func writeOutput(ctx context.Context, tCtx webapi.StatusContext, OutputLocation string) error {
	taskTemplate, err := tCtx.TaskReader().Read(ctx)
	if err != nil {
		return err
	}

	if taskTemplate.Interface == nil || taskTemplate.Interface.Outputs == nil || taskTemplate.Interface.Outputs.Variables == nil {
		logger.Infof(ctx, "The task declares no outputs. Skipping writing the outputs.")
		return nil
	}

	resultsStructuredDatasetType, exists := taskTemplate.Interface.Outputs.Variables["results"]
	if !exists {
		logger.Infof(ctx, "The task declares no outputs. Skipping writing the outputs.")
		return nil
	}
	return tCtx.OutputWriter().Put(ctx, ioutils.NewInMemoryOutputReader(
		&flyteIdlCore.LiteralMap{
			Literals: map[string]*flyteIdlCore.Literal{
				"results": {
					Value: &flyteIdlCore.Literal_Scalar{
						Scalar: &flyteIdlCore.Scalar{
							Value: &flyteIdlCore.Scalar_StructuredDataset{
								StructuredDataset: &flyteIdlCore.StructuredDataset{
									Uri: OutputLocation,
									Metadata: &flyteIdlCore.StructuredDatasetMetadata{
										StructuredDatasetType: resultsStructuredDatasetType.GetType().GetStructuredDatasetType(),
									},
								},
							},
						},
					},
				},
			},
		}, nil, nil))
}

func handleCreateError(createError *googleapi.Error, taskInfo *core.TaskInfo) core.PhaseInfo {
	code := fmt.Sprintf("http%d", createError.Code)

	userExecutionError := &flyteIdlCore.ExecutionError{
		Message: createError.Message,
		Kind:    flyteIdlCore.ExecutionError_USER,
		Code:    code,
	}

	systemExecutionError := &flyteIdlCore.ExecutionError{
		Message: createError.Message,
		Kind:    flyteIdlCore.ExecutionError_SYSTEM,
		Code:    code,
	}

	if createError.Code >= http.StatusBadRequest && createError.Code < http.StatusInternalServerError {
		return core.PhaseInfoFailed(pluginsCore.PhasePermanentFailure, userExecutionError, taskInfo)
	}

	if createError.Code >= http.StatusInternalServerError {
		return core.PhaseInfoFailed(pluginsCore.PhaseRetryableFailure, systemExecutionError, taskInfo)
	}

	// something unexpected happened, just terminate task
	return core.PhaseInfoFailed(pluginsCore.PhasePermanentFailure, systemExecutionError, taskInfo)
}

func handleErrorResult(reason string, message string, taskInfo *core.TaskInfo) core.PhaseInfo {
	phaseCode := reason
	phaseReason := message

	// see https://cloud.google.com/bigquery/docs/error-messages

	// user errors are errors where users have to take action, e.g. fix their code
	// all errors with project configuration are also considered as user errors

	// system errors are errors where system doesn't work well and system owners have to take action
	// all errors internal to BigQuery are also considered as system errors

	// transient errors are retryable, if any action is needed, errors are permanent

	switch reason {
	case "":
		return pluginsCore.PhaseInfoSuccess(taskInfo)

	// This error returns when you try to access a resource such as a dataset, table, view, or job that you
	// don't have access to. This error also returns when you try to modify a read-only object.
	case "accessDenied":
		return pluginsCore.PhaseInfoFailure(phaseCode, phaseReason, taskInfo)

	// This error returns when there is a temporary server failure such as a network connection problem or
	// a server overload.
	case "backendError":
		return pluginsCore.PhaseInfoSystemRetryableFailure(phaseCode, phaseReason, taskInfo)

	// This error returns when billing isn't enabled for the project.
	case "billingNotEnabled":
		return pluginsCore.PhaseInfoFailure(phaseCode, phaseReason, taskInfo)

	// This error returns when BigQuery has temporarily denylisted the operation you attempted to perform,
	// usually to prevent a service outage. This error rarely occurs.
	case "blocked":
		return pluginsCore.PhaseInfoFailure(phaseCode, phaseReason, taskInfo)

	// This error returns when trying to create a job, dataset, or table that already exists. The error also
	// returns when a job's writeDisposition property is set to WRITE_EMPTY and the destination table accessed
	// by the job already exists.
	case "duplicate":
		return pluginsCore.PhaseInfoFailure(phaseCode, phaseReason, taskInfo)

	// This error returns when an internal error occurs within BigQuery.
	case "internalError":
		return pluginsCore.PhaseInfoSystemRetryableFailure(phaseCode, phaseReason, taskInfo)

	// This error returns when there is any kind of invalid input other than an invalid query, such as missing
	// required fields or an invalid table schema. Invalid queries return an invalidQuery error instead.
	case "invalid":
		return pluginsCore.PhaseInfoFailure(phaseCode, phaseReason, taskInfo)

	// This error returns when you attempt to run an invalid query.
	case "invalidQuery":
		return pluginsCore.PhaseInfoFailure(phaseCode, phaseReason, taskInfo)

	// This error returns when you attempt to schedule a query with invalid user credentials.
	case "invalidUser":
		return pluginsCore.PhaseInfoSystemRetryableFailure(phaseCode, phaseReason, taskInfo)

	// This error returns when you refer to a resource (a dataset, a table, or a job) that doesn't exist.
	// This can also occur when using snapshot decorators to refer to deleted tables that have recently been
	// streamed to.
	case "notFound":
		return pluginsCore.PhaseInfoFailure(phaseCode, phaseReason, taskInfo)

	// This job error returns when you try to access a feature that isn't implemented.
	case "notImplemented":
		return pluginsCore.PhaseInfoFailure(phaseCode, phaseReason, taskInfo)

	// This error returns when your project exceeds a BigQuery quota, a custom quota, or when you haven't set up
	// billing and you have exceeded the free tier for queries.
	case "quotaExceeded":
		return pluginsCore.PhaseInfoRetryableFailure(phaseCode, phaseReason, taskInfo)

	case "rateLimitExceeded":
		return pluginsCore.PhaseInfoRetryableFailure(phaseCode, phaseReason, taskInfo)

	// This error returns when you try to delete a dataset that contains tables or when you try to delete a job
	// that is currently running.
	case "resourceInUse":
		return pluginsCore.PhaseInfoSystemRetryableFailure(phaseCode, phaseReason, taskInfo)

	// This error returns when your query uses too many resources.
	case "resourcesExceeded":
		return pluginsCore.PhaseInfoFailure(phaseCode, phaseReason, taskInfo)

	// This error returns when your query's results are larger than the maximum response size. Some queries execute
	// in multiple stages, and this error returns when any stage returns a response size that is too large, even if
	// the final result is smaller than the maximum. This error commonly returns when queries use an ORDER BY
	// clause.
	case "responseTooLarge":
		return pluginsCore.PhaseInfoFailure(phaseCode, phaseReason, taskInfo)

	// This status code returns when a job is canceled.
	case "stopped":
		return pluginsCore.PhaseInfoFailure(phaseCode, phaseReason, taskInfo)

	// Certain BigQuery tables are backed by data managed by other Google product teams. This error indicates that
	// one of these tables is unavailable.
	case "tableUnavailable":
		return pluginsCore.PhaseInfoSystemRetryableFailure(phaseCode, phaseReason, taskInfo)

	// The job timed out.
	case "timeout":
		return pluginsCore.PhaseInfoFailure(phaseCode, phaseReason, taskInfo)

	default:
		return pluginsCore.PhaseInfoSystemFailure(phaseCode, phaseReason, taskInfo)
	}
}

func createTaskInfo(resourceMeta *ResourceMetaWrapper) *core.TaskInfo {
	timeNow := time.Now()
	j := formatJobReferenceForQueryParam(resourceMeta.JobReference)

	return &core.TaskInfo{
		OccurredAt: &timeNow,
		Logs: []*flyteIdlCore.TaskLog{
			{
				Uri: fmt.Sprintf("%s?project=%v&j=%v&page=queryresults",
					bigqueryConsolePath,
					resourceMeta.JobReference.ProjectId,
					j),
				Name: "BigQuery Console",
			},
		},
	}
}

func formatJobReference(reference bigquery.JobReference) string {
	return fmt.Sprintf("%s:%s.%s", reference.ProjectId, reference.Location, reference.JobId)
}

func formatJobReferenceForQueryParam(jobReference bigquery.JobReference) string {
	return fmt.Sprintf("bq:%s:%s", jobReference.Location, jobReference.JobId)
}

func (p Plugin) newBigQueryClient(ctx context.Context, identity google.Identity) (*bigquery.Service, error) {
	options := []option.ClientOption{
		option.WithScopes("https://www.googleapis.com/auth/bigquery"),
		// FIXME how do I access current version?
		option.WithUserAgent(fmt.Sprintf("%s/%s", "flytepropeller", "LATEST")),
	}

	// for mocking/testing purposes
	if p.cfg.bigQueryEndpoint != "" {
		options = append(options,
			option.WithEndpoint(p.cfg.bigQueryEndpoint),
			option.WithTokenSource(oauth2.StaticTokenSource(&oauth2.Token{})))
	} else if p.cfg.GoogleTokenSource.Type != "default" {

		tokenSource, err := p.googleTokenSource.GetTokenSource(ctx, identity)

		if err != nil {
			return nil, pluginErrors.Wrapf(pluginErrors.RuntimeFailure, err, "unable to get token source")
		}

		options = append(options, option.WithTokenSource(tokenSource))
	} else {
		logger.Infof(ctx, "BigQuery client read $GOOGLE_APPLICATION_CREDENTIALS by default")
	}

	return bigquery.NewService(ctx, options...)
}

func NewPlugin(cfg *Config, metricScope promutils.Scope) (*Plugin, error) {
	googleTokenSource, err := google.NewTokenSourceFactory(cfg.GoogleTokenSource)

	if err != nil {
		return nil, pluginErrors.Wrapf(pluginErrors.PluginInitializationFailed, err, "failed to get google token source")
	}

	return &Plugin{
		metricScope:       metricScope,
		cfg:               cfg,
		googleTokenSource: googleTokenSource,
	}, nil
}

func newBigQueryJobTaskPlugin() webapi.PluginEntry {
	return webapi.PluginEntry{
		ID:                 "bigquery",
		SupportedTaskTypes: []core.TaskType{bigqueryQueryJobTask},
		PluginLoader: func(ctx context.Context, iCtx webapi.PluginSetupContext) (webapi.AsyncPlugin, error) {
			cfg := GetConfig()

			return NewPlugin(cfg, iCtx.MetricsScope())
		},
	}
}

func init() {
	gob.Register(ResourceMetaWrapper{})
	gob.Register(ResourceWrapper{})

	pluginmachinery.PluginRegistry().RegisterRemotePlugin(newBigQueryJobTaskPlugin())
}
