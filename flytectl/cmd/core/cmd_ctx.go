package cmdcore

import (
	"io"

	"github.com/flyteorg/flyteidl/clients/go/admin"

	"github.com/flyteorg/flytectl/pkg/ext"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/service"
)

type CommandContext struct {
	clientSet             *admin.Clientset
	adminClientFetcherExt ext.AdminFetcherExtInterface
	adminClientUpdateExt  ext.AdminUpdaterExtInterface
	adminClientDeleteExt  ext.AdminDeleterExtInterface
	in                    io.Reader
	out                   io.Writer
}

// NewCommandContextNoClient returns a new commandContext
func NewCommandContextNoClient(out io.Writer) CommandContext {
	return NewCommandContext(nil, out)
}

func NewCommandContext(clientSet *admin.Clientset, out io.Writer) CommandContext {
	var adminClient service.AdminServiceClient
	if clientSet != nil {
		adminClient = clientSet.AdminClient()
	}
	return NewCommandContextWithExt(
		clientSet,
		&ext.AdminFetcherExtClient{AdminClient: adminClient},
		&ext.AdminUpdaterExtClient{AdminClient: adminClient},
		&ext.AdminDeleterExtClient{AdminClient: adminClient},
		out,
	)
}

// NewCommandContextWithExt construct command context with injected extensions. Helps in injecting mocked ones for testing.
func NewCommandContextWithExt(
	clientSet *admin.Clientset,
	fetcher ext.AdminFetcherExtInterface,
	updater ext.AdminUpdaterExtInterface,
	deleter ext.AdminDeleterExtInterface,
	out io.Writer) CommandContext {
	return CommandContext{
		clientSet:             clientSet,
		out:                   out,
		adminClientFetcherExt: fetcher,
		adminClientUpdateExt:  updater,
		adminClientDeleteExt:  deleter,
	}
}

func (c CommandContext) AdminClient() service.AdminServiceClient {
	return c.clientSet.AdminClient()
}

func (c CommandContext) ClientSet() *admin.Clientset {
	return c.clientSet
}

func (c CommandContext) OutputPipe() io.Writer {
	return c.out
}

func (c CommandContext) InputPipe() io.Reader {
	return c.in
}

func (c CommandContext) AdminFetcherExt() ext.AdminFetcherExtInterface {
	return c.adminClientFetcherExt
}

func (c CommandContext) AdminUpdaterExt() ext.AdminUpdaterExtInterface {
	return c.adminClientUpdateExt
}

func (c CommandContext) AdminDeleterExt() ext.AdminDeleterExtInterface {
	return c.adminClientDeleteExt
}
