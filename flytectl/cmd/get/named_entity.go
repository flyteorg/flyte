package get

import (
	"github.com/flyteorg/flytectl/pkg/printer"
)

var entityColumns = []printer.Column{
	{Header: "Domain", JSONPath: "$.domain"},
	{Header: "Name", JSONPath: "$.name"},
	{Header: "Project", JSONPath: "$.project"},
}
