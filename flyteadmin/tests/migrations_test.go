//go:build migration_test && integration
// +build migration_test,integration

package tests

import (
	"context"
	"fmt"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
	"sort"
	"testing"
)

// This file is used to test the migration from resource to configuration and configuration to resource.
// It is not included in the integration tests because it needs to run under different admin settings.

// To test resource to configuration migration and configuration to resource migration, we need to follow the following steps:
// (resource to configuration migration)
// 1: Run TestMigration0
// 2: Start admin with resource mode
// 3: Run TestMigration1
// 4: Start admin with configuration mode (which will run the migration)
// 5: Run TestMigration2
// (configuration to resource migration)
// 1: Run TestMigration0
// 2: Start admin with configuration mode
// 3: Run TestMigration1
// 4: Start admin with resource mode (which will run the migration)
// 5: Run TestMigration2

func TestMigration0(t *testing.T) {
	ctx := context.Background()
	db, err := repositories.GetDB(ctx, getDbConfig(), getLoggerConfig())
	assert.Nil(t, err)
	db.Exec("DELETE FROM migrations WHERE id = 'pg-continue-2024-04-resources-to-configurations'")
	db.Exec("DELETE FROM migrations WHERE id = 'pg-continue-2024-04-configurations-to-resources'")
	truncateTableForTesting(db, "resources")
	truncateTableForTesting(db, "configuration_document_metadata")
	numberOfProjects := 10
	for i := 0; i < numberOfProjects; i++ {
		project := fmt.Sprintf("test-project-%d", i)
		db.Exec("INSERT INTO projects (id, created_at, updated_at, identifier, org, name, state) VALUES (?, NOW(), NOW(), ?, ?, ?, ?)", i, project, "", project, 0)
	}
	sqlDB, err := db.DB()
	assert.Nil(t, err)
	err = sqlDB.Close()
	assert.Nil(t, err)
}

func TestMigration1(t *testing.T) {
	ctx := context.Background()
	numberOfProjects := 10
	domains := []string{"development", "staging", "production"}
	workflows := []string{"workflow-1", "workflow-2", "workflow-3"}

	client, conn := GetTestAdminServiceClient()
	defer conn.Close()

	for i := 0; i < numberOfProjects; i++ {
		project := fmt.Sprintf("test-project-%d", i)
		for _, domain := range domains {
			for _, workflow := range workflows {
				_, err := client.UpdateWorkflowAttributes(ctx, &admin.WorkflowAttributesUpdateRequest{
					Attributes: &admin.WorkflowAttributes{
						Project:            project,
						Domain:             domain,
						Workflow:           workflow,
						MatchingAttributes: matchingTaskResourceAttributes,
					},
				})
				assert.Nil(t, err)
				_, err = client.UpdateWorkflowAttributes(ctx, &admin.WorkflowAttributesUpdateRequest{
					Attributes: &admin.WorkflowAttributes{
						Project:            project,
						Domain:             domain,
						Workflow:           workflow,
						MatchingAttributes: matchingExecutionQueueAttributes,
					},
				})
				assert.Nil(t, err)
			}
			_, err := client.UpdateProjectDomainAttributes(ctx, &admin.ProjectDomainAttributesUpdateRequest{
				Attributes: &admin.ProjectDomainAttributes{
					Project:            project,
					Domain:             domain,
					MatchingAttributes: matchingTaskResourceAttributes,
				},
			})
			assert.Nil(t, err)
		}
		_, err := client.UpdateProjectAttributes(ctx, &admin.ProjectAttributesUpdateRequest{
			Attributes: &admin.ProjectAttributes{
				Project:            project,
				MatchingAttributes: matchingExecutionQueueAttributes,
			},
		})
		assert.Nil(t, err)
	}
}

func TestMigration2(t *testing.T) {
	client, conn := GetTestAdminServiceClient()
	defer conn.Close()
	ctx := context.Background()

	domains := []string{"development", "production", "staging"}
	workflows := []string{"workflow-1", "workflow-2", "workflow-3"}
	// sort the response
	var countLevel = func(item *admin.MatchableAttributesConfiguration) int {
		count := 0
		if item.Project != "" {
			count++
		}
		if item.Domain != "" {
			count++
		}
		if item.Workflow != "" {
			count++
		}
		return count
	}
	var sortAttributes = func(items []*admin.MatchableAttributesConfiguration) {
		sort.Slice(items, func(i, j int) bool {
			a, b := items[i], items[j]
			if countLevel(a) != countLevel(b) {
				return countLevel(a) > countLevel(b)
			}
			if a.Project != b.Project {
				return a.Project < b.Project
			}
			if a.Domain != b.Domain {
				return a.Domain < b.Domain
			}
			return a.Workflow < b.Workflow
		})
	}
	response, err := client.ListMatchableAttributes(ctx, &admin.ListMatchableAttributesRequest{
		ResourceType: admin.MatchableResource_TASK_RESOURCE,
	})
	assert.Nil(t, err)
	assert.Len(t, response.Configurations, 120)
	sortAttributes(response.Configurations)
	for i, item := range response.Configurations {
		if i < 90 {
			assert.True(t, proto.Equal(item, &admin.MatchableAttributesConfiguration{
				Project:    fmt.Sprintf("test-project-%d", i/9),
				Domain:     domains[(i/3)%3],
				Workflow:   workflows[i%3],
				Attributes: matchingTaskResourceAttributes,
			}))
		} else {
			i := i - 90
			assert.True(t, proto.Equal(item, &admin.MatchableAttributesConfiguration{
				Project:    fmt.Sprintf("test-project-%d", i/3),
				Domain:     domains[i%3],
				Attributes: matchingTaskResourceAttributes,
			}))
		}
	}

	response, err = client.ListMatchableAttributes(ctx, &admin.ListMatchableAttributesRequest{
		ResourceType: admin.MatchableResource_EXECUTION_QUEUE,
	})
	assert.Nil(t, err)
	assert.Len(t, response.Configurations, 100)
	sortAttributes(response.Configurations)
	for i, item := range response.Configurations {
		if i < 90 {
			assert.True(t, proto.Equal(item, &admin.MatchableAttributesConfiguration{
				Project:    fmt.Sprintf("test-project-%d", i/9),
				Domain:     domains[(i/3)%3],
				Workflow:   workflows[i%3],
				Attributes: matchingExecutionQueueAttributes,
			}))
		} else {
			i := i - 90
			assert.True(t, proto.Equal(item, &admin.MatchableAttributesConfiguration{
				Project:    fmt.Sprintf("test-project-%d", i),
				Attributes: matchingExecutionQueueAttributes,
			}))
		}
	}

	response, err = client.ListMatchableAttributes(ctx, &admin.ListMatchableAttributesRequest{
		ResourceType: admin.MatchableResource_CLUSTER_RESOURCE,
	})
	assert.Nil(t, err)
	assert.Len(t, response.Configurations, 0)
}

func TestMigration3(t *testing.T) {
	ctx := context.Background()
	db, err := repositories.GetDB(ctx, getDbConfig(), getLoggerConfig())
	assert.Nil(t, err)
	db.Exec("DELETE FROM migrations WHERE id = 'pg-continue-2024-04-configurations-to-resources'")
	rollbackToDefaultConfiguration(db)
}
