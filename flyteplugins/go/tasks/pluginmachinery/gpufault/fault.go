package gpufault

// UnknownGPUIndex marks an Attribution whose GPU index could not be determined,
// which is not the same as index 0.
const UnknownGPUIndex = -1

// Kind separates GPU faults (Xid) from NVSwitch faults (SXid). The two share a
// numbering space but not a meaning, so a code alone is never enough to identify a
// fault.
type Kind string

const (
	KindXid  Kind = "xid"
	KindSXid Kind = "sxid"
)

// Severity is the operational reading of a fault code: whether the workload caused
// it, whether it is a warning worth surfacing, or whether the GPU is in trouble.
type Severity string

const (
	// SeverityUser marks faults a workload causes (illegal address, bad push buffer)
	// and that a healthy GPU recovers from once the offending process exits.
	SeverityUser Severity = "user"
	// SeverityWarn marks faults worth surfacing that do not by themselves condemn
	// the GPU.
	SeverityWarn Severity = "warn"
	// SeverityCritical marks faults after which the GPU is generally unusable until
	// it is reset or replaced.
	SeverityCritical Severity = "critical"
)

// Label is the upper-case form used in the human half of the event message.
func (s Severity) Label() string {
	switch s {
	case SeverityUser:
		return "USER"
	case SeverityCritical:
		return "CRITICAL"
	case SeverityWarn:
		return "WARN"
	default:
		return "WARN"
	}
}

// Fault is one GPU or NVSwitch fault as it travels through the event message.
//
// It deliberately carries only what the message carries. The emitter parses more out
// of the kernel line (the fault address, the channel id, the raw text, the time the
// line was read) and keeps it in its own log, because folding that variable text into
// the message would defeat the event recorder's correlator: repeats of the same fault
// on the same GPU render byte for byte the same message and are aggregated into a
// single Event with a count instead of flooding the API server.
type Fault struct {
	Kind     Kind
	Code     int
	Name     string
	Severity Severity
	// PCI is the normalized sysfs bus id of the GPU or NVSwitch, always with a
	// function suffix, for example 0000:3b:00.0.
	PCI string
	// PID is the host pid the driver blamed, or 0 when the line carried none.
	PID int
	// Process is the process name the driver blamed, empty when unknown.
	Process string
}

// IsSXid reports whether the fault came from an NVSwitch rather than a GPU.
func (f Fault) IsSXid() bool { return f.Kind == KindSXid }

// Attribution is where the fault happened, as far as the emitter could resolve it.
// The pods the emitter blamed are not part of it: the fault is reported as an event
// on the pod itself, so the reader already knows which pod it is looking at.
type Attribution struct {
	NodeName string
	// GPUUUID is the driver's UUID for the faulting GPU, empty when the PCI id
	// could not be resolved.
	GPUUUID string
	// GPUIndex is the host device index, UnknownGPUIndex when unresolved.
	GPUIndex int
}
