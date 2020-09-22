package cmdcore

import (
	"io"

	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/service"
)

type CommandContext struct {
	adminClient service.AdminServiceClient
	in          io.Reader
	out         io.Writer
}

func (c CommandContext) AdminClient() service.AdminServiceClient {
	return c.adminClient
}

func (c CommandContext) OutputPipe() io.Writer {
	return c.out
}
