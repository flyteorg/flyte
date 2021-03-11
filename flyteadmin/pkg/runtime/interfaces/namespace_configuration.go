package interfaces

import "github.com/flyteorg/flyteadmin/pkg/common"

type NamespaceMappingConfig struct {
	Mapping string `json:"mapping"`
}

type NamespaceMappingConfiguration interface {
	GetNamespaceMappingConfig() common.NamespaceMapping
}
