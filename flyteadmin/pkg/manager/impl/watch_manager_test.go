package impl

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc/codes"

	"github.com/flyteorg/flyte/flyteadmin/pkg/config"
	flyteAdminErrors "github.com/flyteorg/flyte/flyteadmin/pkg/errors"
	repoInterfaces "github.com/flyteorg/flyte/flyteadmin/pkg/repositories/interfaces"
	repositoryMocks "github.com/flyteorg/flyte/flyteadmin/pkg/repositories/mocks"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyte/flyteidl/clients/go/admin/mocks"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/watch"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
)

const (
	cluster     = "clust"
	maxPageSize = 3
	delta       = 10 * time.Millisecond
)

var (
	execKey1 = models.ExecutionKey{
		Project: project,
		Domain:  domain,
		Name:    "foo",
	}
	execKey2 = models.ExecutionKey{
		Project: project,
		Domain:  domain,
		Name:    "bar",
	}
	execKey3 = models.ExecutionKey{
		Project: project,
		Domain:  domain,
		Name:    "baz",
	}
	execKey4 = models.ExecutionKey{
		Project: project,
		Domain:  domain,
		Name:    "boo",
	}
	execKey5 = models.ExecutionKey{
		Project: project,
		Domain:  domain,
		Name:    "zoo",
	}
)

type watchManagerSuite struct {
	suite.Suite
	ctx     context.Context
	repo    *repositoryMocks.ExecutionRepoInterface
	srv     *mocks.WatchService_WatchExecutionStatusUpdatesServer
	manager *WatchManager
}

func (s *watchManagerSuite) SetupTest() {
	s.ctx = context.TODO()
	cfg := config.GetConfig()
	cfg.WatchService.MaxPageSize = maxPageSize
	cfg.WatchService.PollInterval.Duration = time.Second
	cfg.WatchService.MaxActiveClusterConnections = 1
	cfg.WatchService.NonTerminalStatusUpdatesInterval.Duration = time.Minute
	db := repositoryMocks.NewMockRepository()
	s.repo = &repositoryMocks.ExecutionRepoInterface{}
	db.(*repositoryMocks.MockRepository).ExecutionRepoIface = s.repo
	s.srv = &mocks.WatchService_WatchExecutionStatusUpdatesServer{}
	s.manager = NewWatchManager(cfg, db, promutils.NewTestScope())
}

func (s *watchManagerSuite) TearDownTest() {
	s.srv.AssertExpectations(s.T())
	s.repo.AssertExpectations(s.T())
}

func (s *watchManagerSuite) Test_readAndSendExecutions_Empty() {
	alreadySent := alreadySentMap{}
	expectedCheckpoint := uint(42)
	s.repo.
		OnFindNextStatusUpdatesCheckpoint(s.ctx, cluster, 0).Return(expectedCheckpoint, nil).Once()
	var expectedExecutions []repoInterfaces.ExecutionStatus
	s.repo.
		OnFindStatusUpdates(s.ctx, cluster, 0, maxPageSize, 0).
		Return(expectedExecutions, nil).
		Once()

	newCheckpoint, err := s.manager.readAndSendExecutions(s.ctx, cluster, 0, alreadySent, s.srv)

	s.NoError(err)
	s.Equal(expectedCheckpoint, newCheckpoint)
	s.Empty(alreadySent)
}

