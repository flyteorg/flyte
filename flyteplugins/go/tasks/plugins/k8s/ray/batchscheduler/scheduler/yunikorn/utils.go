package yunikorn

import (
	"fmt"
)

const (
	TaskGroupGenericName = "task-group"
)

func GenerateTaskGroupName(master bool, index int) string {
	if master {
		return fmt.Sprintf("%s-%s", TaskGroupGenericName, "head")
	}
	return fmt.Sprintf("%s-%s-%d", TaskGroupGenericName, "worker", index)
}
