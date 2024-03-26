package v1alpha1

type ArrayNodeSpec struct {
	SubNodeSpec     *NodeSpec
	Parallelism     uint32
	MinSuccesses    *uint32
	MinSuccessRatio *float32
}

func (a *ArrayNodeSpec) GetSubNodeSpec() *NodeSpec {
	return a.SubNodeSpec
}

func (a *ArrayNodeSpec) GetParallelism() uint32 {
	return a.Parallelism
}

func (a *ArrayNodeSpec) GetMinSuccesses() *uint32 {
	return a.MinSuccesses
}

func (a *ArrayNodeSpec) GetMinSuccessRatio() *float32 {
	return a.MinSuccessRatio
}