func (s *watchManagerSuite) Test_readAndSendExecutions_Success() {
	alreadySent := alreadySentMap{}
	expectedCheckpoint := uint(1)
	s.repo.
		OnFindNextStatusUpdatesCheckpoint(s.ctx, cluster, 0).Return(expectedCheckpoint, nil).Once()
	expectedExecutions := []repoInterfaces.ExecutionStatus{
		{
			ExecutionKey: execKey2,
			ID:           2,
			Closure: s.marshal(&admin.ExecutionClosure{
				Phase: core.WorkflowExecution_RUNNING,
			}),
		},
		{
			ExecutionKey: execKey1,
			ID:           1,
			Closure: s.marshal(&admin.ExecutionClosure{
				OutputResult: &admin.ExecutionClosure_Outputs{
					Outputs: &admin.LiteralMapBlob{
						Data: &admin.LiteralMapBlob_Uri{Uri: "s3://data"},
					},
				},
				Phase: core.WorkflowExecution_SUCCEEDED,
			}),
		},
	}
	s.repo.
		OnFindStatusUpdates(s.ctx, cluster, 0, maxPageSize, 0).
		Return(expectedExecutions, nil).
		Once()
	s.srv.
		OnSend(&watch.WatchExecutionStatusUpdatesResponse{
			Id: &core.WorkflowExecutionIdentifier{
				Project: execKey1.Project,
				Domain:  execKey1.Domain,
				Name:    execKey1.Name,
			},
			Phase:     core.WorkflowExecution_SUCCEEDED,
			OutputUri: "s3://data",
		}).
		Return(nil).
		Once()
	s.srv.
		OnSend(&watch.WatchExecutionStatusUpdatesResponse{
			Id: &core.WorkflowExecutionIdentifier{
				Project: execKey2.Project,
				Domain:  execKey2.Domain,
				Name:    execKey2.Name,
			},
			Phase: core.WorkflowExecution_RUNNING,
		}).
		Return(nil).
		Once()

	newCheckpoint, err := s.manager.readAndSendExecutions(s.ctx, cluster, 0, alreadySent, s.srv)

	s.NoError(err)
	s.Equal(expectedCheckpoint, newCheckpoint)
	s.Len(alreadySent, 1)
	lastUpdate, ok := alreadySent[2]
	if s.True(ok) {
		s.False(lastUpdate.terminal)
		s.WithinDuration(time.Now(), lastUpdate.sentAt, delta)
	}
}

func (s *watchManagerSuite) Test_readAndSendExecutions_Pagination() {
	alreadySent := alreadySentMap{}
	expectedCheckpoint := uint(1)
	s.repo.
		OnFindNextStatusUpdatesCheckpoint(s.ctx, cluster, 0).Return(expectedCheckpoint, nil).Once()
	expectedExecutions := []repoInterfaces.ExecutionStatus{
		{
			ExecutionKey: execKey1,
			ID:           1,
			Closure: s.marshal(&admin.ExecutionClosure{
				OutputResult: &admin.ExecutionClosure_Outputs{
					Outputs: &admin.LiteralMapBlob{
						Data: &admin.LiteralMapBlob_Uri{Uri: "s3://data"},
					},
				},
				Phase: core.WorkflowExecution_SUCCEEDED,
			}),
		},
		{
			ExecutionKey: execKey2,
			ID:           2,
			Closure: s.marshal(&admin.ExecutionClosure{
				Phase: core.WorkflowExecution_QUEUED,
			}),
		},
		{
			ExecutionKey: execKey3,
			ID:           3,
			Closure: s.marshal(&admin.ExecutionClosure{
				OutputResult: &admin.ExecutionClosure_Error{
					Error: &core.ExecutionError{Message: "wrong!"},
				},
				Phase: core.WorkflowExecution_FAILED,
			}),
		},
	}
	s.repo.
		OnFindStatusUpdates(s.ctx, cluster, 0, maxPageSize, 0).
		Return(expectedExecutions, nil).
		Once()
	s.srv.
		OnSend(&watch.WatchExecutionStatusUpdatesResponse{
			Id: &core.WorkflowExecutionIdentifier{
				Project: execKey1.Project,
				Domain:  execKey1.Domain,
				Name:    execKey1.Name,
			},
			Phase:     core.WorkflowExecution_SUCCEEDED,
			OutputUri: "s3://data",
		}).
		Return(nil).
		Once()
	s.srv.
		OnSend(&watch.WatchExecutionStatusUpdatesResponse{
			Id: &core.WorkflowExecutionIdentifier{
				Project: execKey2.Project,
				Domain:  execKey2.Domain,
				Name:    execKey2.Name,
			},
			Phase: core.WorkflowExecution_QUEUED,
		}).
		Return(nil).
		Once()
	s.srv.
		OnSend(&watch.WatchExecutionStatusUpdatesResponse{
			Id: &core.WorkflowExecutionIdentifier{
				Project: execKey3.Project,
				Domain:  execKey3.Domain,
				Name:    execKey3.Name,
			},
			Phase: core.WorkflowExecution_FAILED,
			Error: &core.ExecutionError{Message: "wrong!"},
		}).
		Return(nil).
		Once()
	s.repo.
		OnFindStatusUpdates(s.ctx, cluster, 0, maxPageSize, maxPageSize).
		Return([]repoInterfaces.ExecutionStatus{}, nil).
		Once()

	newCheckpoint, err := s.manager.readAndSendExecutions(s.ctx, cluster, 0, alreadySent, s.srv)

	s.NoError(err)
	s.Equal(expectedCheckpoint, newCheckpoint)
	s.Len(alreadySent, 2)
	lastUpdate, ok := alreadySent[2]
	if s.True(ok) {
		s.False(lastUpdate.terminal)
		s.WithinDuration(time.Now(), lastUpdate.sentAt, delta)
	}
	lastUpdate, ok = alreadySent[3]
	if s.True(ok) {
		s.True(lastUpdate.terminal)
		s.WithinDuration(time.Now(), lastUpdate.sentAt, delta)
	}
}

