package gpufault

import (
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
)

// ToProto renders a fault and where it happened as the IDL message that travels on
// ClusterEvent and ExecutionError.
func ToProto(f Fault, a Attribution) *core.GpuFault {
	out := &core.GpuFault{
		Kind:     kindToProto(f.Kind),
		Code:     uint32(max(f.Code, 0)),
		Name:     f.Name,
		Severity: severityToProto(f.Severity),
		GpuUuid:  a.GPUUUID,
		PciBusId: f.PCI,
		Node:     a.NodeName,
		Pid:      uint32(max(f.PID, 0)),
		Process:  f.Process,
	}
	if a.GPUIndex >= 0 {
		index := uint32(a.GPUIndex)
		out.GpuIndex = &index
	}
	return out
}

// FromProto is the inverse of ToProto. An unset severity or kind is resolved from the
// code, so a producer that only filled in the numbers still yields a usable fault.
func FromProto(p *core.GpuFault) (Fault, Attribution) {
	if p == nil {
		return Fault{}, Attribution{GPUIndex: UnknownGPUIndex}
	}

	kind := kindFromProto(p.GetKind())
	code := int(p.GetCode())

	fault := Fault{
		Kind:     kind,
		Code:     code,
		Name:     p.GetName(),
		Severity: severityFromProto(p.GetSeverity()),
		PCI:      p.GetPciBusId(),
		PID:      int(p.GetPid()),
		Process:  p.GetProcess(),
	}
	if fault.Name == "" {
		fault.Name = NameFor(kind, code)
	}
	if fault.Severity == "" {
		fault.Severity = SeverityFor(kind, code)
	}

	attribution := Attribution{
		NodeName: p.GetNode(),
		GPUUUID:  p.GetGpuUuid(),
		GPUIndex: UnknownGPUIndex,
	}
	if p.GpuIndex != nil {
		attribution.GPUIndex = int(p.GetGpuIndex())
	}

	return fault, attribution
}

// FromEventMessage turns the message body of a Kubernetes Event into a typed fault,
// returning nil for every event that is not one of the emitter's.
func FromEventMessage(msg string) *core.GpuFault {
	fault, attribution, ok := ParseEventMessage(msg)
	if !ok {
		return nil
	}
	// The severity in the message tail is display data from whoever wrote the
	// event. Classification decides retry budgets and, later, quarantine, so it
	// re-derives severity from this package's own table: a message cannot talk a
	// consumer into treating an unknown or user-class code as critical.
	fault.Severity = SeverityFor(fault.Kind, fault.Code)
	return ToProto(fault, attribution)
}

func kindToProto(k Kind) core.GpuFault_Kind {
	if k == KindSXid {
		return core.GpuFault_KIND_SXID
	}
	return core.GpuFault_KIND_XID
}

func kindFromProto(k core.GpuFault_Kind) Kind {
	if k == core.GpuFault_KIND_SXID {
		return KindSXid
	}
	return KindXid
}

func severityToProto(s Severity) core.GpuFault_Severity {
	switch s {
	case SeverityUser:
		return core.GpuFault_SEVERITY_USER
	case SeverityCritical:
		return core.GpuFault_SEVERITY_CRITICAL
	case SeverityWarn:
		return core.GpuFault_SEVERITY_WARN
	default:
		return core.GpuFault_SEVERITY_UNSPECIFIED
	}
}

// severityFromProto returns the empty Severity for an unset value so that callers can
// fall back to the code table instead of silently reading it as a warning.
func severityFromProto(s core.GpuFault_Severity) Severity {
	switch s {
	case core.GpuFault_SEVERITY_USER:
		return SeverityUser
	case core.GpuFault_SEVERITY_WARN:
		return SeverityWarn
	case core.GpuFault_SEVERITY_CRITICAL:
		return SeverityCritical
	default:
		return ""
	}
}
