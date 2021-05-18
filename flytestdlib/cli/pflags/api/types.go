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
	Raw         TestStrategy = "Raw"
)

type FieldInfo struct {
	Name              string
	GoName            string
	Typ               types.Type
	DefaultValue      string
	UsageString       string
	FlagMethodName    string
	TestValue         string
	TestStrategy      TestStrategy
	ShouldBindDefault bool
	ShouldTestDefault bool
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
