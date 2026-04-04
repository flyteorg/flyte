package cache_service

const (
	TaskVersionKey     = "task-version"
	ExecNameKey        = "execution-name"
	ExecDomainKey      = "exec-domain"
	ExecProjectKey     = "exec-project"
	ExecNodeIDKey      = "exec-node"
	ExecTaskAttemptKey = "exec-attempt"
	ExecOrgKey         = "exec-org"
)

func getMetadataValue(values map[string]string, key, fallback string) string {
	if value, ok := values[key]; ok && value != "" {
		return value
	}

	return fallback
}
