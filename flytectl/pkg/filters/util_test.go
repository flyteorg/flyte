package filters

import (
	"testing"

	"github.com/flyteorg/flytectl/cmd/config"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/stretchr/testify/assert"
)

var (
	project = "flytesnack"
	domain  = "staging"
	name    = "test"
	output  = "json"
)

func TestListRequestWithoutNameFunc(t *testing.T) {
	config.GetConfig().Output = output
	config.GetConfig().Project = project
	config.GetConfig().Domain = domain
	filter := Filters{
		Limit:  100,
		SortBy: "created_at",
		Asc:    true,
	}
	request, err := BuildResourceListRequestWithName(filter, project, domain, "")
	expectedResponse := &admin.ResourceListRequest{
		Id: &admin.NamedEntityIdentifier{
			Project: project,
			Domain:  domain,
		},
		Limit: 100,
		SortBy: &admin.Sort{
			Key:       "created_at",
			Direction: admin.Sort_ASCENDING,
		},
		Filters: "",
	}
	assert.Nil(t, err)
	assert.Equal(t, expectedResponse, request)
}

func TestProjectListRequestFunc(t *testing.T) {
	config.GetConfig().Output = output
	config.GetConfig().Project = project
	config.GetConfig().Domain = domain
	filter := Filters{
		Limit:  100,
		SortBy: "created_at",
	}
	request, err := BuildProjectListRequest(filter)
	expectedResponse := &admin.ProjectListRequest{
		Limit:   100,
		Filters: "",
		SortBy: &admin.Sort{
			Key:       "created_at",
			Direction: admin.Sort_DESCENDING,
		},
	}
	assert.Nil(t, err)
	assert.Equal(t, expectedResponse, request)
}

func TestProjectListWithRequestFuncError(t *testing.T) {
	config.GetConfig().Output = output
	config.GetConfig().Project = project
	config.GetConfig().Domain = domain
	filter := Filters{
		FieldSelector: "Hello=",
		Limit:         100,
	}
	request, err := BuildProjectListRequest(filter)
	assert.NotNil(t, err)
	assert.Nil(t, request)
}

func TestListRequestWithNameFunc(t *testing.T) {
	config.GetConfig().Output = output
	filter := Filters{
		Limit:  100,
		SortBy: "created_at",
	}
	request, err := BuildResourceListRequestWithName(filter, project, domain, name)
	expectedResponse := &admin.ResourceListRequest{
		Id: &admin.NamedEntityIdentifier{
			Project: project,
			Domain:  domain,
			Name:    name,
		},
		Limit: 100,
		SortBy: &admin.Sort{
			Key:       "created_at",
			Direction: admin.Sort_DESCENDING,
		},
	}
	assert.Nil(t, err)
	assert.Equal(t, expectedResponse, request)
}

func TestListRequestWithNameFuncError(t *testing.T) {
	config.GetConfig().Output = output
	filter := Filters{
		Limit:         100,
		SortBy:        "created_at",
		FieldSelector: "hello=",
	}
	request, err := BuildResourceListRequestWithName(filter, project, domain, name)
	assert.NotNil(t, err)
	assert.Nil(t, request)
}
