package v1alpha1

import (
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flytestdlib/utils"
)

type ConditionKind string

func (n ConditionKind) String() string {
	return string(n)
}

const (
	ConditionKindApprove ConditionKind = "approve"
	ConditionKindSignal  ConditionKind = "signal"
	ConditionKindSleep   ConditionKind = "sleep"
)

type ApproveCondition struct {
	*core.ApproveCondition
}

func (in ApproveCondition) MarshalJSON() ([]byte, error) {
	return utils.MarshalPbToBytes(in.ApproveCondition)
}

func (in *ApproveCondition) UnmarshalJSON(b []byte) error {
	in.ApproveCondition = &core.ApproveCondition{}
	return utils.UnmarshalBytesToPb(b, in.ApproveCondition)
}

type SignalCondition struct {
	*core.SignalCondition
}

func (in SignalCondition) MarshalJSON() ([]byte, error) {
	return utils.MarshalPbToBytes(in.SignalCondition)
}

func (in *SignalCondition) UnmarshalJSON(b []byte) error {
	in.SignalCondition = &core.SignalCondition{}
	return utils.UnmarshalBytesToPb(b, in.SignalCondition)
}

type SleepCondition struct {
	*core.SleepCondition
}

func (in SleepCondition) MarshalJSON() ([]byte, error) {
	return utils.MarshalPbToBytes(in.SleepCondition)
}

func (in *SleepCondition) UnmarshalJSON(b []byte) error {
	in.SleepCondition = &core.SleepCondition{}
	return utils.UnmarshalBytesToPb(b, in.SleepCondition)
}

type GateNodeSpec struct {
	Kind    ConditionKind     `json:"kind"`
	Approve *ApproveCondition `json:"approve,omitempty"`
	Signal  *SignalCondition  `json:"signal,omitempty"`
	Sleep   *SleepCondition   `json:"sleep,omitempty"`
}

func (g *GateNodeSpec) GetKind() ConditionKind {
	return g.Kind
}

func (g *GateNodeSpec) GetApprove() *core.ApproveCondition {
	return g.Approve.ApproveCondition
}

func (g *GateNodeSpec) GetSignal() *core.SignalCondition {
	return g.Signal.SignalCondition
}

func (g *GateNodeSpec) GetSleep() *core.SleepCondition {
	return g.Sleep.SleepCondition
}
