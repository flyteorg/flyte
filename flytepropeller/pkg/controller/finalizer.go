package controller

import v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

const FinalizerKey = "flyte-finalizer"

// NOTE: Some of these APIs are exclusive and do not compare the actual values of the finalizers.
// the intention of this module is to set only one opaque finalizer at a time. If you want to set multiple (not common)
// finalizers, use this module carefully and at your own risk!

// Sets a new finalizer in case the finalizer is empty
func SetFinalizerIfEmpty(meta v1.Object, finalizer string) {
	if !HasFinalizer(meta) {
		meta.SetFinalizers([]string{finalizer})
	}
}

// Check if the deletion timestamp is set, this is set automatically when an object is deleted
func IsDeleted(meta v1.Object) bool {
	return meta.GetDeletionTimestamp() != nil
}

// Reset all the finalizers on the object
func ResetFinalizers(meta v1.Object) {
	meta.SetFinalizers([]string{})
}

// Currently we only compare the lengths of finalizers. If you add finalizers directly these API;'s will not work
func FinalizersIdentical(o1 v1.Object, o2 v1.Object) bool {
	return len(o1.GetFinalizers()) == len(o2.GetFinalizers())
}

// Check if any finalizer is set
func HasFinalizer(meta v1.Object) bool {
	return len(meta.GetFinalizers()) != 0
}
