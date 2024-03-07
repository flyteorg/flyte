package models

import "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"

type Identifier struct {
	ResourceType core.ResourceType `dynamodbav:"resourceType"`
	Project      string            `dynamodbav:"project"`
	Domain       string            `dynamodbav:"domain"`
	Name         string            `dynamodbav:"name"`
	Version      string            `dynamodbav:"version"`
	Org          string            `dynamodbav:"org"`
}

type CachedOutput struct {
	ID                 string     `dynamodbav:"id"`
	OutputURI          string     `dynamodbav:"outputUri"`
	OutputLiteral      []byte     `dynamodbav:"outputLiteral"`
	SerializedMetadata []byte     `dynamodbav:"keyMap"`
	Identifier         Identifier `dynamodbav:"identifier"`
}
