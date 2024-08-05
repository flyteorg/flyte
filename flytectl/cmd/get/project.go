package get

import (
	"context"
	"fmt"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/service"
	"google.golang.org/grpc"

	cmdCore "github.com/flyteorg/flyte/flytectl/cmd/core"
	"github.com/flyteorg/flyte/flytectl/pkg/printer"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/golang/protobuf/proto"
)

const (
	projectShort = "Gets project resources"
	projectLong  = `
Retrieve all the projects:
::

 flytectl get project

.. note::
	  The terms project/projects are interchangeable in these commands.

Retrieve project by name:

::

 flytectl get project flytesnacks

Retrieve all the projects with filters:
::

  flytectl get project --filter.fieldSelector="project.name=flytesnacks"

Retrieve all the projects with limit and sorting:
::

  flytectl get project --filter.sortBy=created_at --filter.limit=1 --filter.asc

Retrieve projects present in other pages by specifying the limit and page number:
::

  flytectl get project --filter.limit=10 --filter.page=2

Retrieve all the projects in yaml format:

::

 flytectl get project -o yaml

Retrieve all the projects in json format:

::

 flytectl get project -o json

Usage
`
)

var projectColumns = []printer.Column{
	{Header: "ID", JSONPath: "$.id"},
	{Header: "Name", JSONPath: "$.name"},
	{Header: "Description", JSONPath: "$.description"},
}

func ProjectToProtoMessages(l []*admin.Project) []proto.Message {
	messages := make([]proto.Message, 0, len(l))
	for _, m := range l {
		messages = append(messages, m)
	}
	return messages
}

func getProjectsFunc(ctx context.Context, args []string, cmdCtx cmdCore.CommandContext) error {

	clientSet := cmdCtx.ClientSet()

	opts := []grpc.CallOption{
		grpc.MaxRecvMsgSizeCallOption{MaxRecvMsgSize: 20 * 1024 * 1024},
	}
	req := service.GetDataRequest{FlyteUrl: "flyte://v1/daniel/development/fafa413ee6d26467fbee/n0/o"}
	_, err := clientSet.DataProxyClient().GetData(ctx, &req, opts...)
	fmt.Printf("err: %v\n", err)

	return err
}
