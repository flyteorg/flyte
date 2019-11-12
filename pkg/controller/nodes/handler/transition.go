package handler

type TransitionType int

const (
	TransitionTypeEphemeral TransitionType = iota
	TransitionTypeBarrier
)

type Transition struct {
	ttype TransitionType
	info  PhaseInfo
}

func (t Transition) Type() TransitionType {
	return t.ttype
}

func (t Transition) Info() PhaseInfo {
	return t.info
}

func (t Transition) WithInfo(p PhaseInfo) Transition {
	t.info = p
	return t
}

var UnknownTransition = Transition{TransitionTypeEphemeral, PhaseInfoUndefined}

func DoTransition(ttype TransitionType, info PhaseInfo) Transition {
	return Transition{ttype: ttype, info: info}
}
