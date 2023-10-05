package testutils

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/flyteorg/flyteidl/clients/go/admin/mocks"

	"github.com/flyteorg/flyteidl/clients/go/admin"

	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flytectl/cmd/config"
	cmdCore "github.com/flyteorg/flytectl/cmd/core"
	extMocks "github.com/flyteorg/flytectl/pkg/ext/mocks"
)

const projectValue = "dummyProject"
const domainValue = "dummyDomain"
const output = "json"

type TestStruct struct {
	Reader          *os.File
	Writer          *os.File
	Err             error
	Ctx             context.Context
	MockClient      *admin.Clientset
	MockAdminClient *mocks.AdminServiceClient
	FetcherExt      *extMocks.AdminFetcherExtInterface
	UpdaterExt      *extMocks.AdminUpdaterExtInterface
	DeleterExt      *extMocks.AdminDeleterExtInterface
	MockOutStream   io.Writer
	CmdCtx          cmdCore.CommandContext
	StdOut          *os.File
	Stderr          *os.File
}

func Setup() (s TestStruct) {
	s.Ctx = context.Background()
	s.Reader, s.Writer, s.Err = os.Pipe()
	if s.Err != nil {
		panic(s.Err)
	}
	s.StdOut = os.Stdout
	s.Stderr = os.Stderr
	os.Stdout = s.Writer
	os.Stderr = s.Writer
	log.SetOutput(s.Writer)
	s.MockClient = admin.InitializeMockClientset()
	s.FetcherExt = new(extMocks.AdminFetcherExtInterface)
	s.UpdaterExt = new(extMocks.AdminUpdaterExtInterface)
	s.DeleterExt = new(extMocks.AdminDeleterExtInterface)
	s.FetcherExt.OnAdminServiceClient().Return(s.MockClient.AdminClient())
	s.UpdaterExt.OnAdminServiceClient().Return(s.MockClient.AdminClient())
	s.DeleterExt.OnAdminServiceClient().Return(s.MockClient.AdminClient())
	s.MockAdminClient = s.MockClient.AdminClient().(*mocks.AdminServiceClient)
	s.MockOutStream = s.Writer
	s.CmdCtx = cmdCore.NewCommandContextWithExt(s.MockClient, s.FetcherExt, s.UpdaterExt, s.DeleterExt, s.MockOutStream)
	config.GetConfig().Project = projectValue
	config.GetConfig().Domain = domainValue
	config.GetConfig().Output = output

	return s
}

func SetupWithExt() (s TestStruct) {
	s.Ctx = context.Background()
	s.Reader, s.Writer, s.Err = os.Pipe()
	if s.Err != nil {
		panic(s.Err)
	}
	s.StdOut = os.Stdout
	s.Stderr = os.Stderr
	os.Stdout = s.Writer
	os.Stderr = s.Writer
	log.SetOutput(s.Writer)
	s.MockClient = admin.InitializeMockClientset()
	s.FetcherExt = new(extMocks.AdminFetcherExtInterface)
	s.UpdaterExt = new(extMocks.AdminUpdaterExtInterface)
	s.DeleterExt = new(extMocks.AdminDeleterExtInterface)
	s.FetcherExt.OnAdminServiceClient().Return(s.MockClient.AdminClient())
	s.UpdaterExt.OnAdminServiceClient().Return(s.MockClient.AdminClient())
	s.DeleterExt.OnAdminServiceClient().Return(s.MockClient.AdminClient())
	s.MockAdminClient = s.MockClient.AdminClient().(*mocks.AdminServiceClient)
	s.MockOutStream = s.Writer
	s.CmdCtx = cmdCore.NewCommandContextWithExt(s.MockClient, s.FetcherExt, s.UpdaterExt, s.DeleterExt, s.MockOutStream)
	config.GetConfig().Project = projectValue
	config.GetConfig().Domain = domainValue
	config.GetConfig().Output = output

	return s
}

// TearDownAndVerify TODO: Change this to verify log lines from context
func (s *TestStruct) TearDownAndVerify(t *testing.T, expectedLog string) {
	if err := s.Writer.Close(); err != nil {
		panic(fmt.Errorf("could not close test context writer: %w", err))
	}

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, s.Reader); err != nil {
		panic(fmt.Errorf("could not read from test context reader: %w", err))
	}

	assert.Equal(t, sanitizeString(expectedLog), sanitizeString(buf.String()))
}

func sanitizeString(str string) string {
	// Not the most comprehensive ANSI pattern, but this should capture common color operations
	// such as \x1b[107;0m and \x1b[0m. Expand if needed (insert regex 2 problems joke here).
	ansiRegex := regexp.MustCompile("\u001B\\[[\\d+\\;]*\\d+m")
	replacer := strings.NewReplacer(
		"\n", "",
		"\t", "",
	)

	str = replacer.Replace(str)
	str = ansiRegex.ReplaceAllString(str, "")
	str = strings.Trim(str, " ")

	return str
}
