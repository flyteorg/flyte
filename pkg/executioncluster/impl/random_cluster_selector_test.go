package impl

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/lyft/flyteadmin/pkg/errors"
	repo_interface "github.com/lyft/flyteadmin/pkg/repositories/interfaces"
	repo_mock "github.com/lyft/flyteadmin/pkg/repositories/mocks"
	"github.com/lyft/flyteadmin/pkg/repositories/models"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"
	"google.golang.org/grpc/codes"

	"github.com/lyft/flyteadmin/pkg/executioncluster"
	interfaces2 "github.com/lyft/flyteadmin/pkg/executioncluster/interfaces"
	"github.com/lyft/flyteadmin/pkg/executioncluster/mocks"
	"github.com/lyft/flyteadmin/pkg/runtime"
	"github.com/lyft/flytestdlib/config"
	"github.com/lyft/flytestdlib/config/viper"
	"github.com/lyft/flytestdlib/promutils"

	"github.com/stretchr/testify/assert"
)

const testProject = "project"
const testDomain = "domain"
const testWorkflow = "name"

func initTestConfig(fileName string) error {
	pwd, err := os.Getwd()
	if err != nil {
		return err
	}

	configAccessor := viper.NewAccessor(config.Options{
		SearchPaths: []string{filepath.Join(pwd, "../testdata", fileName)},
		StrictMode:  false,
	})
	return configAccessor.UpdateConfig(context.Background())
}

func getRandomClusterSelectorForTest(t *testing.T) interfaces2.ClusterInterface {
	var clusterScope promutils.Scope
	err := initTestConfig("clusters_config.yaml")
	assert.NoError(t, err)

	db := repo_mock.NewMockRepository()
	db.ResourceRepo().(*repo_mock.MockResourceRepo).GetFunction = func(ctx context.Context, ID repo_interface.ResourceID) (resource models.Resource, e error) {
		assert.Equal(t, "EXECUTION_CLUSTER_LABEL", ID.ResourceType)
		if ID.Project == "" {
			return models.Resource{}, errors.NewFlyteAdminErrorf(codes.NotFound,
				"Resource [%+v] not found", ID)
		}
		response := models.Resource{
			Project:      ID.Project,
			Domain:       ID.Domain,
			Workflow:     ID.Workflow,
			ResourceType: ID.ResourceType,
			LaunchPlan:   ID.LaunchPlan,
		}
		if ID.Project == testProject && ID.Domain == testDomain {
			matchingAttributes := &admin.MatchingAttributes{
				Target: &admin.MatchingAttributes_ExecutionClusterLabel{
					ExecutionClusterLabel: &admin.ExecutionClusterLabel{
						Value: "test",
					},
				},
			}
			marshalledMatchingAttributes, _ := proto.Marshal(matchingAttributes)
			response.Attributes = marshalledMatchingAttributes
		} else {
			matchingAttributes := &admin.MatchingAttributes{
				Target: &admin.MatchingAttributes_ExecutionClusterLabel{
					ExecutionClusterLabel: &admin.ExecutionClusterLabel{
						Value: "all",
					},
				},
			}
			marshalledMatchingAttributes, _ := proto.Marshal(matchingAttributes)
			response.Attributes = marshalledMatchingAttributes
		}
		return response, nil
	}
	configProvider := runtime.NewConfigurationProvider()
	randomCluster, err := NewRandomClusterSelector(clusterScope, configProvider, &mocks.MockExecutionTargetProvider{}, db)
	assert.NoError(t, err)
	return randomCluster
}

func TestRandomClusterSelectorGetTarget(t *testing.T) {
	cluster := getRandomClusterSelectorForTest(t)
	target, err := cluster.GetTarget(context.Background(), &executioncluster.ExecutionTargetSpec{TargetID: "testcluster"})
	assert.Nil(t, err)
	assert.Equal(t, "testcluster", target.ID)
	assert.False(t, target.Enabled)
	target, err = cluster.GetTarget(context.Background(), &executioncluster.ExecutionTargetSpec{TargetID: "testcluster2"})
	assert.Nil(t, err)
	assert.Equal(t, "testcluster2", target.ID)
	assert.True(t, target.Enabled)
}

func TestRandomClusterSelectorGetTargetForDomain(t *testing.T) {
	cluster := getRandomClusterSelectorForTest(t)
	target, err := cluster.GetTarget(context.Background(), &executioncluster.ExecutionTargetSpec{
		Project:     testProject,
		Domain:      testDomain,
		ExecutionID: "e",
	})
	assert.Nil(t, err)
	assert.Equal(t, "testcluster2", target.ID)
	assert.True(t, target.Enabled)
}

func TestRandomClusterSelectorGetTargetForExecution(t *testing.T) {
	cluster := getRandomClusterSelectorForTest(t)
	target, err := cluster.GetTarget(context.Background(), &executioncluster.ExecutionTargetSpec{
		Project:     testProject,
		Domain:      "different",
		Workflow:    testWorkflow,
		ExecutionID: "e1",
	})
	assert.Nil(t, err)
	assert.Equal(t, "testcluster3", target.ID)
	assert.True(t, target.Enabled)
}

func TestRandomClusterSelectorGetTargetForDomainAndExecution2(t *testing.T) {
	cluster := getRandomClusterSelectorForTest(t)
	target, err := cluster.GetTarget(context.Background(), &executioncluster.ExecutionTargetSpec{
		Project:     testProject,
		Domain:      "different",
		Workflow:    testWorkflow,
		ExecutionID: "e22",
	})
	assert.Nil(t, err)
	assert.Equal(t, "testcluster2", target.ID)
	assert.True(t, target.Enabled)
}

func TestRandomClusterSelectorGetRandomTarget(t *testing.T) {
	cluster := getRandomClusterSelectorForTest(t)
	target, err := cluster.GetTarget(context.Background(), &executioncluster.ExecutionTargetSpec{
		Project: "",
	})
	assert.Nil(t, err)
	assert.True(t, target.ID == "testcluster1" || target.ID == "testcluster2" || target.ID == "testcluster3")
	assert.True(t, target.Enabled)
}

func TestRandomClusterSelectorGetRandomTargetUsingEmptySpec(t *testing.T) {
	cluster := getRandomClusterSelectorForTest(t)
	_, err := cluster.GetTarget(context.Background(), nil)
	assert.NotNil(t, err)
	assert.EqualError(t, err, "empty executionTargetSpec")
}

func TestRandomClusterSelectorGetRemoteTarget(t *testing.T) {
	cluster := getRandomClusterSelectorForTest(t)
	_, err := cluster.GetTarget(context.Background(), &executioncluster.ExecutionTargetSpec{TargetID: "cluster-3"})
	assert.NotNil(t, err)
	assert.EqualError(t, err, "invalid cluster target cluster-3")
}

func TestRandomClusterSelectorGetAllValidTargets(t *testing.T) {
	cluster := getRandomClusterSelectorForTest(t)
	targets := cluster.GetAllValidTargets()
	assert.Equal(t, 2, len(targets))
}