func (s *watchManagerSuite) Test_readAndSendExecutions_SkipAlreadySent() {
	alreadySent := alreadySentMap{
		2: {terminal: true},                                            // skipped, because its terminal
		3: {terminal: false, sentAt: time.Now()},                       // skipped, because it was recently sent
		4: {terminal: false, sentAt: time.Now().Add(-2 * time.Minute)}, // sent, non-terminal update that was sent long time ago
		5: {terminal: false, sentAt: time.Now()},                       // sent, recently it was non-terminal, but the latest is terminal
	}
	expectedCheckpoint := uint(2)
	s.repo.
		OnFindNextStatusUpdatesCheckpoint(s.ctx, cluster, 0).Return(expectedCheckpoint, nil).Once()
	expectedExecutions := []repoInterfaces.ExecutionStatus{
		{
			ExecutionKey: execKey1,
			ID:           1,
			Closure: s.marshal(&admin.ExecutionClosure{
				OutputResult: &admin.ExecutionClosure_Outputs{
					Outputs: &admin.LiteralMapBlob{
						Data: &admin.LiteralMapBlob_Uri{Uri: "s3://data"},
					},
				},
				Phase: core.WorkflowExecution_SUCCEEDED,
			}),
		},
		{
			ExecutionKey: execKey2,
			ID:           2,
			Closure: s.marshal(&admin.ExecutionClosure{
				OutputResult: &admin.ExecutionClosure_Error{
					Error: &core.ExecutionError{Message: "wrong!"},
				},
				Phase: core.WorkflowExecution_FAILED,
			}),
		},
		{
			ExecutionKey: execKey3,
			ID:           3,
			Closure: s.marshal(&admin.ExecutionClosure{
				Phase: core.WorkflowExecution_QUEUED,
			}),
		},
		{
			ExecutionKey: execKey4,
			ID:           4,
			Closure: s.marshal(&admin.ExecutionClosure{
				Phase: core.WorkflowExecution_RUNNING,
			}),
		},
		{
			ExecutionKey: execKey5,
			ID:           5,
			Closure: s.marshal(&admin.ExecutionClosure{
				Phase: core.WorkflowExecution_ABORTED,
			}),
		},
	}
	s.repo.
		OnFindStatusUpdates(s.ctx, cluster, 0, maxPageSize, 0).
		Return(expectedExecutions[:3], nil).
		Once()
	s.repo.
		OnFindStatusUpdates(s.ctx, cluster, 0, maxPageSize, maxPageSize).
		Return(expectedExecutions[3:], nil).
		Once()
	s.srv.
		OnSend(&watch.WatchExecutionStatusUpdatesResponse{
			Id: &core.WorkflowExecutionIdentifier{
				Project: execKey1.Project,
				Domain:  execKey1.Domain,
				Name:    execKey1.Name,
			},
			Phase:     core.WorkflowExecution_SUCCEEDED,
			OutputUri: "s3://data",
		}).
		Return(nil).
		Once()
	s.srv.
		OnSend(&watch.WatchExecutionStatusUpdatesResponse{
			Id: &core.WorkflowExecutionIdentifier{
				Project: execKey4.Project,
				Domain:  execKey4.Domain,
				Name:    execKey4.Name,
			},
			Phase: core.WorkflowExecution_RUNNING,
		}).
		Return(nil).
		Once()
	s.srv.
		OnSend(&watch.WatchExecutionStatusUpdatesResponse{
			Id: &core.WorkflowExecutionIdentifier{
				Project: execKey5.Project,
				Domain:  execKey5.Domain,
				Name:    execKey5.Name,
			},
			Phase: core.WorkflowExecution_ABORTED,
		}).
		Return(nil).
		Once()

	newCheckpoint, err := s.manager.readAndSendExecutions(s.ctx, cluster, 0, alreadySent, s.srv)

	s.NoError(err)
	s.Equal(expectedCheckpoint, newCheckpoint)
	s.Len(alreadySent, 3)
	lastUpdate, ok := alreadySent[3]
	if s.True(ok) {
		s.False(lastUpdate.terminal)
		s.WithinDuration(time.Now(), lastUpdate.sentAt, delta)
	}
	lastUpdate, ok = alreadySent[4]
	if s.True(ok) {
		s.False(lastUpdate.terminal)
		s.WithinDuration(time.Now(), lastUpdate.sentAt, delta)
	}
	lastUpdate, ok = alreadySent[5]
	if s.True(ok) {
		s.True(lastUpdate.terminal)
		s.WithinDuration(time.Now(), lastUpdate.sentAt, delta)
	}
}

