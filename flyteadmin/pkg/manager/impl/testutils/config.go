package testutils

import (
	"github.com/flyteorg/flyteadmin/pkg/common"
	runtimeInterfaces "github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	runtimeMocks "github.com/flyteorg/flyteadmin/pkg/runtime/mocks"
)

func GetApplicationConfigWithDefaultDomains() runtimeInterfaces.ApplicationConfiguration {
	config := runtimeMocks.MockApplicationProvider{}
	config.SetDomainsConfig(runtimeInterfaces.DomainsConfig{
		{
			ID:   "development",
			Name: "development",
		},
		{
			ID:   "staging",
			Name: "staging",
		},
		{
			ID:   "production",
			Name: "production",
		},
		{
			ID:   "domain",
			Name: "domain",
		},
	})
	config.SetRemoteDataConfig(runtimeInterfaces.RemoteDataConfig{
		Scheme: common.Local, SignedURL: runtimeInterfaces.SignedURL{
			Enabled: true,
		}})
	return &config
}
