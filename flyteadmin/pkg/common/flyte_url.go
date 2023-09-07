package common

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
)

// transform to snake case to make lower case
//go:generate enumer --type=ArtifactType --trimprefix=ArtifactType -transform=snake

type ArtifactType int

// The suffixes in these constants are used to match against the tail end of the flyte url, to keep tne flyte url simpler
const (
	ArtifactTypeUndefined ArtifactType = iota
	ArtifactTypeI                      // inputs
	ArtifactTypeO                      // outputs
	ArtifactTypeD                      // deck
)

var re = regexp.MustCompile("flyte://v1/(?P<project>[a-zA-Z0-9_-]+)/(?P<domain>[a-zA-Z0-9_-]+)/(?P<exec>[a-zA-Z0-9_-]+)/(?P<node>[a-zA-Z0-9_-]+)(?:/(?P<attempt>[0-9]+))?/(?P<ioType>[iod])$")

var reSpecificOutput = regexp.MustCompile("flyte://v1/(?P<project>[a-zA-Z0-9_-]+)/(?P<domain>[a-zA-Z0-9_-]+)/(?P<exec>[a-zA-Z0-9_-]+)/(?P<node>[a-zA-Z0-9_-]+)(?:/(?P<attempt>[0-9]+))?/(?P<ioType>[io])/(?P<literalName>[a-zA-Z0-9_-]+)$")

func MatchRegex(reg *regexp.Regexp, input string) map[string]string {
	names := reg.SubexpNames()
	res := reg.FindAllStringSubmatch(input, -1)
	if len(res) == 0 {
		return nil
	}
	dict := make(map[string]string, len(names))
	for i := 1; i < len(res[0]); i++ {
		dict[names[i]] = res[0][i]
	}
	return dict
}

type ParsedExecution struct {
	// Returned when the user does not request a specific attempt
	NodeExecID *core.NodeExecutionIdentifier

	// This is returned in the case where the user requested a specific attempt. But the TaskID portion of this
	// will be empty since that information is not encapsulated in the flyte url.
	PartialTaskExecID *core.TaskExecutionIdentifier

	// The name of the input or output in the literal map
	LiteralName string

	IOType ArtifactType
}

func tryMatches(flyteURL string) map[string]string {
	var matches map[string]string

	if matches = MatchRegex(re, flyteURL); len(matches) > 0 {
		return matches
	} else if matches = MatchRegex(reSpecificOutput, flyteURL); len(matches) > 0 {
		return matches
	}
	return nil
}

func ParseFlyteURLToExecution(flyteURL string) (ParsedExecution, error) {
	// flyteURL can be of the following forms
	//  flyte://v1/project/domain/execution_id/node_id/attempt/[iod]
	//  flyte://v1/project/domain/execution_id/node_id/attempt/[io]/output_name
	//  flyte://v1/project/domain/execution_id/node_id/[io]/output_name
	//  flyte://v1/project/domain/execution_id/node_id/[iod]

	// where i stands for inputs.pb o for outputs.pb and d for the flyte deck
	// If the retry attempt is missing, the io requested is assumed to be for the node instead of the task execution

	matches := tryMatches(flyteURL)
	if matches == nil {
		return ParsedExecution{}, fmt.Errorf("failed to parse [%s]", flyteURL)
	}

	proj := matches["project"]
	domain := matches["domain"]
	executionID := matches["exec"]
	nodeID := matches["node"]
	ioType, err := ArtifactTypeString(matches["ioType"])
	if err != nil {
		return ParsedExecution{}, err
	}
	literalName := matches["literalName"] // may be empty

	// node and task level outputs
	nodeExecID := core.NodeExecutionIdentifier{
		NodeId: nodeID,
		ExecutionId: &core.WorkflowExecutionIdentifier{
			Project: proj,
			Domain:  domain,
			Name:    executionID,
		},
	}

	// if attempt is there, that means a task execution
	if attempt := matches["attempt"]; len(attempt) > 0 {
		a, err := strconv.Atoi(attempt)
		if err != nil {
			return ParsedExecution{}, fmt.Errorf("failed to parse attempt [%v], %v", attempt, err)
		}
		taskExecID := core.TaskExecutionIdentifier{
			NodeExecutionId: &nodeExecID,
			// checking for overflow here is probably unreasonable
			RetryAttempt: uint32(a),
		}
		return ParsedExecution{
			PartialTaskExecID: &taskExecID,
			IOType:            ioType,
			LiteralName:       literalName,
		}, nil
	}

	return ParsedExecution{
		NodeExecID:  &nodeExecID,
		IOType:      ioType,
		LiteralName: literalName,
	}, nil

}

func FlyteURLsFromNodeExecutionID(nodeExecutionID core.NodeExecutionIdentifier, deck bool) *admin.FlyteURLs {
	base := fmt.Sprintf("flyte://v1/%s/%s/%s/%s", nodeExecutionID.ExecutionId.Project,
		nodeExecutionID.ExecutionId.Domain, nodeExecutionID.ExecutionId.Name, nodeExecutionID.NodeId)

	res := &admin.FlyteURLs{
		Inputs:  fmt.Sprintf("%s/%s", base, ArtifactTypeI),
		Outputs: fmt.Sprintf("%s/%s", base, ArtifactTypeO),
	}
	if deck {
		res.Deck = fmt.Sprintf("%s/%s", base, ArtifactTypeD)
	}
	return res
}

func FlyteURLsFromTaskExecutionID(taskExecutionID core.TaskExecutionIdentifier, deck bool) *admin.FlyteURLs {
	base := fmt.Sprintf("flyte://v1/%s/%s/%s/%s/%s", taskExecutionID.NodeExecutionId.ExecutionId.Project,
		taskExecutionID.NodeExecutionId.ExecutionId.Domain, taskExecutionID.NodeExecutionId.ExecutionId.Name, taskExecutionID.NodeExecutionId.NodeId, strconv.Itoa(int(taskExecutionID.RetryAttempt)))

	res := &admin.FlyteURLs{
		Inputs:  fmt.Sprintf("%s/%s", base, ArtifactTypeI),
		Outputs: fmt.Sprintf("%s/%s", base, ArtifactTypeO),
	}
	if deck {
		res.Deck = fmt.Sprintf("%s/%s", base, ArtifactTypeD)
	}
	return res
}
