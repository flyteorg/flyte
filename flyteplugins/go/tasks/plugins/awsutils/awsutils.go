package awsutils

import (
	core2 "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
)

func GetRoleFromSecurityContext(roleKey string, taskExecutionMetadata core2.TaskExecutionMetadata) string {
	var role string
	securityContext := taskExecutionMetadata.GetSecurityContext()
	if securityContext.GetRunAs() != nil {
		role = securityContext.GetRunAs().GetIamRole()
	}

	// Continue this for backward compatibility
	if len(role) == 0 {
		role = getRole(roleKey, taskExecutionMetadata.GetAnnotations())
	}
	return role
}

func getRole(roleKey string, keyValueMap map[string]string) string {
	if len(roleKey) > 0 {
		return keyValueMap[roleKey]
	}

	return ""
}
