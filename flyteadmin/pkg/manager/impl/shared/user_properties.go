package shared

import "context"

type UserProperties struct {
	Org string
	// Limit on number of active launch plans for this org
	ActiveLaunchPlans int
	// Limit on number of active executions for this org
	ActiveExecutions int
}

var defaultUserProperties = UserProperties{}

type GetUserProperties = func(ctx context.Context) UserProperties

func DefaultGetUserPropertiesFunc(_ context.Context) UserProperties {
	return defaultUserProperties
}
