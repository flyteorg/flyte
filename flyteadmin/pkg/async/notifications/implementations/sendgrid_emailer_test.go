package implementations

import (
	"io/ioutil"
	"os"
	"path"
	"testing"

	runtimeInterfaces "github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/stretchr/testify/assert"
)

func TestAddresses(t *testing.T) {
	addresses := []string{"alice@example.com", "bob@example.com"}
	sgAddresses := getEmailAddresses(addresses)
	assert.Equal(t, sgAddresses[0].Address, "alice@example.com")
	assert.Equal(t, sgAddresses[1].Address, "bob@example.com")
}

func TestGetEmail(t *testing.T) {
	emailNotification := admin.EmailMessage{
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
