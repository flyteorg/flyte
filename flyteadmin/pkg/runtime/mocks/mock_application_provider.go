package mocks

import (
	"github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flytestdlib/database"
)

type MockApplicationProvider struct {
	dbConfig             database.DbConfig
	topLevelConfig       interfaces.ApplicationConfig
	schedulerConfig      interfaces.SchedulerConfig
	remoteDataConfig     interfaces.RemoteDataConfig
	notificationsConfig  interfaces.NotificationsConfig
	domainsConfig        interfaces.DomainsConfig
	externalEventsConfig interfaces.ExternalEventsConfig
	cloudEventConfig     interfaces.CloudEventsConfig
}

func (p *MockApplicationProvider) GetDbConfig() *database.DbConfig {
	return &p.dbConfig
}

func (p *MockApplicationProvider) SetDbConfig(dbConfig database.DbConfig) {
	p.dbConfig = dbConfig
}

func (p *MockApplicationProvider) GetTopLevelConfig() *interfaces.ApplicationConfig {
	return &p.topLevelConfig
}

func (p *MockApplicationProvider) SetTopLevelConfig(topLevelConfig interfaces.ApplicationConfig) {
	p.topLevelConfig = topLevelConfig
}

func (p *MockApplicationProvider) GetSchedulerConfig() *interfaces.SchedulerConfig {
	return &p.schedulerConfig
}

func (p *MockApplicationProvider) SetSchedulerConfig(schedulerConfig interfaces.SchedulerConfig) {
	p.schedulerConfig = schedulerConfig
}

func (p *MockApplicationProvider) GetRemoteDataConfig() *interfaces.RemoteDataConfig {
	return &p.remoteDataConfig
}

func (p *MockApplicationProvider) SetRemoteDataConfig(remoteDataConfig interfaces.RemoteDataConfig) {
	p.remoteDataConfig = remoteDataConfig
}

func (p *MockApplicationProvider) GetNotificationsConfig() *interfaces.NotificationsConfig {
	return &p.notificationsConfig
}

func (p *MockApplicationProvider) SetNotificationsConfig(notificationsConfig interfaces.NotificationsConfig) {
	p.notificationsConfig = notificationsConfig
}

func (p *MockApplicationProvider) GetDomainsConfig() *interfaces.DomainsConfig {
	return &p.domainsConfig
}

func (p *MockApplicationProvider) SetDomainsConfig(domainsConfig interfaces.DomainsConfig) {
	p.domainsConfig = domainsConfig
}

func (p *MockApplicationProvider) SetExternalEventsConfig(externalEventsConfig interfaces.ExternalEventsConfig) {
	p.externalEventsConfig = externalEventsConfig
}

func (p *MockApplicationProvider) GetExternalEventsConfig() *interfaces.ExternalEventsConfig {
	return &p.externalEventsConfig
}

func (p *MockApplicationProvider) SetCloudEventsConfig(cloudEventConfig interfaces.CloudEventsConfig) {
	p.cloudEventConfig = cloudEventConfig
}

func (p *MockApplicationProvider) GetCloudEventsConfig() *interfaces.CloudEventsConfig {
	return &p.cloudEventConfig
}
