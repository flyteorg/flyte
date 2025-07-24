package ext

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyte/flyteidl/clients/go/admin/mocks"
)

var deleterFetcherClient *AdminDeleterExtClient

func TestAdminDeleterExtClient_AdminServiceClient(t *testing.T) {
	adminClient = new(mocks.AdminServiceClient)
	deleterFetcherClient = nil
	client := deleterFetcherClient.AdminServiceClient()
	assert.Nil(t, client)
}
