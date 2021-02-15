package get

import (
	"github.com/lyft/flytectl/pkg/printer"
)

var entityColumns = []printer.Column{
	{Header: "Domain", JSONPath: "$.domain"},
	{Header: "Name", JSONPath: "$.name"},
	{Header: "Project", JSONPath: "$.project"},
}
