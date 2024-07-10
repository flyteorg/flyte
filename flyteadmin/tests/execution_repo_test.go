//go:build integration
// +build integration

package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/errors"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
)

const cluster = "clust"

var terminalPhases = sets.New[core.WorkflowExecution_Phase](
	core.WorkflowExecution_FAILED,
	core.WorkflowExecution_ABORTED,
	core.WorkflowExecution_SUCCEEDED,
	core.WorkflowExecution_TIMED_OUT)

type executionRepoSuite struct {
	suite.Suite
	db   *gorm.DB
	repo interfaces.ExecutionRepoInterface
	ctx  context.Context
}

func (s *executionRepoSuite) SetupSuite() {
	ctx := context.Background()
	db, err := repositories.GetDB(ctx, getDbConfig(), getLoggerConfig())
	s.Require().NoError(err)
	s.db = db
	scope := promutils.NewTestScope()
	dbScope := scope.NewSubScope("database")
	repo := repositories.NewGormRepo(
		db, errors.NewPostgresErrorTransformer(scope.NewSubScope("errors")), dbScope)
	s.repo = repo.ExecutionRepo()
}

func (s *executionRepoSuite) SetupTest() {
	s.Require().NoError(s.db.Exec("TRUNCATE TABLE executions").Error)
	s.ctx = context.TODO()
}

func (s *executionRepoSuite) Test_FindNextStatusUpdatesCheckpoint_Empty() {
	id, err := s.repo.FindNextStatusUpdatesCheckpoint(s.ctx, cluster, uint(42))
	s.NoError(err)
	s.Equal(uint(42), id)
}

func (s *executionRepoSuite) Test_FindNextStatusUpdatesCheckpoint_EmptyForRequestedCluster() {
	execName := ""
	for i := 0; i < 10; i++ {
		execName = rand.String(10)
		s.insertExecution(execName, s.terminalPhase(), "other")
	}

	id, err := s.repo.FindNextStatusUpdatesCheckpoint(s.ctx, cluster, uint(0))
	s.NoError(err)
	s.Equal(uint(0), id)
}

func (s *executionRepoSuite) Test_FindNextStatusUpdatesCheckpoint_AllExecutionsTerminal() {
	execName := ""
	for i := 0; i < 10; i++ {
		execName = rand.String(10)
		s.insertExecution(execName, s.terminalPhase(), cluster)
	}
	exec, err := s.repo.Get(s.ctx, interfaces.Identifier{Project: project, Domain: domain, Name: execName})
	s.NoError(err)

	id, err := s.repo.FindNextStatusUpdatesCheckpoint(s.ctx, cluster, 0)
	s.NoError(err)
	s.Equal(exec.ID, id)
}

func (s *executionRepoSuite) Test_FindNextStatusUpdatesCheckpoint_AllExecutionsNonTerminal() {
	execName := ""
	for i := 0; i < 10; i++ {
		execName = rand.String(10)
		s.insertExecution(execName, s.nonTerminalPhase(), cluster)
	}

	id, err := s.repo.FindNextStatusUpdatesCheckpoint(s.ctx, cluster, 0)
	s.NoError(err)
	s.Equal(uint(0), id)
}

func (s *executionRepoSuite) Test_FindNextStatusUpdatesCheckpoint_NextLastTerminalInStreak() {
	s.insertExecution("a", core.WorkflowExecution_SUCCEEDED, cluster)
	s.insertExecution("b", core.WorkflowExecution_SUCCEEDED, cluster)
	s.insertExecution("c", core.WorkflowExecution_SUCCEEDED, cluster)
	s.insertExecution("d", core.WorkflowExecution_RUNNING, cluster)
	s.insertExecution("e", core.WorkflowExecution_FAILED, cluster)
	s.insertExecution("f", core.WorkflowExecution_QUEUED, cluster)
	exec, err := s.repo.Get(s.ctx, interfaces.Identifier{Project: project, Domain: domain, Name: "a"})
	s.NoError(err)
	nextExec, err := s.repo.Get(s.ctx, interfaces.Identifier{Project: project, Domain: domain, Name: "c"})
	s.NoError(err)

	id, err := s.repo.FindNextStatusUpdatesCheckpoint(s.ctx, cluster, exec.ID)
	s.NoError(err)
	s.Equal(nextExec.ID, id)
}

func (s *executionRepoSuite) Test_FindNextStatusUpdatesCheckpoint_NextLastTerminalInStreakForRequestedCluster() {
	s.insertExecution("a", core.WorkflowExecution_SUCCEEDED, cluster)
	s.insertExecution("b", core.WorkflowExecution_SUCCEEDED, cluster)
	s.insertExecution("c", core.WorkflowExecution_SUCCEEDED, cluster)
	s.insertExecution("d", core.WorkflowExecution_RUNNING, cluster)
	s.insertExecution("e", core.WorkflowExecution_FAILED, cluster)
	s.insertExecution("f", core.WorkflowExecution_TIMED_OUT, cluster)
	s.insertExecution("g", core.WorkflowExecution_ABORTED, cluster)
	s.insertExecution("h", core.WorkflowExecution_QUEUED, cluster)
	exec, err := s.repo.Get(s.ctx, interfaces.Identifier{Project: project, Domain: domain, Name: "a"})
	s.NoError(err)
	nextExec, err := s.repo.Get(s.ctx, interfaces.Identifier{Project: project, Domain: domain, Name: "c"})
	s.NoError(err)

	cluster2 := "cluster2"
	s.insertExecution("i", core.WorkflowExecution_FAILED, cluster2)
	s.insertExecution("j", core.WorkflowExecution_TIMED_OUT, cluster2)
	s.insertExecution("k", core.WorkflowExecution_ABORTED, cluster2)
	s.insertExecution("l", core.WorkflowExecution_QUEUED, cluster2)
	s.insertExecution("m", core.WorkflowExecution_RUNNING, cluster2)
	s.insertExecution("n", core.WorkflowExecution_SUCCEEDED, cluster2)
	exec2, err := s.repo.Get(s.ctx, interfaces.Identifier{Project: project, Domain: domain, Name: "j"})
	s.NoError(err)
	nextExec2, err := s.repo.Get(s.ctx, interfaces.Identifier{Project: project, Domain: domain, Name: "k"})
	s.NoError(err)

	id, err := s.repo.FindNextStatusUpdatesCheckpoint(s.ctx, cluster, exec.ID)
	s.NoError(err)
	s.Equal(nextExec.ID, id)

	id2, err := s.repo.FindNextStatusUpdatesCheckpoint(s.ctx, cluster2, exec2.ID)
	s.NoError(err)
	s.Equal(nextExec2.ID, id2)
}

