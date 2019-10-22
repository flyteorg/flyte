package interfaces

import "github.com/lyft/flyteadmin/pkg/common"

type NamespaceMappingConfig struct {
	Mapping string `json:"mapping"`
}

type NamespaceMappingConfiguration interface {
	GetNamespaceMappingConfig() common.NamespaceMapping
}
