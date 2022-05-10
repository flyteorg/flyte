//go:build integration
// +build integration

package tests

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/golang/protobuf/proto"

	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
)

var resourceTypes = []core.ResourceType{
	core.ResourceType_TASK,
	core.ResourceType_WORKFLOW,
	core.ResourceType_LAUNCH_PLAN,
}

var namedEntityIdentifier = admin.NamedEntityIdentifier{
	Project: "admintests",
	Domain:  "development",
	Name:    "workflow",
}

var namedEntityMetadata = admin.NamedEntityMetadata{
	Description: "description",
}

func getNamedEntityUpdateRequest() admin.NamedEntityUpdateRequest {
	return admin.NamedEntityUpdateRequest{
		ResourceType: core.ResourceType_WORKFLOW,
		Id: &admin.NamedEntityIdentifier{
			Project: "admintests",
			Domain:  "development",
			Name:    "workflow",
		},
		Metadata: &admin.NamedEntityMetadata{
			Description: "description",
		},
	}
}

func TestGetNamedEntities(t *testing.T) {
	truncateAllTablesForTestingOnly()
	client, conn := GetTestAdminServiceClient()
	defer conn.Close()
	insertTasksForTests(t, client)
	insertWorkflowsForTests(t, client)
	insertLaunchPlansForTests(t, client)

	t.Run("TestGetNamedEntityGrpc", testGetNamedEntityGrpc)
	t.Run("TestGetNamedEntityHTTP", testGetNamedEntityHTTP)
	t.Run("TestListNamedEntityGrpc", testListNamedEntityGrpc)
	t.Run("TestListNamedEntityHTTP", testListNamedEntityHTTP)
	t.Run("TestListNamedEntity_PaginationGrpc", testListNamedEntity_PaginationGrpc)
	t.Run("TestListNamedEntity_PaginationHTTP", testListNamedEntity_PaginationHTTP)
}

func testGetNamedEntityGrpc(t *testing.T) {
	ctx := context.Background()
	client, conn := GetTestAdminServiceClient()
	defer conn.Close()
	identifier := admin.NamedEntityIdentifier{
		Project: "admintests",
		Domain:  "development",
		Name:    "name_a",
	}
	for _, resourceType := range resourceTypes {
		entity, err := client.GetNamedEntity(ctx, &admin.NamedEntityGetRequest{
			ResourceType: resourceType,
			Id:           &identifier,
		})
		assert.NoError(t, err)
		assert.True(t, proto.Equal(&identifier, entity.Id))
	}
}

func testGetNamedEntityHTTP(t *testing.T) {
	for _, resourceType := range resourceTypes {
		url := fmt.Sprintf("%s/api/v1/named_entities/%d/admintests/development/name_a",
			GetTestHostEndpoint(),
			resourceType)
		getRequest, err := http.NewRequest("GET", url, nil)
		assert.Nil(t, err)
		addHTTPRequestHeaders(getRequest)

		httpClient := &http.Client{}
		resp, err := httpClient.Do(getRequest)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()

		body, _ := ioutil.ReadAll(resp.Body)
		octetStreamedNamedEntity := admin.NamedEntity{}
		proto.Unmarshal(body, &octetStreamedNamedEntity)
		assert.Equal(t, "admintests", octetStreamedNamedEntity.Id.GetProject())
		assert.Equal(t, "development", octetStreamedNamedEntity.Id.GetDomain())
		assert.Equal(t, "name_a", octetStreamedNamedEntity.Id.GetName())
		assert.Equal(t, resourceType, octetStreamedNamedEntity.GetResourceType())
	}

}

func testListNamedEntityGrpc(t *testing.T) {
	ctx := context.Background()
	client, conn := GetTestAdminServiceClient()
	defer conn.Close()
	for _, resourceType := range resourceTypes {
		result, err := client.ListNamedEntities(ctx, &admin.NamedEntityListRequest{
			ResourceType: resourceType,
			Project:      "admintests",
			Domain:       "development",
			Limit:        20,
			SortBy: &admin.Sort{
				Direction: admin.Sort_ASCENDING,
				Key:       "name",
			},
		})
		assert.NoError(t, err)
		assert.Len(t, result.Entities, 3)

		for i, entity := range result.Entities {
			assert.Equal(t, "admintests", entity.Id.Project)
			assert.Equal(t, "development", entity.Id.Domain)
			assert.Equal(t, resourceType, entity.ResourceType)
			assert.Equal(t, entityNames[i], entity.Id.Name)
		}
	}
}

