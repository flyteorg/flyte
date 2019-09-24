package interfaces

import "github.com/aws/aws-sdk-go/service/cloudwatchevents"

// A subset of the AWS CloudWatchEvents service client.
type CloudWatchEventClient interface {
	PutRule(input *cloudwatchevents.PutRuleInput) (*cloudwatchevents.PutRuleOutput, error)
	PutTargets(input *cloudwatchevents.PutTargetsInput) (*cloudwatchevents.PutTargetsOutput, error)
	DeleteRule(input *cloudwatchevents.DeleteRuleInput) (*cloudwatchevents.DeleteRuleOutput, error)
	RemoveTargets(input *cloudwatchevents.RemoveTargetsInput) (*cloudwatchevents.RemoveTargetsOutput, error)
}
