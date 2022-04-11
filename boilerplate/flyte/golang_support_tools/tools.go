//go:build tools
// +build tools

package tools

import (
	_ "github.com/EngHabu/mockery/cmd/mockery"
	_ "github.com/alvaroloes/enumer"
	_ "github.com/flyteorg/flytestdlib/cli/pflags"
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint"
	_ "github.com/pseudomuto/protoc-gen-doc/cmd/protoc-gen-doc"
)
