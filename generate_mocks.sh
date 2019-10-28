#!/bin/bash
set -e
set -x
which mockery || (go install github.com/lyft/flyteidl/vendor/github.com/vektra/mockery/cmd/mockery)

mockery -dir=gen/pb-go/flyteidl/service/ -name=AdminServiceClient -output=clients/go/admin/mocks
mockery -dir=gen/pb-go/flyteidl/datacatalog/ -name=ArtifactsClient -output=clients/go/datacatalog/mocks