func (s *executionRepoSuite) Test_FindStatusUpdates_Empty() {
	st, err := s.repo.FindStatusUpdates(s.ctx, cluster, uint(0), 0, 100)
	s.NoError(err)
	s.Empty(st)
}

func (s *executionRepoSuite) Test_FindStatusUpdates_EmptyForRequestedCluster() {
	for i := 0; i < 10; i++ {
		execName := rand.String(10)
		s.insertExecution(execName, s.anyPhase(), "other")
	}

	st, err := s.repo.FindStatusUpdates(s.ctx, cluster, uint(0), 0, 100)
	s.NoError(err)
	s.Empty(st)
}

func (s *executionRepoSuite) Test_FindStatusUpdates_ReturnAllUpdates() {
	s.insertExecution("a", core.WorkflowExecution_SUCCEEDED, cluster)
	s.insertExecution("b", core.WorkflowExecution_SUCCEEDED, cluster)
	s.insertExecution("c", core.WorkflowExecution_SUCCEEDED, cluster)
	s.insertExecution("d", core.WorkflowExecution_RUNNING, cluster)
	s.insertExecution("e", core.WorkflowExecution_FAILED, cluster)
	s.insertExecution("f", core.WorkflowExecution_TIMED_OUT, cluster)
	s.insertExecution("g", core.WorkflowExecution_ABORTED, cluster)
	s.insertExecution("h", core.WorkflowExecution_QUEUED, cluster)

	st, err := s.repo.FindStatusUpdates(s.ctx, cluster, uint(0), 5, 0)
	s.NoError(err)
	s.Len(st, 5)

	st, err = s.repo.FindStatusUpdates(s.ctx, cluster, uint(0), 5, 5)
	s.NoError(err)
	s.Len(st, 3)
}

func (s *executionRepoSuite) Test_FindStatusUpdates_ReturnAllUpdatesForRequestedCluster() {
	s.insertExecution("a", core.WorkflowExecution_SUCCEEDED, cluster)
	s.insertExecution("b", core.WorkflowExecution_SUCCEEDED, cluster)
	s.insertExecution("c", core.WorkflowExecution_SUCCEEDED, cluster)
	s.insertExecution("d", core.WorkflowExecution_RUNNING, cluster)
	s.insertExecution("e", core.WorkflowExecution_FAILED, cluster)
	s.insertExecution("f", core.WorkflowExecution_TIMED_OUT, cluster)
	s.insertExecution("g", core.WorkflowExecution_ABORTED, cluster)
	s.insertExecution("h", core.WorkflowExecution_QUEUED, cluster)

	cluster2 := "clust2"
	s.insertExecution("i", core.WorkflowExecution_FAILED, cluster2)
	s.insertExecution("j", core.WorkflowExecution_TIMED_OUT, cluster2)
	s.insertExecution("k", core.WorkflowExecution_ABORTED, cluster2)
	s.insertExecution("l", core.WorkflowExecution_QUEUED, cluster2)
	s.insertExecution("m", core.WorkflowExecution_RUNNING, cluster2)
	s.insertExecution("n", core.WorkflowExecution_SUCCEEDED, cluster2)

	st, err := s.repo.FindStatusUpdates(s.ctx, cluster, uint(0), 20, 0)
	s.NoError(err)
	s.Len(st, 8)

	st, err = s.repo.FindStatusUpdates(s.ctx, cluster2, uint(0), 20, 0)
	s.NoError(err)
	s.Len(st, 6)
}

func (s *executionRepoSuite) terminalPhase() core.WorkflowExecution_Phase {
	n := terminalPhases.Len()
	return terminalPhases.UnsortedList()[rand.Intn(n)]
}

func (s *executionRepoSuite) anyPhase() core.WorkflowExecution_Phase {
	n := len(core.WorkflowExecution_Phase_name)
	return core.WorkflowExecution_Phase(rand.Intn(n))
}

func (s *executionRepoSuite) nonTerminalPhase() core.WorkflowExecution_Phase {
	for {
		phase := s.anyPhase()
		if !terminalPhases.Has(phase) {
			return phase
		}
	}
}

func (s *executionRepoSuite) insertExecution(name string, phase core.WorkflowExecution_Phase, parentCluster string) {
	err := s.repo.Create(s.ctx, models.Execution{
		ExecutionKey: models.ExecutionKey{
			Project: project,
			Domain:  domain,
			Name:    name,
		},
		Phase:         phase.String(),
		Spec:          []byte("foo"),
		ParentCluster: parentCluster,
	}, nil)
	s.Require().NoError(err)
}

func TestExecutionRepoSuite(t *testing.T) {
	suite.Run(t, new(executionRepoSuite))
}
