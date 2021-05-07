package testutils

import (
	"bytes"
	"context"
	"io"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/flyteorg/flytectl/cmd/config"
	cmdCore "github.com/flyteorg/flytectl/cmd/core"
	"github.com/flyteorg/flytectl/pkg/ext"
	"github.com/flyteorg/flyteidl/clients/go/admin/mocks"

	"github.com/stretchr/testify/assert"
)

const projectValue = "dummyProject"
const domainValue = "dummyDomain"
const output = "json"

var (
	reader        *os.File
	writer        *os.File
	Err           error
	Ctx           context.Context
	MockClient    *mocks.AdminServiceClient
	FetcherExt    ext.AdminFetcherExtInterface
	mockOutStream io.Writer
	CmdCtx        cmdCore.CommandContext
	stdOut        *os.File
	stderr        *os.File
)

func Setup() {
	Ctx = context.Background()
	reader, writer, Err = os.Pipe()
	if Err != nil {
		panic(Err)
	}
	stdOut = os.Stdout
	stderr = os.Stderr
	os.Stdout = writer
	os.Stderr = writer
	log.SetOutput(writer)
	MockClient = new(mocks.AdminServiceClient)
	mockOutStream = writer
	CmdCtx = cmdCore.NewCommandContext(MockClient, mockOutStream)
	config.GetConfig().Project = projectValue
	config.GetConfig().Domain = domainValue
	config.GetConfig().Output = output
}

func TearDownAndVerify(t *testing.T, expectedLog string) {
	writer.Close()
	os.Stdout = stdOut
	os.Stderr = stderr
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, reader); err == nil {
		assert.Equal(t, santizeString(expectedLog), santizeString(buf.String()))
	}
}

func santizeString(str string) string {
	return strings.Trim(strings.ReplaceAll(strings.ReplaceAll(str, "\n", ""), "\t", ""), " \t")
}
