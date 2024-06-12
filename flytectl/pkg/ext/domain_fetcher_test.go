package ext

import (
	"testing"

	"github.com/flyteorg/flyte/flyteidl/clients/go/admin/mocks"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAdminFetcherExtClient_GetDomains(t *testing.T) {
	domain1 := &admin.Domain{
		Id:   "development",
		Name: "development",
	}
	domain2 := &admin.Domain{
		Id:   "staging",
		Name: "staging",
	}
	domain3 := &admin.Domain{
		Id:   "production",
		Name: "production",
	}
	domains := &admin.GetDomainsResponse{
		Domains: []*admin.Domain{domain1, domain2, domain3},
	}

	adminClient := new(mocks.AdminServiceClient)
	adminFetcherExt := AdminFetcherExtClient{AdminClient: adminClient}

	adminClient.OnGetDomainsMatch(mock.Anything, mock.Anything).Return(domains, nil)
	_, err := adminFetcherExt.GetDomains(ctx)
	assert.Nil(t, err)
}