func (s *watchManagerSuite) Test_readAndSendExecutions_FindNextStatusUpdatesCheckpointError() {
	alreadySent := alreadySentMap{}
	expectedCheckpoint := uint(0)
	expectedErr := errors.New("fail")
	s.repo.
		OnFindNextStatusUpdatesCheckpoint(s.ctx, cluster, 0).Return(expectedCheckpoint, expectedErr).Once()

	newCheckpoint, err := s.manager.readAndSendExecutions(s.ctx, cluster, 0, alreadySent, s.srv)

	s.Equal(expectedErr, err)
	s.Equal(expectedCheckpoint, newCheckpoint)
	s.Empty(alreadySent)
}

func (s *watchManagerSuite) Test_readAndSendExecutions_FindTerminalStatusUpdatesError() {
	alreadySent := alreadySentMap{}
	s.repo.
		OnFindNextStatusUpdatesCheckpoint(s.ctx, cluster, 0).Return(uint(42), nil).Once()
	expectedErr := errors.New("fail")
	s.repo.
		OnFindStatusUpdates(s.ctx, cluster, 0, maxPageSize, 0).
		Return([]repoInterfaces.ExecutionStatus{}, expectedErr).
		Once()

	newCheckpoint, err := s.manager.readAndSendExecutions(s.ctx, cluster, 0, alreadySent, s.srv)

	s.Equal(expectedErr, err)
	s.Equal(uint(0), newCheckpoint)
	s.Empty(alreadySent)
}

func (s *watchManagerSuite) Test_WatchExecutionStatusUpdates_MissingCluster() {
	s.srv.OnContext().Return(s.ctx).Once()

	err := s.manager.WatchExecutionStatusUpdates(&watch.WatchExecutionStatusUpdatesRequest{}, s.srv)

	s.Equal(flyteAdminErrors.NewFlyteAdminError(codes.InvalidArgument, "missing 'cluster'"), err)
}

func (s *watchManagerSuite) Test_WatchExecutionStatusUpdates_FindFirstStatusUpdatesCheckpointError() {
	s.srv.OnContext().Return(s.ctx).Once()
	expected := errors.New("fail")
	s.repo.OnFindNextStatusUpdatesCheckpoint(s.ctx, cluster, uint(0)).Return(uint(0), expected).Once()

	err := s.manager.WatchExecutionStatusUpdates(&watch.WatchExecutionStatusUpdatesRequest{Cluster: cluster}, s.srv)

	s.Equal(expected, err)
}

func (s *watchManagerSuite) Test_WatchExecutionStatusUpdates_FindNextStatusUpdatesCheckpointError() {
	s.srv.OnContext().Return(s.ctx).Once()
	s.repo.OnFindNextStatusUpdatesCheckpoint(s.ctx, cluster, uint(0)).Return(uint(42), nil).Once()
	expected := errors.New("fail")
	s.repo.OnFindNextStatusUpdatesCheckpoint(s.ctx, cluster, uint(42)).Return(uint(0), expected).Once()

	err := s.manager.WatchExecutionStatusUpdates(&watch.WatchExecutionStatusUpdatesRequest{Cluster: cluster}, s.srv)

	s.Equal(expected, err)
}

func (s *watchManagerSuite) marshal(m proto.Message) []byte {
	bts, err := proto.Marshal(m)
	s.Require().NoError(err)
	return bts
}

func TestWatchManagerSuite(t *testing.T) {
	suite.Run(t, new(watchManagerSuite))
}
