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

var re = regexp.MustCompile("flyte://v1/(?P<project>[a-zA-Z0-9_-]+)/(?P<domain>[a-zA-Z0-9_-]+)/(?P<exec>[a-zA-Z0-9_-]+)/(?P<node>[a-zA-Z0-9_-]+)(?:/(?P<attempt>[0-9]+))?/(?P<artifactType>[iod])$")

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

func ParseFlyteURL(flyteURL string) (core.NodeExecutionIdentifier, *int, ArtifactType, error) {
	// flyteURL is of the form flyte://v1/project/domain/execution_id/node_id/attempt/[iod]
	// where i stands for inputs.pb o for outputs.pb and d for the flyte deck
	// If the retry attempt is missing, the io requested is assumed to be for the node instead of the task execution
	matches := MatchRegex(re, flyteURL)
	proj := matches["project"]
	domain := matches["domain"]
	executionID := matches["exec"]
	nodeID := matches["node"]
	var attemptPtr *int // nil means node execution, not a task execution
	if attempt := matches["attempt"]; len(attempt) > 0 {
		a, err := strconv.Atoi(attempt)
		if err != nil {
			return core.NodeExecutionIdentifier{}, nil, ArtifactTypeUndefined, fmt.Errorf("failed to parse attempt [%v], %v", attempt, err)
		}
		attemptPtr = &a
	}
	ioType, err := ArtifactTypeString(matches["artifactType"])
	if err != nil {
		return core.NodeExecutionIdentifier{}, nil, ArtifactTypeUndefined, err
	}

	return core.NodeExecutionIdentifier{
		NodeId: nodeID,
		ExecutionId: &core.WorkflowExecutionIdentifier{
			Project: proj,
			Domain:  domain,
			Name:    executionID,
		},
	}, attemptPtr, ioType, nil
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
