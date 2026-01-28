package secret

import (
	"context"
	"fmt"
	"testing"

	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/secret/config"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/secret/mocks"
	stdlibErrors "github.com/flyteorg/flyte/v2/flytestdlib/errors"
	"github.com/flyteorg/flyte/v2/flytestdlib/promutils"
)

var (
	gcpClient  *mocks.GCPSecretManagerClient
	gcpProject string
)

func SetupGCPTest() {
	scope = promutils.NewTestScope()
	ctx = context.Background()
	gcpClient = &mocks.GCPSecretManagerClient{}
	gcpProject = "project"
}

func TestGetSecretValueGCP(t *testing.T) {
	t.Run("get secret successful", func(t *testing.T) {
		SetupGCPTest()
		gcpSecretsFetcher := NewGCPSecretFetcher(config.GCPConfig{
			Project: gcpProject,
		}, gcpClient)
		gcpClient.OnAccessSecretVersionMatch(ctx, &secretmanagerpb.AccessSecretVersionRequest{
			Name: fmt.Sprintf(GCPSecretNameFormat, gcpProject, secretID),
		}).Return(&secretmanagerpb.AccessSecretVersionResponse{
			Payload: &secretmanagerpb.SecretPayload{
				Data: []byte("secretValue"),
			},
		}, nil)

		_, err := gcpSecretsFetcher.GetSecretValue(ctx, "secretID")
		assert.NoError(t, err)
	})

	t.Run("get secret not found", func(t *testing.T) {
		SetupGCPTest()
		gcpSecretsFetcher := NewGCPSecretFetcher(config.GCPConfig{
			Project: gcpProject,
		}, gcpClient)
		cause := status.Errorf(codes.NotFound, "secret not found")
		gcpClient.OnAccessSecretVersionMatch(ctx, &secretmanagerpb.AccessSecretVersionRequest{
			Name: fmt.Sprintf(GCPSecretNameFormat, gcpProject, secretID),
		}).Return(nil, cause)

		_, err := gcpSecretsFetcher.GetSecretValue(ctx, "secretID")
		assert.Equal(t, stdlibErrors.Wrapf(ErrCodeSecretNotFound, cause, fmt.Sprintf(SecretNotFoundErrorFormat, secretID)), err)
	})

	t.Run("get secret read failure", func(t *testing.T) {
		SetupGCPTest()
		gcpSecretsFetcher := NewGCPSecretFetcher(config.GCPConfig{
			Project: gcpProject,
		}, gcpClient)
		cause := fmt.Errorf("some error")
		gcpClient.OnAccessSecretVersionMatch(ctx, &secretmanagerpb.AccessSecretVersionRequest{
			Name: fmt.Sprintf(GCPSecretNameFormat, gcpProject, secretID),
		}).Return(nil, cause)

		_, err := gcpSecretsFetcher.GetSecretValue(ctx, "secretID")
		assert.Equal(t, stdlibErrors.Wrapf(ErrCodeSecretReadFailure, cause, fmt.Sprintf(SecretReadFailureErrorFormat, secretID)), err)
	})
}
