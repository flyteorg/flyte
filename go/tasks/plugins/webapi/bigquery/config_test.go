package bigquery

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestUnmarshalBigQueryQueryConfig(t *testing.T) {
	custom := structpb.Struct{
		Fields: map[string]*structpb.Value{
			"projectId": structpb.NewStringValue("project-id"),
			"location":  structpb.NewStringValue("EU"),
			"query":     structpb.NewStringValue("SELECT 1"),
		},
	}

	config, err := unmarshalQueryJobConfig(&custom)

	assert.NoError(t, err)

	assert.Equal(t, config, &QueryJobConfig{
		ProjectID: "project-id",
		Location:  "EU",
		Query:     "SELECT 1",
	})
}
