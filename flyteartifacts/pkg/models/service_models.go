package models

import "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/artifact"

// This is a wrapper object for easier handling of additional fields
type Artifact struct {
	artifact.Artifact
	OffloadedMetadata string
}
