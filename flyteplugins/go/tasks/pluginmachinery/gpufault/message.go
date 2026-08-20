package gpufault

import (
	"fmt"
	"strconv"
	"strings"
)

// MessagePrefix marks every message the GPU health emitter writes. Consumers filter
// on it, so it must never change.
// Kubernetes Event reasons the GPU fault emitter uses. Consumers filter on these
// before parsing the message, so ordinary events never reach the parser and a
// free-text message alone cannot pose as a fault report.
const (
	EventReasonXid  = "GPUXidError"
	EventReasonSXid = "GPUSXidError"
)

const MessagePrefix = "[gpu-health]"

// The event message is two halves. The first is a sentence a human reads in the run's
// Logs tab without knowing anything about Xid codes. The second is a k=v tail a
// program reads back with ParseEventMessage, which is how the typed core.GpuFault is
// produced without re-parsing kernel text.
const (
	keyXid      = "xid"
	keySXid     = "sxid"
	keySeverity = "severity"
	keyGPUUUID  = "gpu_uuid"
	keyGPUIndex = "gpu_index"
	keyPCI      = "pci"
	keyNode     = "node"
	keyPID      = "pid"
	keyProcess  = "process"
)

// FormatEventMessage renders the message body of the Kubernetes Event for a fault.
func FormatEventMessage(f Fault, a Attribution) string {
	return Sentence(f, a) + machineTail(f, a)
}

// Sentence is the human half of the event message: everything up to and including the
// full stop, with no k=v tail. Failure classification prepends it to the failure
// message so the user reads what the driver reported instead of a bare exit code.
func Sentence(f Fault, a Attribution) string {
	var sb strings.Builder

	sb.WriteString(MessagePrefix)
	sb.WriteString(" [")
	sb.WriteString(f.Severity.Label())
	sb.WriteString("] ")

	if f.IsSXid() {
		fmt.Fprintf(&sb, "SXid %d on NVSwitch %s.", f.Code, f.PCI)
	} else {
		fmt.Fprintf(&sb, "Xid %d (%s)%s.", f.Code, f.Name, gpuPhrase(f, a))
	}

	return sb.String()
}

// machineTail is the k=v half, leading space included so it appends straight onto the
// sentence.
func machineTail(f Fault, a Attribution) string {
	var sb strings.Builder

	codeKey := keyXid
	if f.IsSXid() {
		codeKey = keySXid
	}

	fmt.Fprintf(&sb, " %s=%d", codeKey, f.Code)
	fmt.Fprintf(&sb, " %s=%s", keySeverity, string(f.Severity))
	if a.GPUUUID != "" {
		fmt.Fprintf(&sb, " %s=%s", keyGPUUUID, a.GPUUUID)
	}
	if a.GPUIndex >= 0 {
		fmt.Fprintf(&sb, " %s=%d", keyGPUIndex, a.GPUIndex)
	}
	if f.PCI != "" {
		fmt.Fprintf(&sb, " %s=%s", keyPCI, f.PCI)
	}
	if a.NodeName != "" {
		fmt.Fprintf(&sb, " %s=%s", keyNode, a.NodeName)
	}
	if f.PID > 0 {
		fmt.Fprintf(&sb, " %s=%d", keyPID, f.PID)
	}
	if f.Process != "" {
		fmt.Fprintf(&sb, " %s=%s", keyProcess, sanitizeValue(f.Process))
	}

	return sb.String()
}

// gpuPhrase names the GPU as precisely as attribution allowed, degrading to the bus
// id when the driver's procfs entry could not be read.
func gpuPhrase(f Fault, a Attribution) string {
	switch {
	case a.GPUUUID != "" && a.GPUIndex >= 0:
		return fmt.Sprintf(" on GPU %d %s", a.GPUIndex, a.GPUUUID)
	case a.GPUUUID != "":
		return fmt.Sprintf(" on GPU %s", a.GPUUUID)
	case a.GPUIndex >= 0:
		return fmt.Sprintf(" on GPU %d", a.GPUIndex)
	case f.PCI != "":
		return fmt.Sprintf(" on GPU at PCI %s", f.PCI)
	default:
		return ""
	}
}

// ParseEventMessage reads back a message produced by FormatEventMessage.
//
// It recovers everything the message carries: kind, code, name, severity, bus id, pid
// and process, plus the node, GPU UUID and GPU index. It reports false for any message
// that is not one of ours, which is how consumers tell a GPU fault event apart from
// every other event recorded on the same pod.
func ParseEventMessage(msg string) (Fault, Attribution, bool) {
	if !strings.HasPrefix(msg, MessagePrefix) {
		return Fault{}, Attribution{}, false
	}

	kind := KindXid
	start := strings.Index(msg, " "+keyXid+"=")
	if start < 0 {
		if start = strings.Index(msg, " "+keySXid+"="); start < 0 {
			return Fault{}, Attribution{}, false
		}
		kind = KindSXid
	}

	fields := map[string]string{}
	for _, token := range strings.Fields(msg[start:]) {
		key, value, ok := strings.Cut(token, "=")
		if !ok {
			continue
		}
		fields[key] = value
	}

	codeKey := keyXid
	if kind == KindSXid {
		codeKey = keySXid
	}
	code, err := strconv.Atoi(fields[codeKey])
	if err != nil {
		return Fault{}, Attribution{}, false
	}

	fault := Fault{
		Kind:     kind,
		Code:     code,
		Name:     NameFor(kind, code),
		Severity: SeverityFor(kind, code),
		PCI:      fields[keyPCI],
		Process:  fields[keyProcess],
	}
	if sev, ok := fields[keySeverity]; ok {
		fault.Severity = ParseSeverity(sev)
	}
	if pid, err := strconv.Atoi(fields[keyPID]); err == nil {
		fault.PID = pid
	}

	attribution := Attribution{
		NodeName: fields[keyNode],
		GPUUUID:  fields[keyGPUUUID],
		GPUIndex: UnknownGPUIndex,
	}
	if index, err := strconv.Atoi(fields[keyGPUIndex]); err == nil {
		attribution.GPUIndex = index
	}

	return fault, attribution, true
}

// sanitizeValue keeps a k=v token parseable. Process names come from the kernel's
// comm field, which is almost always a single word but is not guaranteed to be.
func sanitizeValue(v string) string {
	return strings.Map(func(r rune) rune {
		switch r {
		case ' ', '\t', '\n', '\r', '=':
			return '_'
		default:
			return r
		}
	}, v)
}
