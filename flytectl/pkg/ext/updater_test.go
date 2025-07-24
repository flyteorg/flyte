package ext

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyte/flyteidl/clients/go/admin/mocks"
)

var updaterFetcherClient *AdminUpdaterExtClient

func TestAdminUpdaterExtClient_AdminServiceClient(t *testing.T) {
	adminClient = new(mocks.AdminServiceClient)
	updaterFetcherClient = nil
	client := updaterFetcherClient.AdminServiceClient()
	assert.Nil(t, client)
}
