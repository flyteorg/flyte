package k8s

import (
	"math"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/golang/protobuf/ptypes"
)

func refInt(i int) *int {
	return &i
}

func refStr(s string) *string {
	return &s
}

func computeRetryStrategy(n *core.Node, t *core.TaskTemplate) *v1alpha1.RetryStrategy {
	if n.GetMetadata() != nil && n.GetMetadata().GetRetries() != nil {
		return &v1alpha1.RetryStrategy{
			MinAttempts: refInt(int(n.GetMetadata().GetRetries().Retries + 1)),
		}
	}

	if t != nil && t.GetMetadata() != nil && t.GetMetadata().GetRetries() != nil {
		return &v1alpha1.RetryStrategy{
			MinAttempts: refInt(int(t.GetMetadata().GetRetries().Retries + 1)),
		}
	}

	return nil
}

func computeDeadline(n *core.Node) (*v1.Duration, error) {
	var deadline *v1.Duration
	if n.GetMetadata() != nil && n.GetMetadata().GetTimeout() != nil {
		duration, err := ptypes.Duration(n.GetMetadata().GetTimeout())
		if err != nil {
			return nil, err
		}
		deadline = &v1.Duration{
			Duration: duration,
		}
	}
	return deadline, nil
}

func toAliasValueArray(aliases []*core.Alias) []v1alpha1.Alias {
	if aliases == nil {
		return nil
	}

	res := make([]v1alpha1.Alias, 0, len(aliases))
	for _, alias := range aliases {
		res = append(res, v1alpha1.Alias{Alias: *alias})
	}

	return res
}

func toBindingValueArray(bindings []*core.Binding) []*v1alpha1.Binding {
	if bindings == nil {
		return nil
	}

	res := make([]*v1alpha1.Binding, 0, len(bindings))
	for _, binding := range bindings {
		res = append(res, &v1alpha1.Binding{Binding: binding})
	}

	return res
}

func minInt(i, j int) int {
	return int(math.Min(float64(i), float64(j)))
}

// StripTypeMetadata strips the type metadata from the given type.
func StripTypeMetadata(t *core.LiteralType) *core.LiteralType {
	if t == nil {
		return nil
	}

	c := *t
	c.Metadata = nil
	c.Annotation = nil
	// Note that we cannot strip `Structure` from the type because the dynamic node output type is used to validate the
	// interface of the dynamically compiled workflow. `Structure` is used to extend type checking information on
	// differnent Flyte types and is therefore required to ensure correct type validation.

	switch underlyingType := c.Type.(type) {
	case *core.LiteralType_UnionType:
		variants := make([]*core.LiteralType, 0, len(c.GetUnionType().Variants))
		for _, variant := range c.GetUnionType().Variants {
			variants = append(variants, StripTypeMetadata(variant))
		}

		underlyingType.UnionType.Variants = variants
	case *core.LiteralType_MapValueType:
		underlyingType.MapValueType = StripTypeMetadata(c.GetMapValueType())
	case *core.LiteralType_CollectionType:
		underlyingType.CollectionType = StripTypeMetadata(c.GetCollectionType())
	case *core.LiteralType_StructuredDatasetType:
		columns := make([]*core.StructuredDatasetType_DatasetColumn, 0, len(c.GetStructuredDatasetType().Columns))
		for _, column := range c.GetStructuredDatasetType().Columns {
			columns = append(columns, &core.StructuredDatasetType_DatasetColumn{
				Name:        column.Name,
				LiteralType: StripTypeMetadata(column.LiteralType),
			})
		}

		underlyingType.StructuredDatasetType.Columns = columns
	}

	return &c
}

func StripInterfaceTypeMetadata(iface *core.TypedInterface) *core.TypedInterface {
	if iface == nil {
		return nil
	}

	newIface := *iface

	if iface.Inputs != nil {
		for name, i := range iface.Inputs.Variables {
			i.Type = StripTypeMetadata(i.Type)
			i.Description = ""
			newIface.Inputs.Variables[name] = i
		}
	}

	if iface.Outputs != nil {
		for name, i := range iface.Outputs.Variables {
			i.Type = StripTypeMetadata(i.Type)
			i.Description = ""
			iface.Outputs.Variables[name] = i
		}
	}

	return &newIface
}
