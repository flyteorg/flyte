package register

import (
	"fmt"
	"testing"

	"github.com/flyteorg/flytestdlib/contextutils"
	"github.com/flyteorg/flytestdlib/promutils/labeled"
	"github.com/flyteorg/flytestdlib/storage"

	rconfig "github.com/flyteorg/flytectl/cmd/config/subcommand/register"
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	s3Output = "s3://dummy/prefix"
)

func TestRegisterFromFiles(t *testing.T) {
	t.Run("Valid registration", func(t *testing.T) {
		setup()
		registerFilesSetup()
		rconfig.DefaultFilesConfig.Archive = true
		args = []string{"testdata/valid-parent-folder-register.tar"}
		mockAdminClient.OnCreateTaskMatch(mock.Anything, mock.Anything).Return(nil, nil)
		mockAdminClient.OnCreateWorkflowMatch(mock.Anything, mock.Anything).Return(nil, nil)
		mockAdminClient.OnCreateLaunchPlanMatch(mock.Anything, mock.Anything).Return(nil, nil)
		err := registerFromFilesFunc(ctx, args, cmdCtx)
		assert.Nil(t, err)
	})
	t.Run("Valid fast registration", func(t *testing.T) {
		setup()
		testScope := promutils.NewTestScope()
		labeled.SetMetricKeys(contextutils.AppNameKey, contextutils.ProjectKey, contextutils.DomainKey)
		registerFilesSetup()
		rconfig.DefaultFilesConfig.Archive = true
		rconfig.DefaultFilesConfig.OutputLocationPrefix = s3Output
		rconfig.DefaultFilesConfig.SourceUploadPath = s3Output
		mockStorage, err := storage.NewDataStore(&storage.Config{
			Type: storage.TypeMemory,
		}, testScope.NewSubScope("flytectl"))
		assert.Nil(t, err)
		Client = mockStorage

		args = []string{"testdata/flytesnacks-core.tgz"}
		mockAdminClient.OnCreateTaskMatch(mock.Anything, mock.Anything).Return(nil, nil)
		mockAdminClient.OnCreateWorkflowMatch(mock.Anything, mock.Anything).Return(nil, nil)
		mockAdminClient.OnCreateLaunchPlanMatch(mock.Anything, mock.Anything).Return(nil, nil)

		err = registerFromFilesFunc(ctx, args, cmdCtx)
		assert.Nil(t, err)
	})
	t.Run("Failed fast registration while uploading the codebase", func(t *testing.T) {
		setup()
		registerFilesSetup()
		testScope := promutils.NewTestScope()
		labeled.SetMetricKeys(contextutils.AppNameKey, contextutils.ProjectKey, contextutils.DomainKey)
		rconfig.DefaultFilesConfig.Archive = true
		rconfig.DefaultFilesConfig.OutputLocationPrefix = s3Output
		s, err := storage.NewDataStore(&storage.Config{
			Type: storage.TypeMemory,
		}, testScope.NewSubScope("flytectl"))
		assert.Nil(t, err)
		Client = s
		args = []string{"testdata/flytesnacks-core.tgz"}
		mockAdminClient.OnCreateTaskMatch(mock.Anything, mock.Anything).Return(nil, nil)
		mockAdminClient.OnCreateWorkflowMatch(mock.Anything, mock.Anything).Return(nil, nil)
		mockAdminClient.OnCreateLaunchPlanMatch(mock.Anything, mock.Anything).Return(nil, nil)
		err = Register(ctx, args, cmdCtx)
		assert.Nil(t, err)
	})
	t.Run("Failed registration because of invalid files", func(t *testing.T) {
		setup()
		registerFilesSetup()
		testScope := promutils.NewTestScope()
		labeled.SetMetricKeys(contextutils.AppNameKey, contextutils.ProjectKey, contextutils.DomainKey)
		rconfig.DefaultFilesConfig.Archive = true
		rconfig.DefaultFilesConfig.SourceUploadPath = ""
		s, err := storage.NewDataStore(&storage.Config{
			Type: storage.TypeMemory,
		}, testScope.NewSubScope("flytectl"))
		Client = s
		assert.Nil(t, err)
		args = []string{"testdata/invalid-fast.tgz"}
		mockAdminClient.OnCreateTaskMatch(mock.Anything, mock.Anything).Return(nil, nil)
		mockAdminClient.OnCreateWorkflowMatch(mock.Anything, mock.Anything).Return(nil, nil)
		mockAdminClient.OnCreateLaunchPlanMatch(mock.Anything, mock.Anything).Return(nil, nil)
		err = registerFromFilesFunc(ctx, args, cmdCtx)
		assert.NotNil(t, err)
	})
	t.Run("Failure registration of fast serialize", func(t *testing.T) {
		setup()
		registerFilesSetup()
		testScope := promutils.NewTestScope()
		labeled.SetMetricKeys(contextutils.AppNameKey, contextutils.ProjectKey, contextutils.DomainKey)
		rconfig.DefaultFilesConfig.Archive = true

		rconfig.DefaultFilesConfig.OutputLocationPrefix = s3Output
		rconfig.DefaultFilesConfig.SourceUploadPath = s3Output
		s, err := storage.NewDataStore(&storage.Config{
			Type: storage.TypeMemory,
		}, testScope.NewSubScope("flytectl"))
		Client = s
		assert.Nil(t, err)
		args = []string{"testdata/flytesnacks-core.tgz"}
		mockAdminClient.OnCreateTaskMatch(mock.Anything, mock.Anything).Return(nil, fmt.Errorf("failed"))
		mockAdminClient.OnCreateWorkflowMatch(mock.Anything, mock.Anything).Return(nil, fmt.Errorf("failed"))
		mockAdminClient.OnCreateLaunchPlanMatch(mock.Anything, mock.Anything).Return(nil, fmt.Errorf("failed"))
		err = registerFromFilesFunc(ctx, args, cmdCtx)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("failed"), err)
	})
	t.Run("Valid registration of fast serialize", func(t *testing.T) {
		setup()
		registerFilesSetup()
		testScope := promutils.NewTestScope()
		labeled.SetMetricKeys(contextutils.AppNameKey, contextutils.ProjectKey, contextutils.DomainKey)
		rconfig.DefaultFilesConfig.Archive = true

		rconfig.DefaultFilesConfig.OutputLocationPrefix = s3Output
		rconfig.DefaultFilesConfig.SourceUploadPath = s3Output
		s, err := storage.NewDataStore(&storage.Config{
			Type: storage.TypeMemory,
		}, testScope.NewSubScope("flytectl"))
		Client = s
		assert.Nil(t, err)
		args = []string{"testdata/flytesnacks-core.tgz"}
		mockAdminClient.OnCreateTaskMatch(mock.Anything, mock.Anything).Return(nil, nil)
		mockAdminClient.OnCreateWorkflowMatch(mock.Anything, mock.Anything).Return(nil, nil)
		mockAdminClient.OnCreateLaunchPlanMatch(mock.Anything, mock.Anything).Return(nil, nil)
		err = registerFromFilesFunc(ctx, args, cmdCtx)
		assert.Nil(t, err)
	})

	t.Run("Registration with proto files ", func(t *testing.T) {
		setup()
		registerFilesSetup()
		testScope := promutils.NewTestScope()
		labeled.SetMetricKeys(contextutils.AppNameKey, contextutils.ProjectKey, contextutils.DomainKey)
		rconfig.DefaultFilesConfig.Archive = false
		rconfig.DefaultFilesConfig.OutputLocationPrefix = s3Output
		rconfig.DefaultFilesConfig.SourceUploadPath = ""
		s, err := storage.NewDataStore(&storage.Config{
			Type: storage.TypeMemory,
		}, testScope.NewSubScope("flytectl"))
		Client = s
		assert.Nil(t, err)
		args = []string{"testdata/69_core.flyte_basics.lp.greet_1.pb"}
		mockAdminClient.OnCreateTaskMatch(mock.Anything, mock.Anything).Return(nil, nil)
		mockAdminClient.OnCreateWorkflowMatch(mock.Anything, mock.Anything).Return(nil, nil)
		mockAdminClient.OnCreateLaunchPlanMatch(mock.Anything, mock.Anything).Return(nil, nil)
		err = registerFromFilesFunc(ctx, args, cmdCtx)
		assert.Nil(t, err)
	})
}
