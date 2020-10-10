package get

import (
	"github.com/lyft/flytectl/printer"
)

var entityColumns = []printer.Column{
	{"Domain", "$.domain"},
	{"Name", "$.name"},
	{"Project", "$.project"},
}
