package testutils

import "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"

func GetArtifactID() *core.ArtifactID {
	return &core.ArtifactID{
		ArtifactKey: &core.ArtifactKey{
			Project: "project",
			Domain:  "domain",
			Name:    "name",
		},
		Version: "artf_v",
	}
}
