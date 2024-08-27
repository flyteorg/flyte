package yunikorn

import (
	"fmt"

	"github.com/google/uuid"
)

const (
	TaskGroupGenericName = "task-group"
)

func GenerateTaskGroupName(master bool, index int) string {
	uid := uuid.New().String()
	if master {
		return fmt.Sprintf("%s-%s-%s", TaskGroupGenericName, "head", uid)
	}
	return fmt.Sprintf("%s-%s-%d-%s", TaskGroupGenericName, "worker", index, uid)
}
