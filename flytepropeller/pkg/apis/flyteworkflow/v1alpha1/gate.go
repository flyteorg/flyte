package v1alpha1

import (
	"bytes"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/golang/protobuf/jsonpb"
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
	if in.ApproveCondition == nil {
		return nilJSON, nil
	}

	var buf bytes.Buffer
	if err := marshaler.Marshal(&buf, in.ApproveCondition); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (in *ApproveCondition) UnmarshalJSON(b []byte) error {
	in.ApproveCondition = &core.ApproveCondition{}
	return jsonpb.Unmarshal(bytes.NewReader(b), in.ApproveCondition)
}

type SignalCondition struct {
	*core.SignalCondition
}

func (in SignalCondition) MarshalJSON() ([]byte, error) {
	if in.SignalCondition == nil {
		return nilJSON, nil
	}

	var buf bytes.Buffer
	if err := marshaler.Marshal(&buf, in.SignalCondition); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (in *SignalCondition) UnmarshalJSON(b []byte) error {
	in.SignalCondition = &core.SignalCondition{}
	return jsonpb.Unmarshal(bytes.NewReader(b), in.SignalCondition)
}

type SleepCondition struct {
	*core.SleepCondition
}

func (in SleepCondition) MarshalJSON() ([]byte, error) {
	if in.SleepCondition == nil {
		return nilJSON, nil
	}

	var buf bytes.Buffer
	if err := marshaler.Marshal(&buf, in.SleepCondition); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (in *SleepCondition) UnmarshalJSON(b []byte) error {
	in.SleepCondition = &core.SleepCondition{}
	return jsonpb.Unmarshal(bytes.NewReader(b), in.SleepCondition)
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
