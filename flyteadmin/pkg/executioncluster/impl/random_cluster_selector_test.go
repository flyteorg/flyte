package impl

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/flyteorg/flyteadmin/pkg/errors"
	repo_interface "github.com/flyteorg/flyteadmin/pkg/repositories/interfaces"
	repo_mock "github.com/flyteorg/flyteadmin/pkg/repositories/mocks"
	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc/codes"

	"github.com/flyteorg/flyteadmin/pkg/executioncluster"
	interfaces2 "github.com/flyteorg/flyteadmin/pkg/executioncluster/interfaces"
	"github.com/flyteorg/flyteadmin/pkg/executioncluster/mocks"
	"github.com/flyteorg/flyteadmin/pkg/runtime"
	"github.com/flyteorg/flytestdlib/config"
	"github.com/flyteorg/flytestdlib/config/viper"
	"github.com/stretchr/testify/assert"
)

const (
	testProject                    = "project"
	testDomain                     = "domain"
	testWorkflow                   = "name"
	testCluster1                   = "testcluster1"
	testCluster2                   = "testcluster2"
	testCluster3                   = "testcluster3"
	clusterConfig1                 = "clusters_config.yaml"
	clusterConfig2                 = "clusters_config2.yaml"
	clusterConfig2WithDefaultLabel = "clusters_config2_default_label.yaml"
)

func initTestConfig(fileName string) error {
	pwd, err := os.Getwd()
	if err != nil {
		return err
	}

	configAccessor := viper.NewAccessor(config.Options{
		SearchPaths: []string{filepath.Join(pwd, "testdata", fileName)},
		StrictMode:  false,
	})
	return configAccessor.UpdateConfig(context.Background())
}

func getRandomClusterSelectorForTest(t *testing.T) interfaces2.ClusterInterface {
	err := initTestConfig(clusterConfig1)
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
	listTargetsProvider := mocks.ListTargetsInterface{}
	validTargets := map[string]*executioncluster.ExecutionTarget{
		testCluster2: {
			ID:      testCluster2,
			Enabled: true,
		},
		testCluster3: {
			ID:      testCluster3,
			Enabled: true,
		},
	}
	targets := map[string]*executioncluster.ExecutionTarget{
		testCluster1: {
			ID: testCluster1,
		},
		testCluster2: {
			ID:      testCluster2,
			Enabled: true,
		},
		testCluster3: {
			ID:      testCluster3,
			Enabled: true,
		},
	}
	listTargetsProvider.OnGetValidTargets().Return(validTargets)
	listTargetsProvider.OnGetAllTargets().Return(targets)
	randomCluster, err := NewRandomClusterSelector(&listTargetsProvider, configProvider, db)
	assert.NoError(t, err)
	return randomCluster
}

func getRandomClusterSelectorWithDefaultLabelForTest(t *testing.T, configFile string) interfaces2.ClusterInterface {
	err := initTestConfig(configFile)
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
						Value: "two",
					},
				},
			}
			marshalledMatchingAttributes, _ := proto.Marshal(matchingAttributes)
			response.Attributes = marshalledMatchingAttributes
		}
		return response, nil
	}
	configProvider := runtime.NewConfigurationProvider()
	listTargetsProvider := mocks.ListTargetsInterface{}
	validTargets := map[string]*executioncluster.ExecutionTarget{
		testCluster1: {
			ID:      testCluster1,
			Enabled: true,
		},
		testCluster2: {
			ID:      testCluster2,
			Enabled: true,
		},
		testCluster3: {
			ID:      testCluster3,
			Enabled: true,
		},
	}
	targets := map[string]*executioncluster.ExecutionTarget{
		testCluster1: {
			ID:      testCluster1,
			Enabled: true,
		},
		testCluster2: {
			ID:      testCluster2,
			Enabled: true,
		},
		testCluster3: {
			ID:      testCluster3,
			Enabled: true,
		},
	}
	listTargetsProvider.OnGetValidTargets().Return(validTargets)
	listTargetsProvider.OnGetAllTargets().Return(targets)
	randomCluster, err := NewRandomClusterSelector(&listTargetsProvider, configProvider, db)
	assert.NoError(t, err)
	return randomCluster
}

func TestRandomClusterSelectorGetTarget(t *testing.T) {
	cluster := getRandomClusterSelectorForTest(t)
	target, err := cluster.GetTarget(context.Background(), &executioncluster.ExecutionTargetSpec{TargetID: testCluster1})
	assert.Nil(t, err)
	assert.Equal(t, testCluster1, target.ID)
	assert.False(t, target.Enabled)
	target, err = cluster.GetTarget(context.Background(), &executioncluster.ExecutionTargetSpec{TargetID: testCluster2})
	assert.Nil(t, err)
	assert.Equal(t, testCluster2, target.ID)
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
	assert.Equal(t, testCluster2, target.ID)
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
	assert.Equal(t, testCluster3, target.ID)
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
	assert.Equal(t, testCluster2, target.ID)
	assert.True(t, target.Enabled)
}

func TestRandomClusterSelectorGetRandomTarget(t *testing.T) {
	cluster := getRandomClusterSelectorForTest(t)
	target, err := cluster.GetTarget(context.Background(), &executioncluster.ExecutionTargetSpec{
		Project: "",
	})
	assert.Nil(t, err)
	assert.True(t, target.ID == testCluster1 || target.ID == testCluster2 || target.ID == testCluster3)
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
	targets := cluster.GetValidTargets()
	assert.Equal(t, 2, len(targets))
}

func TestRandomClusterSelectorGetTargetWithFallbackToDefault1(t *testing.T) {
	cluster := getRandomClusterSelectorWithDefaultLabelForTest(t, clusterConfig2)
	target, err := cluster.GetTarget(context.Background(), &executioncluster.ExecutionTargetSpec{
		Project:     testProject,
		Domain:      "different",
		Workflow:    testWorkflow,
		ExecutionID: "e3",
	})
	assert.Nil(t, err)
	assert.Equal(t, testCluster3, target.ID)
	assert.True(t, target.Enabled)
}

func TestRandomClusterSelectorGetTargetWithFallbackToDefault2(t *testing.T) {
	cluster := getRandomClusterSelectorWithDefaultLabelForTest(t, clusterConfig2)
	target, err := cluster.GetTarget(context.Background(), &executioncluster.ExecutionTargetSpec{
		Project:     testProject,
		Domain:      testDomain,
		Workflow:    testWorkflow,
		ExecutionID: "e3",
	})
	assert.Nil(t, err)
	assert.Equal(t, testCluster2, target.ID)
	assert.True(t, target.Enabled)
}

func TestRandomClusterSelectorGetTargetWithFallbackToDefault3(t *testing.T) {
	cluster := getRandomClusterSelectorWithDefaultLabelForTest(t, clusterConfig2WithDefaultLabel)
	target, err := cluster.GetTarget(context.Background(), &executioncluster.ExecutionTargetSpec{
		Project:     testProject,
		Domain:      "different",
		Workflow:    testWorkflow,
		ExecutionID: "e3",
	})
	assert.Nil(t, err)
	assert.Equal(t, testCluster1, target.ID)
	assert.True(t, target.Enabled)
}
