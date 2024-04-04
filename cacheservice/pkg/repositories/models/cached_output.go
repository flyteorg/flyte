package models

import (
	"time"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
)

type BaseModel struct {
	ID        string `gorm:"primary_key" dynamodbav:"id"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `gorm:"index"`
}

type Identifier struct {
	ResourceType core.ResourceType `dynamodbav:"resourceType"`
	Project      string            `dynamodbav:"project"`
	Domain       string            `dynamodbav:"domain"`
	Name         string            `dynamodbav:"name"`
	Version      string            `dynamodbav:"version"`
	Org          string            `dynamodbav:"org"`
}

type CachedOutput struct {
	BaseModel
	OutputURI          string `dynamodbav:"outputUri"`
	OutputLiteral      []byte `dynamodbav:"outputLiteral"`
	SerializedMetadata []byte `dynamodbav:"keyMap"`
	Identifier
}
