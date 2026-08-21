package gpufault

import "fmt"

// xidNames is the human name NVIDIA documents for each Xid code. It does not need to
// be exhaustive: unknown codes fall back to "Xid <n>" so a new driver release never
// makes a fault get dropped on the floor.
var xidNames = map[int]string{
	13:  "Graphics Engine Exception",
	31:  "GPU memory page fault",
	32:  "Invalid or corrupted push buffer stream",
	38:  "Driver firmware error",
	43:  "GPU stopped processing",
	45:  "Preemptive cleanup, due to previous errors",
	48:  "Double Bit ECC Error",
	61:  "Internal micro-controller breakpoint/warning",
	62:  "Internal micro-controller halt",
	63:  "ECC page retirement or row remapping recording event",
	64:  "ECC page retirement or row remapper recording failure",
	68:  "NVDEC0 Exception",
	69:  "Graphics Engine class error",
	74:  "NVLink Error",
	79:  "GPU has fallen off the bus",
	92:  "High single-bit ECC error rate",
	94:  "Contained ECC error",
	95:  "Uncontained ECC error",
	109: "Context Switch Timeout Error",
	119: "GSP RPC Timeout",
	120: "GSP Error",
	140: "Unrecovered ECC Error",
	154: "GPU recovery action changed",
}

// userXids are the codes a workload causes rather than the hardware. They still
// interrupt the run, but the GPU itself is fine once the process is gone.
var userXids = map[int]bool{
	13: true,
	31: true,
	43: true,
	45: true,
}

// criticalXids are the codes after which the GPU is generally unusable until it is
// reset or replaced.
var criticalXids = map[int]bool{
	48:  true,
	62:  true,
	63:  true,
	64:  true,
	74:  true,
	79:  true,
	94:  true,
	95:  true,
	109: true,
	119: true,
	120: true,
	// 140 must stay in step with CodeFor, which maps it to CodeGpuEccUncorrectable:
	// an unrecovered ECC error is a device fault, not a warning.
	140: true,
}

// warnXids are codes that are neither the workload's fault nor immediately fatal.
// Anything not listed anywhere lands here too.
var warnXids = map[int]bool{
	92: true,
}

// NameFor returns the documented name for a fault code, or a generic name when the
// code is not in the table.
func NameFor(kind Kind, code int) string {
	if kind == KindSXid {
		return fmt.Sprintf("SXid %d", code)
	}
	if name, ok := xidNames[code]; ok {
		return name
	}
	return fmt.Sprintf("Xid %d", code)
}

// SeverityFor classifies a fault code. Every SXid is critical: NVSwitch faults take
// down fabric links shared by every GPU on the node, so there is no benign case.
func SeverityFor(kind Kind, code int) Severity {
	if kind == KindSXid {
		return SeverityCritical
	}
	switch {
	case userXids[code]:
		return SeverityUser
	case criticalXids[code]:
		return SeverityCritical
	case warnXids[code]:
		return SeverityWarn
	default:
		return SeverityWarn
	}
}

// ParseSeverity turns the wire form back into a Severity, defaulting to warn for
// anything unrecognized so a future severity value never fails a parse.
func ParseSeverity(s string) Severity {
	switch Severity(s) {
	case SeverityUser:
		return SeverityUser
	case SeverityCritical:
		return SeverityCritical
	case SeverityWarn:
		return SeverityWarn
	default:
		// Unknown labels are reported as such rather than guessed at, so that a
		// consumer falls back to its own table instead of silently downgrading.
		return ""
	}
}
