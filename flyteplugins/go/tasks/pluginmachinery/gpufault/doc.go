// Package gpufault is the one definition of how a GPU fault travels from the node it
// happened on to the user who has to act on it.
//
// The contract is a Kubernetes Event message. A GPU health daemon on every GPU node
// reads the kernel log, recognizes NVIDIA Xid lines (a fault on a GPU) and NVSwitch
// SXid lines (a fault on the fabric), works out which pod held the device, and records
// a Warning Event against that pod with reason GPUXidError or GPUSXidError. The
// message it writes is FormatEventMessage: a sentence a human can read followed by a
// k=v tail a program reads back with ParseEventMessage. Events already recorded on
// running clusters have to keep parsing, so the format is fixed and every producer and
// consumer renders it through this package instead of writing its own.
//
// The consumer in this repository is the Kubernetes plugin manager in the executor.
// The event watcher already forwards every event recorded on a task's pod into the
// attempt's cluster events; when an attempt ends in failure the plugin manager reads
// the events back, turns each GPU fault message into a core.GpuFault with
// FromEventMessage, and hands the list to ClassifyFailure along with the failure the
// plugin reported. The fault ends up on ExecutionError.gpu_fault, which reaches the
// console and the SDK as typed data, so nobody downstream parses the message text.
//
// ClassifyFailure only ever looks at a failed attempt, and only when at least one
// fault was recorded. What it does depends on the worst severity it finds.
//
// A critical fault, such as Xid 79 (the GPU fell off the bus) or any SXid, turns the
// failure into a system retryable failure with the code CodeFor gives for that fault.
// The device or the node is not trustworthy at that point and the workload did not
// cause it, so the failure must not consume one of the user's retries, and the retry
// wants to land on different hardware.
//
// A user fault, such as Xid 31 (a GPU memory page fault from an out-of-bounds access),
// leaves the plugin's verdict alone: same phase, same error kind, same retry budget.
// It only names the failure, replacing a generic code such as UnknownError or a bare
// exit status with CodeGpuXidError and putting the driver's sentence in front of the
// message.
//
// A warning-only fault changes nothing at all beyond attaching the fault, so that the
// console can show what the GPU reported while the task was running.
package gpufault