func testListNamedEntityHTTP(t *testing.T) {
	for _, resourceType := range resourceTypes {
		url := fmt.Sprintf(
			"%s/api/v1/named_entities/%d/admintests/development?sort_by.direction=ASCENDING&sort_by.key=name&limit=20",
			GetTestHostEndpoint(), resourceType)
		getRequest, err := http.NewRequest("GET", url, nil)
		assert.Nil(t, err)
		addHTTPRequestHeaders(getRequest)

		httpClient := &http.Client{}
		resp, err := httpClient.Do(getRequest)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()

		body, _ := ioutil.ReadAll(resp.Body)
		octetStreamedNamedEntityList := admin.NamedEntityList{}
		proto.Unmarshal(body, &octetStreamedNamedEntityList)
		assert.Len(t, octetStreamedNamedEntityList.Entities, 3)

		for i, entity := range octetStreamedNamedEntityList.Entities {
			assert.Equal(t, "admintests", entity.Id.Project)
			assert.Equal(t, "development", entity.Id.Domain)
			assert.Equal(t, resourceType, entity.ResourceType)
			assert.Equal(t, entityNames[i], entity.Id.Name)
		}
	}
}

func testListNamedEntity_PaginationGrpc(t *testing.T) {
	ctx := context.Background()
	client, conn := GetTestAdminServiceClient()
	defer conn.Close()

	for _, resourceType := range resourceTypes {
		result, err := client.ListNamedEntities(ctx, &admin.NamedEntityListRequest{
			Project:      "admintests",
			Domain:       "development",
			ResourceType: resourceType,
			Limit:        2,
		})
		assert.NoError(t, err)
		assert.Len(t, result.Entities, 2)

		firstResponseNames := make([]string, 2)
		for idx, entity := range result.Entities {
			assert.Equal(t, "admintests", entity.Id.Project)
			assert.Equal(t, "development", entity.Id.Domain)
			assert.Contains(t, entityNames, entity.Id.Name)
			firstResponseNames[idx] = entity.Id.Name
		}

		result, err = client.ListNamedEntities(ctx, &admin.NamedEntityListRequest{
			Project:      "admintests",
			Domain:       "development",
			ResourceType: resourceType,
			Limit:        2,
			Token:        "2",
		})
		assert.NoError(t, err)
		assert.Len(t, result.Entities, 1)

		for _, entity := range result.Entities {
			assert.Equal(t, "admintests", entity.Id.Project)
			assert.Equal(t, "development", entity.Id.Domain)
			assert.Contains(t, entityNames, entity.Id.Name)
			assert.NotContains(t, firstResponseNames, entity.Id.Name)
		}
	}
}

func testListNamedEntity_PaginationHTTP(t *testing.T) {
	for _, resourceType := range resourceTypes {
		url := fmt.Sprintf(
			"%s/api/v1/named_entities/%d/admintests/development?sort_by.direction=ASCENDING&sort_by.key=name&limit=2",
			GetTestHostEndpoint(), resourceType)
		getRequest, err := http.NewRequest("GET", url, nil)
		assert.Nil(t, err)
		addHTTPRequestHeaders(getRequest)

		httpClient := &http.Client{}
		resp, err := httpClient.Do(getRequest)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()

		body, _ := ioutil.ReadAll(resp.Body)
		octetStreamedNamedEntityList := admin.NamedEntityList{}
		proto.Unmarshal(body, &octetStreamedNamedEntityList)
		assert.Len(t, octetStreamedNamedEntityList.Entities, 2)

		firstResponseNames := make([]string, 2)
		for idx, entity := range octetStreamedNamedEntityList.Entities {
			assert.Equal(t, "admintests", entity.Id.Project)
			assert.Equal(t, "development", entity.Id.Domain)
			assert.Contains(t, entityNames, entity.Id.Name)

			firstResponseNames[idx] = entity.Id.Name
		}

		url = fmt.Sprintf(
			"%s/api/v1/named_entities/%d/admintests/development?sort_by.direction=ASCENDING&sort_by.key=name&limit=2&token=%s",
			GetTestHostEndpoint(), resourceType, octetStreamedNamedEntityList.Token)
		getRequest, err = http.NewRequest("GET", url, nil)
		assert.Nil(t, err)
		addHTTPRequestHeaders(getRequest)
		resp, err = httpClient.Do(getRequest)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()

		body, _ = ioutil.ReadAll(resp.Body)
		octetStreamedNamedEntityList = admin.NamedEntityList{}
		proto.Unmarshal(body, &octetStreamedNamedEntityList)
		assert.Len(t, octetStreamedNamedEntityList.Entities, 1)

		for _, entity := range octetStreamedNamedEntityList.Entities {
			assert.Equal(t, "admintests", entity.Id.Project)
			assert.Equal(t, "development", entity.Id.Domain)
			assert.Contains(t, entityNames, entity.Id.Name)
			assert.NotContains(t, firstResponseNames, entity.Id.Name)
		}
	}
}

