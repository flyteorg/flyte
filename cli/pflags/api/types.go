package api

import (
	"go/types"
	"time"
)

// Determines how tests should be generated.
type TestStrategy string

const (
	JSON        TestStrategy = "Json"
	SliceJoined TestStrategy = "SliceJoined"
	SliceRaw    TestStrategy = "SliceRaw"
)

type FieldInfo struct {
	Name           string
	GoName         string
	Typ            types.Type
	DefaultValue   string
	UsageString    string
	FlagMethodName string
	TestValue      string
	TestStrategy   TestStrategy
}

// Holds the finalized information passed to the template for evaluation.
type TypeInfo struct {
	Timestamp time.Time
	Fields    []FieldInfo
	Package   string
	Name      string
	TypeRef   string
	Imports   map[string]string
}
