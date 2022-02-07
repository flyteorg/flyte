/*
 * Copyright (c) 2018 Lyft. All rights reserved.
 */

// This package deals with the communication with AWS-Batch and adopting its APIs to the flyte-plugin model.
package awsbatch

import (
	"context"
	"fmt"

	a "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/batch"
	"github.com/flyteorg/flyteplugins/go/tasks/aws"
	definition2 "github.com/flyteorg/flyteplugins/go/tasks/plugins/array/awsbatch/definition"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/utils"
)

//go:generate mockery -all -case=underscore

// AWS Batch Client interface.
type Client interface {
	// Submits a new job to AWS Batch and retrieves job info. Note that submitted jobs will not have status populated.
	SubmitJob(ctx context.Context, input *batch.SubmitJobInput) (jobID string, err error)

	// Attempts to terminate a job. If the job hasn't started yet, it'll just get deleted.
	TerminateJob(ctx context.Context, jobID JobID, reason string) error

	// Retrieves jobs' details from AWS Batch.
	GetJobDetailsBatch(ctx context.Context, ids []JobID) ([]*batch.JobDetail, error)

	// Registers a new Job Definition with AWS Batch provided a name, image and role.
	RegisterJobDefinition(ctx context.Context, name, image, role string, platformCapabilities string) (arn string, err error)

	// Gets the single region this client interacts with.
	GetRegion() string

	GetAccountID() string
}

// BatchServiceClient is an interface on top of the native AWS Batch client to allow for mocking and alternative implementations.
type BatchServiceClient interface {
	SubmitJobWithContext(ctx a.Context, input *batch.SubmitJobInput, opts ...request.Option) (*batch.SubmitJobOutput, error)
	TerminateJobWithContext(ctx a.Context, input *batch.TerminateJobInput, opts ...request.Option) (*batch.TerminateJobOutput, error)
	DescribeJobsWithContext(ctx a.Context, input *batch.DescribeJobsInput, opts ...request.Option) (*batch.DescribeJobsOutput, error)
	RegisterJobDefinitionWithContext(ctx a.Context, input *batch.RegisterJobDefinitionInput, opts ...request.Option) (*batch.RegisterJobDefinitionOutput, error)
}

type client struct {
	Batch              BatchServiceClient
	getRateLimiter     utils.RateLimiter
	defaultRateLimiter utils.RateLimiter
	region             string
	accountID          string
}

func (b client) GetRegion() string {
	return b.region
}

func (b client) GetAccountID() string {
	return b.accountID
}

// Registers a new job definition. There is no deduping on AWS side (even for the same name).
func (b *client) RegisterJobDefinition(ctx context.Context, name, image, role string, platformCapabilities string) (arn definition2.JobDefinitionArn, err error) {
	logger.Infof(ctx, "Registering job definition with name [%v], image [%v], role [%v], platformCapabilities [%v]", name, image, role, platformCapabilities)

	res, err := b.Batch.RegisterJobDefinitionWithContext(ctx, &batch.RegisterJobDefinitionInput{
		Type:                 refStr(batch.JobDefinitionTypeContainer),
		JobDefinitionName:    refStr(name),
		PlatformCapabilities: refStrSlice([]string{platformCapabilities}),
		ContainerProperties: &batch.ContainerProperties{
			Image:      refStr(image),
			JobRoleArn: refStr(role),

			// These will be overwritten on execution
			Vcpus:  refInt(1),
			Memory: refInt(100),
		},
	})
	if err != nil {
		return "", err
	}

	return *res.JobDefinitionArn, nil
}

// Submits a new job to a desired queue
func (b *client) SubmitJob(ctx context.Context, input *batch.SubmitJobInput) (jobID string, err error) {
	if input == nil {
		return "", nil
	}

	if err := b.defaultRateLimiter.Wait(ctx); err != nil {
		return "", err
	}

	output, err := b.Batch.SubmitJobWithContext(ctx, input)
	if err != nil {
		return "", err
	}

	if output.JobId == nil {
		logger.Errorf(ctx, "Job submitted has no ID and no error is returned. This is an AWS-issue. Input [%v]", input.JobName)
		return "", fmt.Errorf("job submitted has no ID and no error is returned. This is an AWS-issue. Input [%v]", input.JobName)
	}

	return *output.JobId, nil
}

// Terminates an in progress job
func (b *client) TerminateJob(ctx context.Context, jobID JobID, reason string) error {
	if err := b.defaultRateLimiter.Wait(ctx); err != nil {
		return err
	}

	input := batch.TerminateJobInput{
		JobId:  refStr(jobID),
		Reason: refStr(reason),
	}

	if _, err := b.Batch.TerminateJobWithContext(ctx, &input); err != nil {
		return err
	}

	return nil
}

func (b *client) GetJobDetailsBatch(ctx context.Context, jobIds []JobID) ([]*batch.JobDetail, error) {
	if err := b.getRateLimiter.Wait(ctx); err != nil {
		return nil, err
	}

	ids := make([]*string, 0, len(jobIds))
	for _, id := range jobIds {
		ids = append(ids, refStr(id))
	}

	input := batch.DescribeJobsInput{
		Jobs: ids,
	}

	output, err := b.Batch.DescribeJobsWithContext(ctx, &input)
	if err != nil {
		return nil, err
	}

	return output.Jobs, nil
}

// Initializes a new Batch Client that can be used to interact with AWS Batch.
func NewBatchClient(awsClient aws.Client,
	getRateLimiter utils.RateLimiter,
	defaultRateLimiter utils.RateLimiter) Client {

	batchClient := batch.New(awsClient.GetSession(), awsClient.GetSdkConfig())
	return NewCustomBatchClient(batchClient, awsClient.GetConfig().AccountID, batchClient.SigningRegion,
		getRateLimiter, defaultRateLimiter)
}

func NewCustomBatchClient(batchClient BatchServiceClient, accountID, region string,
	getRateLimiter utils.RateLimiter,
	defaultRateLimiter utils.RateLimiter) Client {
	return &client{
		Batch:              batchClient,
		accountID:          accountID,
		region:             region,
		getRateLimiter:     getRateLimiter,
		defaultRateLimiter: defaultRateLimiter,
	}
}
