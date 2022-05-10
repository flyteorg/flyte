//go:build integration
// +build integration

package entrypoints

// This is an integration test because the token will show up as expired, you will need a live token

import (
	"context"
	"fmt"
	"testing"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/service"
	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/oauth"
)

func TestClient(t *testing.T) {
	ctx := context.Background()
	endpoint := "localhost:8088"

	var opts []grpc.DialOption

	creds, err := credentials.NewClientTLSFromFile("/path/to/server.pem", ":8088")
	assert.NoError(t, err)
	opts = append(opts, grpc.WithTransportCredentials(creds))

	token := oauth2.Token{
		AccessToken: "j.w.t",
	}
	tokenRpcCredentials := oauth.NewOauthAccess(&token)
	tokenDialOption := grpc.WithPerRPCCredentials(tokenRpcCredentials)
	opts = append(opts, tokenDialOption)

	conn, err := grpc.Dial(endpoint, opts...)
	if err != nil {
		fmt.Printf("Dial error %v\n", err)
	}
	assert.NoError(t, err)
	client := service.NewAdminServiceClient(conn)
	resp, err := client.ListProjects(ctx, &admin.ProjectListRequest{})
	if err != nil {
		fmt.Printf("Error %v\n", err)
	}
	assert.NoError(t, err)
	fmt.Printf("Response: %v\n", resp)
}
