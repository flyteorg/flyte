package v1alpha1

type ArrayNodeSpec struct {
	SubNodeSpec     *NodeSpec
	Parallelism     int64
	MinSuccesses    *uint32
	MinSuccessRatio *float32
}

func (a *ArrayNodeSpec) GetSubNodeSpec() *NodeSpec {
	return a.SubNodeSpec
}

func (a *ArrayNodeSpec) GetParallelism() int64 {
	return a.Parallelism
}

func (a *ArrayNodeSpec) GetMinSuccesses() *uint32 {
	return a.MinSuccesses
}

func (a *ArrayNodeSpec) GetMinSuccessRatio() *float32 {
	return a.MinSuccessRatio
}
