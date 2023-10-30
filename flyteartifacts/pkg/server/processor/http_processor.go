package processor

import "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/artifact"

// HTTPCloudEventProcessor could potentially run another local http service to handle traffic instead of a go channel
// todo: either implement or remote this file.
type HTTPCloudEventProcessor struct {
	service *artifact.UnimplementedArtifactRegistryServer
}
