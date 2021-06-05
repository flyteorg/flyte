package ext

import (
	"testing"

	"github.com/flyteorg/flyteidl/clients/go/admin/mocks"
	"github.com/stretchr/testify/assert"
)

var fetcherClient *AdminFetcherExtClient

func TestAdminFetcherExtClient_AdminServiceClient(t *testing.T) {
	adminClient = new(mocks.AdminServiceClient)
	fetcherClient = nil
	client := fetcherClient.AdminServiceClient()
	assert.Nil(t, client)
}
