package mocks

import (
	"github.com/aws/aws-sdk-go/service/cloudwatchevents"
	"github.com/flyteorg/flyteadmin/pkg/async/schedule/aws/interfaces"
)

type putRuleFunc func(input *cloudwatchevents.PutRuleInput) (*cloudwatchevents.PutRuleOutput, error)
type putTargetsFunc func(input *cloudwatchevents.PutTargetsInput) (*cloudwatchevents.PutTargetsOutput, error)
type deleteRuleFunc func(input *cloudwatchevents.DeleteRuleInput) (*cloudwatchevents.DeleteRuleOutput, error)
type removeTargetsFunc func(input *cloudwatchevents.RemoveTargetsInput) (*cloudwatchevents.RemoveTargetsOutput, error)

// A mock implementation of CloudWatchEventClient for use in tests.
type MockCloudWatchEventClient struct {
	putRule       putRuleFunc
	putTargets    putTargetsFunc
	deleteRule    deleteRuleFunc
	removeTargets removeTargetsFunc
}

func (c *MockCloudWatchEventClient) SetPutRuleFunc(putRule putRuleFunc) {
	c.putRule = putRule
}

func (c *MockCloudWatchEventClient) PutRule(input *cloudwatchevents.PutRuleInput) (
	*cloudwatchevents.PutRuleOutput, error) {
	if c.putRule != nil {
		return c.putRule(input)
	}
	return nil, nil
}

func (c *MockCloudWatchEventClient) SetPutTargetsFunc(putTargets putTargetsFunc) {
	c.putTargets = putTargets
}

func (c *MockCloudWatchEventClient) PutTargets(input *cloudwatchevents.PutTargetsInput) (
	*cloudwatchevents.PutTargetsOutput, error) {
	if c.putTargets != nil {
		return c.putTargets(input)
	}
	return nil, nil
}

func (c *MockCloudWatchEventClient) SetDeleteRuleFunc(deleteRule deleteRuleFunc) {
	c.deleteRule = deleteRule
}

func (c *MockCloudWatchEventClient) DeleteRule(input *cloudwatchevents.DeleteRuleInput) (
	*cloudwatchevents.DeleteRuleOutput, error) {
	if c.deleteRule != nil {
		return c.deleteRule(input)
	}
	return nil, nil
}

func (c *MockCloudWatchEventClient) SetRemoveTargetsFunc(removeTargets removeTargetsFunc) {
	c.removeTargets = removeTargets
}

func (c *MockCloudWatchEventClient) RemoveTargets(input *cloudwatchevents.RemoveTargetsInput) (
	*cloudwatchevents.RemoveTargetsOutput, error) {
	if c.removeTargets != nil {
		return c.removeTargets(input)
	}
	return nil, nil
}

func NewMockCloudWatchEventClient() interfaces.CloudWatchEventClient {
	return &MockCloudWatchEventClient{}
}
