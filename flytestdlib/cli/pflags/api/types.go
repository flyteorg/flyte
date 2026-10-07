package api

import (
	"go/types"
	"time"
)

// TestStrategy determines how tests should be generated.
type TestStrategy string

const (
	JSON        TestStrategy = "Json"
	SliceJoined TestStrategy = "SliceJoined"
	Raw         TestStrategy = "Raw"
)

type FieldInfo struct {
	Name               string
	GoName             string
	Typ                types.Type
	LocalTypeName      string
	DefaultValue       string
	UsageString        string
	FlagMethodName     string
	TestFlagMethodName string
	TestValue          string
	TestStrategy       TestStrategy
	ShouldBindDefault  bool
	ShouldTestDefault  bool
}

// TypeInfo holds the finalized information passed to the template for evaluation.
type TypeInfo struct {
	Timestamp       time.Time
	Fields          []FieldInfo
	PFlagValueTypes []PFlagValueType
	Package         string
	Name            string
	TypeRef         string
	Imports         map[string]string
}

type PFlagValueType struct {
	Name                     string
	ShouldGenerateSetAndType bool
}
