/*
 * Copyright (c) 2018 Lyft. All rights reserved.
 */

package mocks

import (
	"context"
	"encoding/base32"
	"hash/fnv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"

	"github.com/aws/aws-sdk-go/service/batch"
)

type MockAwsBatchClient struct {
	jobs                               map[string]*batch.JobDetail
	SubmitJobWithContextCb             func(ctx context.Context, input *batch.SubmitJobInput, opts ...request.Option) (*batch.SubmitJobOutput, error)
	TerminateJobWithContextCb          func(ctx context.Context, input *batch.TerminateJobInput, opts ...request.Option) (*batch.TerminateJobOutput, error)
	DescribeJobsWithContextCb          func(ctx context.Context, input *batch.DescribeJobsInput, opts ...request.Option) (*batch.DescribeJobsOutput, error)
	RegisterJobDefinitionWithContextCb func(ctx context.Context, input *batch.RegisterJobDefinitionInput, opts ...request.Option) (*batch.RegisterJobDefinitionOutput, error)
}

const specialEncoderKey = "abcdefghijklmnopqrstuvwxyz123456"

var base32Encoder = base32.NewEncoding(specialEncoderKey).WithPadding(base32.NoPadding)

func (m *MockAwsBatchClient) SubmitJobWithContext(ctx aws.Context, input *batch.SubmitJobInput, opts ...request.Option) (
	*batch.SubmitJobOutput, error) {

	m.jobs[*input.JobName] = &batch.JobDetail{
		JobName: input.JobName,
		JobId:   input.JobName,
		Status:  ref(batch.JobStatusSubmitted),
	}

	if m.SubmitJobWithContextCb != nil {
		return m.SubmitJobWithContextCb(ctx, input, opts...)
	}

	hasher := fnv.New32a()
	_, err := hasher.Write([]byte(*input.JobName))
	if err != nil {
		return nil, err
	}

	b := hasher.Sum(nil)
	return &batch.SubmitJobOutput{
		JobId:   ref(base32Encoder.EncodeToString(b)),
		JobName: input.JobName,
	}, nil
}

func (m *MockAwsBatchClient) TerminateJobWithContext(ctx aws.Context, input *batch.TerminateJobInput,
	opts ...request.Option) (*batch.TerminateJobOutput, error) {

	if m.TerminateJobWithContextCb != nil {
		return m.TerminateJobWithContextCb(ctx, input, opts...)
	}

	return &batch.TerminateJobOutput{}, nil
}

func (m *MockAwsBatchClient) DescribeJobsWithContext(ctx aws.Context, input *batch.DescribeJobsInput,
	opts ...request.Option) (*batch.DescribeJobsOutput, error) {

	if m.DescribeJobsWithContextCb != nil {
		return m.DescribeJobsWithContextCb(ctx, input, opts...)
	}

	res := make([]*batch.JobDetail, 0, len(input.Jobs))
	for _, idRef := range input.Jobs {
		id := *idRef
		if currentDetails, found := m.jobs[id]; found {
			res = append(res, currentDetails)
			currentDetails.Status = ref(nextStatus(*currentDetails.Status))
			m.jobs[id] = currentDetails
		} else {
			currentDetails = &batch.JobDetail{
				JobName: ref(id),
				JobId:   ref(id),
				Status:  ref(batch.JobStatusSubmitted),
			}

			m.jobs[id] = currentDetails
			res = append(res, currentDetails)
		}
	}

	return &batch.DescribeJobsOutput{Jobs: res}, nil
}

func (m *MockAwsBatchClient) RegisterJobDefinitionWithContext(ctx aws.Context, input *batch.RegisterJobDefinitionInput,
	opts ...request.Option) (*batch.RegisterJobDefinitionOutput, error) {

	if m.RegisterJobDefinitionWithContextCb != nil {
		return m.RegisterJobDefinitionWithContextCb(ctx, input, opts...)
	}

	return &batch.RegisterJobDefinitionOutput{
		JobDefinitionArn: ref("my-arn"),
	}, nil
}

func ref(s string) *string {
	return &s
}

func nextStatus(s string) string {
	switch s {
	case batch.JobStatusSubmitted:
		return batch.JobStatusPending
	case batch.JobStatusPending:
		return batch.JobStatusRunnable
	case batch.JobStatusRunnable:
		return batch.JobStatusStarting
	case batch.JobStatusStarting:
		return batch.JobStatusRunning
	case batch.JobStatusRunning:
		return batch.JobStatusSucceeded
	case batch.JobStatusFailed:
		return batch.JobStatusFailed
	case batch.JobStatusSucceeded:
		return batch.JobStatusSucceeded
	default:
		panic("Unrecognized status")
	}
}

// Creates a new client with overridable behavior for certain APIs. The default implementation will progress submitted jobs
// until completion every time GetJobDetails is called.
func NewMockAwsBatchClient() *MockAwsBatchClient {
	return &MockAwsBatchClient{
		jobs: map[string]*batch.JobDetail{},
	}
}
