package register

import (
	"fmt"
	"testing"

	"github.com/flyteorg/flyte/flyteidl/clients/go/admin/mocks"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/service"

	"github.com/flyteorg/flyte/flytestdlib/contextutils"
	"github.com/flyteorg/flyte/flytestdlib/promutils/labeled"
	"github.com/flyteorg/flyte/flytestdlib/storage"
	"github.com/flyteorg/flytectl/cmd/config"

	"github.com/flyteorg/flyte/flytestdlib/promutils"
	rconfig "github.com/flyteorg/flytectl/cmd/config/subcommand/register"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	s3Output = "s3://dummy/prefix"
)

func TestRegisterFromFiles(t *testing.T) {
	t.Run("Valid registration", func(t *testing.T) {
		s := setup()
		registerFilesSetup()
		rconfig.DefaultFilesConfig.Archive = true
		args := []string{"testdata/valid-parent-folder-register.tar"}
		s.MockAdminClient.OnCreateTaskMatch(mock.Anything, mock.Anything).Return(nil, nil)
		s.MockAdminClient.OnCreateWorkflowMatch(mock.Anything, mock.Anything).Return(nil, nil)
		s.MockAdminClient.OnCreateLaunchPlanMatch(mock.Anything, mock.Anything).Return(nil, nil)
		s.MockAdminClient.OnUpdateLaunchPlanMatch(mock.Anything, mock.Anything).Return(nil, nil)
		err := registerFromFilesFunc(s.Ctx, args, s.CmdCtx)
		assert.Nil(t, err)
	})
	t.Run("Valid fast registration", func(t *testing.T) {
		s := setup()
		testScope := promutils.NewTestScope()
		labeled.SetMetricKeys(contextutils.AppNameKey, contextutils.ProjectKey, contextutils.DomainKey)
		registerFilesSetup()
		rconfig.DefaultFilesConfig.Archive = true
		rconfig.DefaultFilesConfig.OutputLocationPrefix = s3Output
		rconfig.DefaultFilesConfig.DeprecatedSourceUploadPath = s3Output
		mockStorage, err := storage.NewDataStore(&storage.Config{
			Type: storage.TypeMemory,
		}, testScope.NewSubScope("flytectl"))
		assert.Nil(t, err)
		Client = mockStorage

		args := []string{"testdata/flytesnacks-core.tgz"}
		s.MockAdminClient.OnCreateTaskMatch(mock.Anything, mock.Anything).Return(nil, nil)
		s.MockAdminClient.OnCreateWorkflowMatch(mock.Anything, mock.Anything).Return(nil, nil)
		s.MockAdminClient.OnCreateLaunchPlanMatch(mock.Anything, mock.Anything).Return(nil, nil)
		s.MockAdminClient.OnUpdateLaunchPlanMatch(mock.Anything, mock.Anything).Return(nil, nil)
		mockDataProxy := s.MockClient.DataProxyClient().(*mocks.DataProxyServiceClient)
		mockDataProxy.OnCreateUploadLocationMatch(s.Ctx, mock.Anything).Return(&service.CreateUploadLocationResponse{}, nil)

		err = registerFromFilesFunc(s.Ctx, args, s.CmdCtx)
		assert.Nil(t, err)
	})
	t.Run("Failed fast registration while uploading the codebase", func(t *testing.T) {
		s := setup()
		registerFilesSetup()
		testScope := promutils.NewTestScope()
		labeled.SetMetricKeys(contextutils.AppNameKey, contextutils.ProjectKey, contextutils.DomainKey)
		rconfig.DefaultFilesConfig.Archive = true
		rconfig.DefaultFilesConfig.OutputLocationPrefix = s3Output
		store, err := storage.NewDataStore(&storage.Config{
			Type: storage.TypeMemory,
		}, testScope.NewSubScope("flytectl"))
		assert.Nil(t, err)
		Client = store
		args := []string{"testdata/flytesnacks-core.tgz"}
		s.MockAdminClient.OnCreateTaskMatch(mock.Anything, mock.Anything).Return(nil, nil)
		s.MockAdminClient.OnCreateWorkflowMatch(mock.Anything, mock.Anything).Return(nil, nil)
		s.MockAdminClient.OnCreateLaunchPlanMatch(mock.Anything, mock.Anything).Return(nil, nil)
		s.MockAdminClient.OnUpdateLaunchPlanMatch(mock.Anything, mock.Anything).Return(nil, nil)
		s.MockClient.DataProxyClient().(*mocks.DataProxyServiceClient).OnCreateUploadLocationMatch(mock.Anything, mock.Anything).Return(&service.CreateUploadLocationResponse{}, nil)
		err = Register(s.Ctx, args, config.GetConfig(), s.CmdCtx)
		assert.Nil(t, err)
	})
	t.Run("Failed registration because of invalid files", func(t *testing.T) {
		s := setup()
		registerFilesSetup()
		testScope := promutils.NewTestScope()
		labeled.SetMetricKeys(contextutils.AppNameKey, contextutils.ProjectKey, contextutils.DomainKey)
		rconfig.DefaultFilesConfig.Archive = true
		rconfig.DefaultFilesConfig.DeprecatedSourceUploadPath = ""
		store, err := storage.NewDataStore(&storage.Config{
			Type: storage.TypeMemory,
		}, testScope.NewSubScope("flytectl"))
		Client = store
		assert.Nil(t, err)
		args := []string{"testdata/invalid-fast.tgz"}
		s.MockAdminClient.OnCreateTaskMatch(mock.Anything, mock.Anything).Return(nil, nil)
		s.MockAdminClient.OnCreateWorkflowMatch(mock.Anything, mock.Anything).Return(nil, nil)
		s.MockAdminClient.OnCreateLaunchPlanMatch(mock.Anything, mock.Anything).Return(nil, nil)
		s.MockAdminClient.OnUpdateLaunchPlanMatch(mock.Anything, mock.Anything).Return(nil, nil)
		err = registerFromFilesFunc(s.Ctx, args, s.CmdCtx)
		assert.NotNil(t, err)
	})
	t.Run("Failure registration of fast serialize", func(t *testing.T) {
		s := setup()
		registerFilesSetup()
		testScope := promutils.NewTestScope()
		labeled.SetMetricKeys(contextutils.AppNameKey, contextutils.ProjectKey, contextutils.DomainKey)
		rconfig.DefaultFilesConfig.Archive = true

		rconfig.DefaultFilesConfig.OutputLocationPrefix = s3Output
		rconfig.DefaultFilesConfig.DeprecatedSourceUploadPath = s3Output
		store, err := storage.NewDataStore(&storage.Config{
			Type: storage.TypeMemory,
		}, testScope.NewSubScope("flytectl"))
		Client = store
		assert.Nil(t, err)
		args := []string{"testdata/flytesnacks-core.tgz"}
		s.MockAdminClient.OnCreateTaskMatch(mock.Anything, mock.Anything).Return(nil, fmt.Errorf("failed")).Call.Times(1)
		s.MockAdminClient.OnCreateWorkflowMatch(mock.Anything, mock.Anything).Return(nil, fmt.Errorf("failed")).Call.Times(1)
		s.MockAdminClient.OnCreateLaunchPlanMatch(mock.Anything, mock.Anything).Return(nil, fmt.Errorf("failed")).Call.Times(1)
		s.MockClient.DataProxyClient().(*mocks.DataProxyServiceClient).OnCreateUploadLocationMatch(mock.Anything, mock.Anything).Return(&service.CreateUploadLocationResponse{}, nil)
		err = registerFromFilesFunc(s.Ctx, args, s.CmdCtx)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("failed"), err)
	})
	t.Run("Failure registration of fast serialize continue on error", func(t *testing.T) {
		s := setup()
		registerFilesSetup()
		testScope := promutils.NewTestScope()
		labeled.SetMetricKeys(contextutils.AppNameKey, contextutils.ProjectKey, contextutils.DomainKey)
		rconfig.DefaultFilesConfig.Archive = true

		rconfig.DefaultFilesConfig.OutputLocationPrefix = s3Output
		rconfig.DefaultFilesConfig.DeprecatedSourceUploadPath = s3Output
		rconfig.DefaultFilesConfig.ContinueOnError = true
		store, err := storage.NewDataStore(&storage.Config{
			Type: storage.TypeMemory,
		}, testScope.NewSubScope("flytectl"))
		Client = store
		assert.Nil(t, err)
		args := []string{"testdata/flytesnacks-core.tgz"}
		s.MockAdminClient.OnCreateTaskMatch(mock.Anything, mock.Anything).Return(nil, fmt.Errorf("failed")).Call.Times(39)
		s.MockAdminClient.OnCreateWorkflowMatch(mock.Anything, mock.Anything).Return(nil, fmt.Errorf("failed")).Call.Times(21)
		s.MockAdminClient.OnCreateLaunchPlanMatch(mock.Anything, mock.Anything).Return(nil, fmt.Errorf("failed")).Call.Times(24)
		s.MockClient.DataProxyClient().(*mocks.DataProxyServiceClient).OnCreateUploadLocationMatch(mock.Anything, mock.Anything).Return(&service.CreateUploadLocationResponse{}, nil)
		err = registerFromFilesFunc(s.Ctx, args, s.CmdCtx)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("failed"), err)
	})
	t.Run("Valid registration of fast serialize", func(t *testing.T) {
		s := setup()
		registerFilesSetup()
		testScope := promutils.NewTestScope()
		labeled.SetMetricKeys(contextutils.AppNameKey, contextutils.ProjectKey, contextutils.DomainKey)
		rconfig.DefaultFilesConfig.Archive = true

		rconfig.DefaultFilesConfig.OutputLocationPrefix = s3Output
		rconfig.DefaultFilesConfig.DeprecatedSourceUploadPath = s3Output
		store, err := storage.NewDataStore(&storage.Config{
			Type: storage.TypeMemory,
		}, testScope.NewSubScope("flytectl"))
		Client = store
		assert.Nil(t, err)
		args := []string{"testdata/flytesnacks-core.tgz"}
		s.MockAdminClient.OnCreateTaskMatch(mock.Anything, mock.Anything).Return(nil, nil)
		s.MockAdminClient.OnCreateWorkflowMatch(mock.Anything, mock.Anything).Return(nil, nil)
		s.MockAdminClient.OnCreateLaunchPlanMatch(mock.Anything, mock.Anything).Return(nil, nil)
		s.MockAdminClient.OnUpdateLaunchPlanMatch(mock.Anything, mock.Anything).Return(nil, nil)
		s.MockClient.DataProxyClient().(*mocks.DataProxyServiceClient).OnCreateUploadLocationMatch(mock.Anything, mock.Anything).Return(&service.CreateUploadLocationResponse{}, nil)
		err = registerFromFilesFunc(s.Ctx, args, s.CmdCtx)
		assert.Nil(t, err)
	})

	t.Run("Registration with proto files ", func(t *testing.T) {
		s := setup()
		registerFilesSetup()
		testScope := promutils.NewTestScope()
		labeled.SetMetricKeys(contextutils.AppNameKey, contextutils.ProjectKey, contextutils.DomainKey)
		rconfig.DefaultFilesConfig.Archive = false
		rconfig.DefaultFilesConfig.OutputLocationPrefix = s3Output
		rconfig.DefaultFilesConfig.DeprecatedSourceUploadPath = ""
		store, err := storage.NewDataStore(&storage.Config{
			Type: storage.TypeMemory,
		}, testScope.NewSubScope("flytectl"))
		Client = store
		assert.Nil(t, err)
		args := []string{"testdata/69_core.flyte_basics.lp.greet_1.pb"}
		s.MockAdminClient.OnCreateTaskMatch(mock.Anything, mock.Anything).Return(nil, nil)
		s.MockAdminClient.OnCreateWorkflowMatch(mock.Anything, mock.Anything).Return(nil, nil)
		s.MockAdminClient.OnCreateLaunchPlanMatch(mock.Anything, mock.Anything).Return(nil, nil)
		s.MockClient.DataProxyClient().(*mocks.DataProxyServiceClient).OnCreateUploadLocationMatch(mock.Anything, mock.Anything).Return(&service.CreateUploadLocationResponse{}, nil)
		err = registerFromFilesFunc(s.Ctx, args, s.CmdCtx)
		assert.Nil(t, err)
	})
}
