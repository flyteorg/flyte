package ext

import (
	"context"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
)

func (a *AdminFetcherExtClient) GetDomains(ctx context.Context) (*admin.GetDomainsResponse, error) {
	domains, err := a.AdminServiceClient().GetDomains(ctx, &admin.GetDomainRequest{})
	if err != nil {
		return nil, err
	}
	return domains, nil
}
