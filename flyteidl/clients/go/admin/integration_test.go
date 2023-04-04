//go:build integration
// +build integration

package admin

import (
	"context"
	"fmt"
	"net/url"
	"testing"
	"time"

	"google.golang.org/grpc"

	"golang.org/x/oauth2/clientcredentials"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flytestdlib/config"
	"github.com/stretchr/testify/assert"
)

func TestLiveAdminClient(t *testing.T) {
	ctx := context.Background()

	u, err := url.Parse("dns:///flyte.lyft.net")
	assert.NoError(t, err)
	client := InitializeAdminClient(ctx, Config{
		Endpoint:                         config.URL{URL: *u},
		UseInsecureConnection:            false,
		UseAuth:                          true,
		ClientID:                         "0oacmtueinpXk72Af1t7",
		ClientSecretLocation:             "/Users/username/.ssh/admin/propeller_secret",
		DeprecatedAuthorizationServerURL: "https://lyft.okta.com/oauth2/ausc5wmjw96cRKvTd1t7",
		Scopes:                           []string{"svc"},
		AuthorizationHeader:              "Flyte-Authorization",
	})

	resp, err := client.ListProjects(ctx, &admin.ProjectListRequest{})
	if err != nil {
		fmt.Printf("Error %v\n", err)
	}
	assert.NoError(t, err)
	fmt.Printf("Response: %v\n", resp)
}

func TestGetDialOption(t *testing.T) {
	ctx := context.Background()

	cfg := Config{
		DeprecatedAuthorizationServerURL: "https://lyft.okta.com/oauth2/ausc5wmjw96cRKvTd1t7",
	}

	dialOption, err := getAuthenticationDialOption(ctx, cfg, []grpc.DialOption{})
	assert.NoError(t, err)
	assert.NotNil(t, dialOption)
}

func TestDirectTokenRetrieval(t *testing.T) {
	ctx := context.Background()
	ccConfig := clientcredentials.Config{
		ClientID:     "client-id",
		ClientSecret: "my-secret",
		TokenURL:     "https://your.idp.com/authserver/v1/token",
		Scopes:       []string{"svc"},
	}

	tSource := ccConfig.TokenSource(ctx)

	for i := 0; i < 100; i++ {
		fmt.Printf("Iteration %d -- ", i)
		token, err := tSource.Token()
		assert.NoError(t, err)
		fmt.Printf("Got token %s\n", token)
		time.Sleep(30 * time.Second)
	}
}
