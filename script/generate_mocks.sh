#!/bin/bash
set -e
set -x

mockery -dir=gen/pb-go/flyteidl/service/ -all -output=gen/clients/go/admin/mocks
mockery -dir=gen/pb-go/flyteidl/datacatalog/ -name=DataCatalogClient -output=gen/clients/go/datacatalog/mocks
