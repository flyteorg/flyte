package errors

type DocKey int

const (
	ResourceLimitError DocKey = iota
)

var DocLinks = map[DocKey]string{
	ResourceLimitError: "https://docs.union.ai/byoc/user-guide/core-concepts/tasks/task-hardware-environment/customizing-task-resources#execution-defaults-and-resource-quotas",
}

func GetDocLink(key DocKey) string {
	if link, exists := DocLinks[key]; exists {
		return link
	}
	return "https://docs.union.ai"
}
