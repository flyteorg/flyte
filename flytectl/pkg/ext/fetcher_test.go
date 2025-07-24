package ext

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyte/flyteidl/clients/go/admin/mocks"
)

var fetcherClient *AdminFetcherExtClient

func TestAdminFetcherExtClient_AdminServiceClient(t *testing.T) {
	adminClient = new(mocks.AdminServiceClient)
	fetcherClient = nil
	client := fetcherClient.AdminServiceClient()
	assert.Nil(t, client)
}
