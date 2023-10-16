package models

import (
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"time"
)

type Artifact struct {
	ID        uint
	CreatedAt time.Time
	UpdatedAt time.Time
	Principal string

	Project      string
	Domain       string
	Name         string
	Version      string
	Partitions   map[string]string
	LiteralType  core.LiteralType
	LiteralValue core.Literal

	Description           string
	MetadataType          string
	OffloadedUserMetadata string
	UserMetadata          []byte

	// Source information
	ExecutionName   string
	WorkflowProject string
	WorkflowDomain  string
	WorkflowName    string
	WorkflowVersion string
	NodeID          string
	TaskProject     string
	TaskDomain      string
	TaskName        string
	TaskVersion     string
	RetryAttempt    uint32
}
