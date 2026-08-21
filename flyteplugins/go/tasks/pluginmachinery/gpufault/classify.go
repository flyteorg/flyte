package gpufault

import (
	"regexp"

	"google.golang.org/protobuf/proto"

	pluginsCore "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
)

// Failure codes a GPU fault can put on an ExecutionError. They are deliberately
// coarse: a code names the class of trouble a user or an operator can act on, and the
// exact Xid number stays available on the typed fault.
const (
	// CodeGpuXidError is the catch-all for a fault with no more specific code.
	CodeGpuXidError = "GpuXidError"
	// CodeGpuFallenOffBus is Xid 79: the driver lost the device on the PCIe bus.
	CodeGpuFallenOffBus = "GpuFallenOffBus"
	// CodeGpuEccUncorrectable covers the ECC errors memory could not correct.
	CodeGpuEccUncorrectable = "GpuEccUncorrectable"
	// CodeGpuRowRemapPending covers row remapping recorded or failed to record, which
	// needs a GPU reset before the memory is usable again.
	CodeGpuRowRemapPending = "GpuRowRemapPending"
	// CodeGpuNvlinkError covers NVLink faults and every NVSwitch SXid, which take
	// down fabric links shared across the node.
	CodeGpuNvlinkError = "GpuNvlinkError"
	// CodeGpuGspError covers the GPU System Processor timing out or erroring.
	CodeGpuGspError = "GpuGspError"
)

// CodeFor maps a fault onto the failure code the user sees.
func CodeFor(f Fault) string {
	if f.IsSXid() {
		return CodeGpuNvlinkError
	}
	switch f.Code {
	case 79:
		return CodeGpuFallenOffBus
	case 48, 94, 95, 140:
		return CodeGpuEccUncorrectable
	case 63, 64:
		return CodeGpuRowRemapPending
	case 74:
		return CodeGpuNvlinkError
	case 119, 120:
		return CodeGpuGspError
	default:
		return CodeGpuXidError
	}
}

// genericCodes are the codes that say nothing about why the task failed. When a GPU
// fault is on record, naming the fault is strictly more useful than keeping one of
// these.
var genericCodes = map[string]bool{
	"":             true,
	"Unknown":      true,
	"UnknownError": true,
	"Error":        true,
	// DemystifyFailure's verdict for a pod that was killed or vanished without the
	// kubelet recording why. That is exactly the shape a hardware GPU fault leaves
	// behind, so a recorded fault is the better name for it.
	"Interrupted": true,
}

// exitCodeStyle matches a code that is only the container's exit status, for example
// "137" or "ExitCode1". Those are exit statuses reported as codes, so they carry no
// more information than a generic code does.
var exitCodeStyle = regexp.MustCompile(`^(?i:exit[ _-]?code[ _-]?)?[0-9]+$`)

func isGenericCode(code string) bool {
	return genericCodes[code] || exitCodeStyle.MatchString(code)
}

// ClassifyFailure folds the GPU faults observed on an attempt's pod into the failure
// the plugin reported. faults are the faults recorded against that pod over the whole
// attempt, in the order they happened; a successful, running or aborted phase and an
// empty fault list are both returned unchanged.
//
// The result always carries the fault as data on ExecutionError.gpu_fault, so a
// consumer never has to read it out of the message text.
func ClassifyFailure(phase pluginsCore.PhaseInfo, faults []*core.GpuFault) pluginsCore.PhaseInfo {
	if len(faults) == 0 || !phase.Phase().IsFailure() {
		return phase
	}

	if fault, f, a := firstOfSeverity(faults, SeverityCritical); fault != nil {
		// A critical Xid means the device or the node is no longer trustworthy: the
		// workload did not cause it and rerunning in place would most likely hit the
		// same hardware. A retryable failure therefore stops charging the user's
		// budget and becomes a system retry; once phase 3 quarantines the node, the
		// reschedule also lands somewhere else. A permanent failure stays permanent:
		// the fault does not make an unrunnable task runnable, it only reclassifies
		// whose problem the failure is.
		err := cloneExecutionError(phase.Err())
		err.Kind = core.ExecutionError_SYSTEM
		// A specific code the plugin worked out, such as OOMKilled, names something
		// the fault does not explain away; only meaningless codes give way to the
		// fault's own.
		if isGenericCode(err.GetCode()) {
			err.Code = CodeFor(f)
		}
		err.Message = prependSentence(f, a, err.GetMessage())
		err.GpuFault = fault
		return keepVerdict(phase, err)
	}

	if fault, f, a := firstOfSeverity(faults, SeverityUser); fault != nil {
		// A user Xid is the workload's own doing, for example an out-of-bounds access
		// (Xid 31). The verdict the plugin reached stands; all this adds is a name for
		// what went wrong, so that the user reads "GPU memory page fault" instead of a
		// bare exit code.
		err := cloneExecutionError(phase.Err())
		if isGenericCode(err.GetCode()) {
			err.Code = CodeGpuXidError
		}
		err.Message = prependSentence(f, a, err.GetMessage())
		err.GpuFault = fault
		return keepVerdict(phase, err)
	}

	// Only warnings: nothing about the failure changes, the fault rides along so the
	// console can show what the GPU reported while the task was running.
	err := cloneExecutionError(phase.Err())
	err.GpuFault = faults[0]
	return keepVerdict(phase, err)
}

// firstOfSeverity returns the first fault of the given severity in time order, along
// with its decoded form so the caller does not decode it twice.
func firstOfSeverity(faults []*core.GpuFault, severity Severity) (*core.GpuFault, Fault, Attribution) {
	for _, fault := range faults {
		f, a := FromProto(fault)
		if f.Severity == severity {
			return fault, f, a
		}
	}
	return nil, Fault{}, Attribution{}
}

// keepVerdict rebuilds the failure with the phase the plugin chose, changing only
// what the fault added to the error, and keeps the cleanup flag: a pod that had to
// be cleaned up before classification still has to be cleaned up after it.
func keepVerdict(phase pluginsCore.PhaseInfo, err *core.ExecutionError) pluginsCore.PhaseInfo {
	out := pluginsCore.PhaseInfoFailed(phase.Phase(), err, phase.Info())
	if phase.CleanupOnFailure() {
		out = out.WithCleanupOnFailure()
	}
	return preserveShape(phase, out)
}

// preserveShape carries over the parts of a PhaseInfo the failure constructors do not
// take: the phase version and the reason accumulated so far.
func preserveShape(phase pluginsCore.PhaseInfo, out pluginsCore.PhaseInfo) pluginsCore.PhaseInfo {
	out = out.WithVersion(phase.Version())
	if reason := phase.Reason(); reason != "" {
		out.WithReason(reason)
	}
	return out
}

func cloneExecutionError(err *core.ExecutionError) *core.ExecutionError {
	if err == nil {
		return &core.ExecutionError{}
	}
	return proto.Clone(err).(*core.ExecutionError)
}

// prependSentence puts the driver's own account of the fault in front of whatever the
// plugin had to say about the failure.
func prependSentence(f Fault, a Attribution, message string) string {
	sentence := Sentence(f, a)
	if message == "" {
		return sentence
	}
	return sentence + " " + message
}
