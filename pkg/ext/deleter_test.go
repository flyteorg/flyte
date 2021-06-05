package ext

import (
	"testing"

	"github.com/flyteorg/flyteidl/clients/go/admin/mocks"
	"github.com/stretchr/testify/assert"
)

var deleterFetcherClient *AdminDeleterExtClient

func TestAdminDeleterExtClient_AdminServiceClient(t *testing.T) {
	adminClient = new(mocks.AdminServiceClient)
	deleterFetcherClient = nil
	client := deleterFetcherClient.AdminServiceClient()
	assert.Nil(t, client)
}
