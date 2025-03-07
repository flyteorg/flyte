package implementations

import (
	"context"
	"errors"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/sendgrid/rest"
	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyte/flyteadmin/pkg/async/notifications/mocks"
	runtimeInterfaces "github.com/flyteorg/flyte/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
)

var (
	emailNotification = &admin.EmailMessage{
		SubjectLine: "Notice: Execution \"name\" has succeeded in \"domain\".",
		SenderEmail: "no-reply@example.com",
		RecipientsEmail: []string{
			"my@example.com",
			"john@example.com",
		},
		Body: "Execution \"name\" has succeeded in \"domain\". View details at " +
			"<a href=\"https://example.com/executions/T/B/D\">" +
			"https://example.com/executions/T/B/D</a>.",
	}
)

func TestAddresses(t *testing.T) {
	addresses := []string{"alice@example.com", "bob@example.com"}
	sgAddresses := getEmailAddresses(addresses)
	assert.Equal(t, sgAddresses[0].Address, "alice@example.com")
	assert.Equal(t, sgAddresses[1].Address, "bob@example.com")
}

func TestGetEmail(t *testing.T) {
	sgEmail := getSendgridEmail(emailNotification)
	assert.Equal(t, `Notice: Execution "name" has succeeded in "domain".`, sgEmail.Personalizations[0].Subject)
	assert.Equal(t, "john@example.com", sgEmail.Personalizations[0].To[1].Address)
	assert.Equal(t, `Execution "name" has succeeded in "domain". View details at <a href="https://example.com/executions/T/B/D">https://example.com/executions/T/B/D</a>.`, sgEmail.Content[0].Value)
}

func TestCreateEmailer(t *testing.T) {
	cfg := getNotificationsConfig()
	cfg.NotificationsEmailerConfig.EmailerConfig.APIKeyEnvVar = "sendgrid_api_key"

	emailer := NewSendGridEmailer(cfg, promutils.NewTestScope())
	assert.NotNil(t, emailer)
}

func TestReadAPIFromEnv(t *testing.T) {
	envVar := "sendgrid_api_test_key"
	orig := os.Getenv(envVar)
	usingEnv := runtimeInterfaces.EmailServerConfig{
		ServiceName:  "test",
		APIKeyEnvVar: envVar,
	}
	err := os.Setenv(envVar, "test_api_key")
	assert.NoError(t, err)
	assert.Equal(t, "test_api_key", getAPIKey(usingEnv))
	_ = os.Setenv(envVar, orig)
}

func TestReadAPIFromFile(t *testing.T) {
	tmpFile, err := ioutil.TempFile(os.TempDir(), "api_key")
	if err != nil {
		t.Fatalf("Cannot create temporary file: %v", err)
	}
	defer tmpFile.Close()
	_, err = tmpFile.WriteString("abc123\n")
	if err != nil {
		t.Fatalf("Cannot write to temporary file: %v", err)
	}

	usingFile := runtimeInterfaces.EmailServerConfig{
		ServiceName:    "test",
		APIKeyFilePath: tmpFile.Name(),
	}
	assert.NoError(t, err)
	assert.Equal(t, "abc123", getAPIKey(usingFile))
}

func TestNoFile(t *testing.T) {
	defer func() { r := recover(); assert.NotNil(t, r) }()

	dir, err := ioutil.TempDir("", "test_prefix")
	if err != nil {
		t.Fatalf("Error creating random directory: %s", err)
	}
	defer os.RemoveAll(dir)
	nonExistentFile := path.Join(dir, "doesnotexist")

	doesNotExistConfig := runtimeInterfaces.EmailServerConfig{
		ServiceName:    "test",
		APIKeyFilePath: nonExistentFile,
	}
	getAPIKey(doesNotExistConfig)

	// shouldn't reach here
	t.Errorf("did not panic")
}

func TestSendEmail(t *testing.T) {
	ctx := context.TODO()
	expectedErr := errors.New("expected")
	t.Run("exhaust all retry attempts", func(t *testing.T) {
		sendgridClient := &mocks.SendgridClient{}
		expectedEmail := getSendgridEmail(emailNotification)
		sendgridClient.EXPECT().Send(expectedEmail).
			Return(nil, expectedErr).Times(3)
		sendgridClient.EXPECT().Send(expectedEmail).
			Return(&rest.Response{Body: "email body"}, nil).Once()
		scope := promutils.NewScope("bademailer")
		emailerMetrics := newEmailMetrics(scope)

		emailer := SendgridEmailer{
			client:        sendgridClient,
			systemMetrics: emailerMetrics,
			cfg: &runtimeInterfaces.NotificationsConfig{
				ReconnectAttempts: 1,
			},
		}

		err := emailer.SendEmail(ctx, emailNotification)
		assert.EqualError(t, err, expectedErr.Error())

		assert.NoError(t, testutil.CollectAndCompare(emailerMetrics.SendError, strings.NewReader(`
		# HELP bademailer:send_error Number of errors when sending email via Emailer
		# TYPE bademailer:send_error counter
		bademailer:send_error 1
		`)))
	})
	t.Run("exhaust all retry attempts", func(t *testing.T) {
		ctx := context.TODO()
		sendgridClient := &mocks.SendgridClient{}
		expectedEmail := getSendgridEmail(emailNotification)
		sendgridClient.EXPECT().Send(expectedEmail).
			Return(nil, expectedErr).Once()
		sendgridClient.EXPECT().Send(expectedEmail).
			Return(&rest.Response{Body: "email body"}, nil).Once()
		scope := promutils.NewScope("goodemailer")
		emailerMetrics := newEmailMetrics(scope)

		emailer := SendgridEmailer{
			client:        sendgridClient,
			systemMetrics: emailerMetrics,
			cfg: &runtimeInterfaces.NotificationsConfig{
				ReconnectAttempts: 1,
			},
		}

		err := emailer.SendEmail(ctx, emailNotification)
		assert.NoError(t, err)

		assert.NoError(t, testutil.CollectAndCompare(emailerMetrics.SendError, strings.NewReader(`
		# HELP goodemailer:send_error Number of errors when sending email via Emailer
		# TYPE goodemailer:send_error counter
		goodemailer:send_error 0
		`)))
	})
}
