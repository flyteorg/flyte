//go:build tools
// +build tools

package tools

import (
	_ "github.com/alvaroloes/enumer"
	_ "github.com/golangci/golangci-lint/v2/cmd/golangci-lint"
	_ "github.com/pseudomuto/protoc-gen-doc/cmd/protoc-gen-doc"
	_ "github.com/vektra/mockery/tools/cmd"

	_ "github.com/flyteorg/flyte/flytestdlib/cli/pflags"
)
