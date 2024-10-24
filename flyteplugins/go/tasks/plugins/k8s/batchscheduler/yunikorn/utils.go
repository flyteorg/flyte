package yunikorn

import (
	"fmt"

	"github.com/google/uuid"
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

func GenerateTaskGroupAppID() string {
	uid := uuid.New().String()
	return fmt.Sprintf("%s-%s", TaskGroupGenericName, uid)
}
