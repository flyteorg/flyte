#!/bin/bash
set -e
set -x

mockery -dir=gen/pb-go/flyteidl/service/ -name=AdminServiceClient -output=clients/go/admin/mocks
mockery -dir=gen/pb-go/flyteidl/datacatalog/ -name=DataCatalogClient -output=clients/go/datacatalog/mocks
