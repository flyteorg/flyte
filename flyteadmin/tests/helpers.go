package tests

import (
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/service"
)

// This returns a gRPC client configured to hit the locally running instance of Flyte admin
// This also returns a gRPC connection - be sure to defer a Close() call!!!
func GetTestAdminServiceClient() (service.AdminServiceClient, *grpc.ClientConn) {
	// Load the running configuration in order to talk to the running flyteadmin instance

	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	conn, err := grpc.Dial(fmt.Sprintf("0.0.0.0:%d", 8089), opts...)
	if err != nil {
		grpclog.Fatalf("fail to dial: %v", err)
	}
	client := service.NewAdminServiceClient(conn)
	return client, conn
}

func GetTestHostEndpoint() string {
	return "http://localhost:8088"
}