func TestUpdateNamedEntity(t *testing.T) {
	truncateAllTablesForTestingOnly()
	client, conn := GetTestAdminServiceClient()
	ctx := context.Background()

	defer conn.Close()
	insertTasksForTests(t, client)
	insertWorkflowsForTests(t, client)
	insertLaunchPlansForTests(t, client)

	identifier := admin.NamedEntityIdentifier{
		Project: "admintests",
		Domain:  "development",
		Name:    "name_a",
	}

	for _, resourceType := range resourceTypes {
		_, err := client.UpdateNamedEntity(ctx, &admin.NamedEntityUpdateRequest{
			ResourceType: resourceType,
			Id:           &identifier,
			Metadata: &admin.NamedEntityMetadata{
				Description: "updated description",
			},
		})
		assert.NoError(t, err)

		entity, err := client.GetNamedEntity(ctx, &admin.NamedEntityGetRequest{
			ResourceType: resourceType,
			Id:           &identifier,
		})
		assert.NoError(t, err)
		assert.Equal(t, "updated description", entity.Metadata.Description)
	}
}

func TestUpdateNamedEntityState(t *testing.T) {
	truncateAllTablesForTestingOnly()
	client, conn := GetTestAdminServiceClient()
	ctx := context.Background()

	defer conn.Close()
	insertTasksForTests(t, client)
	insertWorkflowsForTests(t, client)

	result, err := client.ListNamedEntities(ctx, &admin.NamedEntityListRequest{
		ResourceType: core.ResourceType_WORKFLOW,
		Project:      "admintests",
		Domain:       "development",
		Limit:        20,
		SortBy: &admin.Sort{
			Direction: admin.Sort_ASCENDING,
			Key:       "name",
		},
		Filters: fmt.Sprintf("eq(named_entity_metadata.state, %v)", int(admin.NamedEntityState_NAMED_ENTITY_ACTIVE)),
	})
	assert.NoError(t, err)
	assert.Len(t, result.Entities, 3)

	identifier := admin.NamedEntityIdentifier{
		Project: "admintests",
		Domain:  "development",
		Name:    "name_a",
	}
	_, err = client.UpdateNamedEntity(ctx, &admin.NamedEntityUpdateRequest{
		ResourceType: core.ResourceType_WORKFLOW,
		Id:           &identifier,
		Metadata: &admin.NamedEntityMetadata{
			Description: "updated description",
			State:       admin.NamedEntityState_NAMED_ENTITY_ARCHIVED,
		},
	})
	assert.NoError(t, err)

	result, err = client.ListNamedEntities(ctx, &admin.NamedEntityListRequest{
		ResourceType: core.ResourceType_WORKFLOW,
		Project:      "admintests",
		Domain:       "development",
		Limit:        20,
		SortBy: &admin.Sort{
			Direction: admin.Sort_ASCENDING,
			Key:       "name",
		},
		Filters: fmt.Sprintf("eq(named_entity_metadata.state, %v)", int(admin.NamedEntityState_NAMED_ENTITY_ACTIVE)),
	})
	assert.NoError(t, err)
	assert.Len(t, result.Entities, 2)

	result, err = client.ListNamedEntities(ctx, &admin.NamedEntityListRequest{
		ResourceType: core.ResourceType_WORKFLOW,
		Project:      "admintests",
		Domain:       "development",
		Limit:        20,
		SortBy: &admin.Sort{
			Direction: admin.Sort_ASCENDING,
			Key:       "name",
		},
		Filters: fmt.Sprintf("eq(named_entity_metadata.state, %v)", int(admin.NamedEntityState_NAMED_ENTITY_ARCHIVED)),
	})
	assert.NoError(t, err)
	assert.Len(t, result.Entities, 1)
}
