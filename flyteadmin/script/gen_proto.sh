#!/usr/bin/env bash

go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28.0

protoc --proto_path=flyteadmin/pkg/repositories/models/protos --go_out=flyteadmin/pkg/repositories/gen flyteadmin/pkg/repositories/models/protos/node_execution.proto
