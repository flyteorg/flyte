package ext

import (
	"testing"

	"github.com/flyteorg/flyteidl/clients/go/admin/mocks"
	"github.com/stretchr/testify/assert"
)

var updaterFetcherClient *AdminUpdaterExtClient

func TestAdminUpdaterExtClient_AdminServiceClient(t *testing.T) {
	adminClient = new(mocks.AdminServiceClient)
	updaterFetcherClient = nil
	client := updaterFetcherClient.AdminServiceClient()
	assert.Nil(t, client)
}
