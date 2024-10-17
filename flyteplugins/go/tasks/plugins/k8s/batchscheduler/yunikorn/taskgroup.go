package yunikorn

import (
	"encoding/json"

	v1 "k8s.io/api/core/v1"
)

type TaskGroup struct {
	Name                      string
	MinMember                 int32
	Labels                    map[string]string
	Annotations               map[string]string
	MinResource               v1.ResourceList
	NodeSelector              map[string]string
	Tolerations               []v1.Toleration
	Affinity                  *v1.Affinity
	TopologySpreadConstraints []v1.TopologySpreadConstraint
}

func Marshal(taskGroups []TaskGroup) ([]byte, error) {
	info, err := json.Marshal(taskGroups)
	return info, err
}
