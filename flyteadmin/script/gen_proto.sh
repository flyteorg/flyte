#!/usr/bin/env bash

go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28.0

protoc --proto_path=pkg/repositories/models/protos --go_out=pkg/repositories/gen node_execution.proto