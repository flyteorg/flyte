package interfaces

type WhitelistScope struct {
	Project string `json:"project"`
	Domain  string `json:"domain"`
}

// Defines specific task types whitelisted for support.
type TaskTypeWhitelist = map[string][]WhitelistScope

type WhitelistConfiguration interface {
	// Returns whitelisted task types defined in runtime configuration files.
	GetTaskTypeWhitelist() TaskTypeWhitelist
}
