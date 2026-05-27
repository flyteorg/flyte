/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0
*/

package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestDefaultMaxConcurrentReconciles locks the default value introduced in
// #7205. The default is 1 — one reconcile worker — to match
// controller-runtime's own default and preserve the historical
// single-worker behavior of the executor controller. Operators tune this
// upward to spend more CPU on parallel reconciles when the TaskAction
// queue grows; bumping the default would change the worker pool sizing of
// every existing deployment, so this test is deliberately a tripwire.
func TestDefaultMaxConcurrentReconciles(t *testing.T) {
	assert.Equal(t, uint32(1), defaultConfig.MaxConcurrentReconciles,
		"changing this default changes worker pool sizing for every existing deployment; "+
			"if intentional, update both this test and the package doc-comment together")
}
