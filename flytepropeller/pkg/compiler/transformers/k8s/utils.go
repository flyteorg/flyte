package k8s

import (
	"math"

	"github.com/golang/protobuf/ptypes"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
)

func refInt(i int) *int {
	return &i
}

func refStr(s string) *string {
	return &s
}

func computeRetryStrategy(n *core.Node, t *core.TaskTemplate) *v1alpha1.RetryStrategy {
	if n.GetMetadata() != nil && n.GetMetadata().GetRetries() != nil && n.GetMetadata().GetRetries().GetRetries() != 0 {
		return &v1alpha1.RetryStrategy{
			MinAttempts: refInt(int(n.GetMetadata().GetRetries().GetRetries() + 1)),
		}
	}

	if t != nil && t.GetMetadata() != nil && t.GetMetadata().GetRetries() != nil && t.GetMetadata().GetRetries().GetRetries() != 0 {
		return &v1alpha1.RetryStrategy{
			MinAttempts: refInt(int(t.GetMetadata().GetRetries().GetRetries() + 1)),
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

	// Special-case the presence of cache-key-metadata. This is a special field that is used to store metadata about
	// used in the cache key generation. This does not affect compatibility and it is purely used for cache key calculations.
	if c.GetAnnotation() != nil {
		annotations := c.GetAnnotation().GetAnnotations().GetFields()
		// The presence of the key `cache-key-metadata` indicates that we should leave the metadata intact.
		if _, ok := annotations["cache-key-metadata"]; !ok {
			c.Annotation = nil
		}
	}

	// Note that we cannot strip `Structure` from the type because the dynamic node output type is used to validate the
	// interface of the dynamically compiled workflow. `Structure` is used to extend type checking information on
	// different Flyte types and is therefore required to ensure correct type validation.

	switch underlyingType := c.GetType().(type) {
	case *core.LiteralType_UnionType:
		variants := make([]*core.LiteralType, 0, len(c.GetUnionType().GetVariants()))
		for _, variant := range c.GetUnionType().GetVariants() {
			variants = append(variants, StripTypeMetadata(variant))
		}

		underlyingType.UnionType.Variants = variants
	case *core.LiteralType_MapValueType:
		underlyingType.MapValueType = StripTypeMetadata(c.GetMapValueType())
	case *core.LiteralType_CollectionType:
		underlyingType.CollectionType = StripTypeMetadata(c.GetCollectionType())
	case *core.LiteralType_StructuredDatasetType:
		columns := make([]*core.StructuredDatasetType_DatasetColumn, 0, len(c.GetStructuredDatasetType().GetColumns()))
		for _, column := range c.GetStructuredDatasetType().GetColumns() {
			columns = append(columns, &core.StructuredDatasetType_DatasetColumn{
				Name:        column.GetName(),
				LiteralType: StripTypeMetadata(column.GetLiteralType()),
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

	if iface.GetInputs() != nil {
		for name, i := range iface.GetInputs().GetVariables() {
			i.Type = StripTypeMetadata(i.GetType())
			i.Description = ""
			newIface.Inputs.Variables[name] = i
		}
	}

	if iface.GetOutputs() != nil {
		for name, i := range iface.GetOutputs().GetVariables() {
			i.Type = StripTypeMetadata(i.GetType())
			i.Description = ""
			iface.Outputs.Variables[name] = i
		}
	}

	return &newIface
}
