package awsbatch

import (
	"github.com/flyteorg/flyteplugins/go/tasks/plugins/array/awsbatch/definition"
	"github.com/flyteorg/flyteplugins/go/tasks/plugins/array/core"
)

type State struct {
	*core.State

	ExternalJobID    *string `json:"externalJobID"`
	JobDefinitionArn definition.JobDefinitionArn
}

func (s State) GetJobDefinitionArn() definition.JobDefinitionArn {
	return s.JobDefinitionArn
}

func (s State) GetExternalJobID() *string {
	return s.ExternalJobID
}

func (s *State) SetJobDefinitionArn(arn definition.JobDefinitionArn) *State {
	s.JobDefinitionArn = arn
	return s
}

func (s *State) SetExternalJobID(jobID string) *State {
	s.ExternalJobID = &jobID
	return s
}
