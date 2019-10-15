package impl

import (
	"context"
	"go/build"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lyft/flyteadmin/pkg/executioncluster"
	interfaces2 "github.com/lyft/flyteadmin/pkg/executioncluster/interfaces"
	"github.com/lyft/flyteadmin/pkg/executioncluster/mocks"
	"github.com/lyft/flyteadmin/pkg/runtime"
	"github.com/lyft/flyteadmin/pkg/runtime/interfaces"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/lyft/flytestdlib/config"
	"github.com/lyft/flytestdlib/config/viper"
	"github.com/lyft/flytestdlib/promutils"

	"github.com/stretchr/testify/assert"
)

var defaultDomains = []interfaces.Domain{{ID: "d1", Name: "d1"}, {ID: "d2", Name: "d2"}, {ID: "d3", Name: "domain3"}}

func initTestConfig(fileName string) error {
	var searchPaths []string
	for _, goPath := range strings.Split(build.Default.GOPATH, string(os.PathListSeparator)) {
		searchPaths = append(searchPaths, filepath.Join(goPath, "src/github.com/lyft/flyteadmin/pkg/executioncluster/testdata/", fileName))
	}

	configAccessor := viper.NewAccessor(config.Options{
		SearchPaths: searchPaths,
		StrictMode:  false,
	})
	return configAccessor.UpdateConfig(context.Background())
}

func getRandomClusterSelectorForTest(t *testing.T, domainsConfig interfaces.DomainsConfig) interfaces2.ClusterInterface {
	var clusterScope promutils.Scope
	err := initTestConfig("clusters_config.yaml")
	assert.NoError(t, err)

	configProvider := runtime.NewConfigurationProvider()
	randomCluster, err := NewRandomClusterSelector(clusterScope, configProvider.ClusterConfiguration(), &mocks.MockExecutionTargetProvider{}, &domainsConfig)
	assert.NoError(t, err)
	return randomCluster
}

func TestRandomClusterSelectorGetTarget(t *testing.T) {
	cluster := getRandomClusterSelectorForTest(t, defaultDomains)
	target, err := cluster.GetTarget(&executioncluster.ExecutionTargetSpec{TargetID: "testcluster"})
	assert.Nil(t, err)
	assert.Equal(t, "testcluster", target.ID)
	assert.False(t, target.Enabled)
	target, err = cluster.GetTarget(&executioncluster.ExecutionTargetSpec{TargetID: "testcluster2"})
	assert.Nil(t, err)
	assert.Equal(t, "testcluster2", target.ID)
	assert.True(t, target.Enabled)
}

func TestRandomClusterSelectorGetTargetForDomain(t *testing.T) {
	cluster := getRandomClusterSelectorForTest(t, defaultDomains)
	target, err := cluster.GetTarget(&executioncluster.ExecutionTargetSpec{ExecutionID: &core.WorkflowExecutionIdentifier{
		Domain: "d1",
	}})
	assert.Nil(t, err)
	assert.Equal(t, "testcluster2", target.ID)
	assert.True(t, target.Enabled)
}

func TestRandomClusterSelectorGetTargetForDomainAndExecution(t *testing.T) {
	cluster := getRandomClusterSelectorForTest(t, defaultDomains)
	target, err := cluster.GetTarget(&executioncluster.ExecutionTargetSpec{ExecutionID: &core.WorkflowExecutionIdentifier{
		Domain: "d2",
		Name:   "exec",
	}})
	assert.Nil(t, err)
	assert.Equal(t, "testcluster3", target.ID)
	assert.True(t, target.Enabled)
}

func TestRandomClusterSelectorGetTargetForDomainAndExecution2(t *testing.T) {
	cluster := getRandomClusterSelectorForTest(t, defaultDomains)
	target, err := cluster.GetTarget(&executioncluster.ExecutionTargetSpec{ExecutionID: &core.WorkflowExecutionIdentifier{
		Domain: "d2",
		Name:   "exec2",
	}})
	assert.Nil(t, err)
	assert.Equal(t, "testcluster2", target.ID)
	assert.True(t, target.Enabled)
}

func TestRandomClusterSelectorGetTargetForInvalidDomain(t *testing.T) {
	cluster := getRandomClusterSelectorForTest(t, defaultDomains)
	_, err := cluster.GetTarget(&executioncluster.ExecutionTargetSpec{ExecutionID: &core.WorkflowExecutionIdentifier{
		Domain: "d4",
		Name:   "exec",
	}})
	assert.EqualError(t, err, "invalid executionTargetSpec { domain:\"d4\" name:\"exec\" }")
}

func TestRandomClusterSelectorGetRandomTarget(t *testing.T) {
	cluster := getRandomClusterSelectorForTest(t, defaultDomains)
	_, err := cluster.GetTarget(nil)
	assert.NotNil(t, err)
	assert.EqualError(t, err, "invalid executionTargetSpec <nil>")
}

func TestRandomClusterSelectorGetRemoteTarget(t *testing.T) {
	cluster := getRandomClusterSelectorForTest(t, defaultDomains)
	_, err := cluster.GetTarget(&executioncluster.ExecutionTargetSpec{TargetID: "cluster-3"})
	assert.NotNil(t, err)
	assert.EqualError(t, err, "invalid cluster target cluster-3")
}

func TestRandomClusterSelectorGetAllValidTargets(t *testing.T) {
	cluster := getRandomClusterSelectorForTest(t, defaultDomains)
	targets := cluster.GetAllValidTargets()
	assert.Equal(t, 2, len(targets))
}
