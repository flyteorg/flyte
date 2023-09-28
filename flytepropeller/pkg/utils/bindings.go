package utils

import (
	"github.com/flyteorg/flyteidl/clients/go/coreutils"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
)

func MakeBindingDataPromise(fromNode, fromVar string) *core.BindingData {
	return &core.BindingData{
		Value: &core.BindingData_Promise{
			Promise: &core.OutputReference{
				Var:    fromVar,
				NodeId: fromNode,
			},
		},
	}
}

func MakeBindingPromise(fromNode, fromVar, toVar string) *core.Binding {
	return &core.Binding{
		Var:     toVar,
		Binding: MakeBindingDataPromise(fromNode, fromVar),
	}
}

func MakeBindingDataCollection(bindings ...*core.BindingData) *core.BindingData {
	return &core.BindingData{
		Value: &core.BindingData_Collection{
			Collection: &core.BindingDataCollection{
				Bindings: bindings,
			},
		},
	}
}

type Pair struct {
	K string
	V *core.BindingData
}

func NewPair(k string, v *core.BindingData) Pair {
	return Pair{K: k, V: v}
}

func MakeBindingDataMap(pairs ...Pair) *core.BindingData {
	bindingsMap := map[string]*core.BindingData{}
	for _, p := range pairs {
		bindingsMap[p.K] = p.V
	}
	return &core.BindingData{
		Value: &core.BindingData_Map{
			Map: &core.BindingDataMap{
				Bindings: bindingsMap,
			},
		},
	}
}

func MakePrimitiveBindingData(v interface{}) (*core.BindingData, error) {
	p, err := coreutils.MakePrimitive(v)
	if err != nil {
		return nil, err
	}
	return &core.BindingData{
		Value: &core.BindingData_Scalar{
			Scalar: &core.Scalar{
				Value: &core.Scalar_Primitive{
					Primitive: p,
				},
			},
		},
	}, nil
}

func MustMakePrimitiveBindingData(v interface{}) *core.BindingData {
	p, err := MakePrimitiveBindingData(v)
	if err != nil {
		panic(err)
	}
	return p
}

func MakeBinding(variable string, b *core.BindingData) *core.Binding {
	return &core.Binding{
		Var:     variable,
		Binding: b,
	}
}
