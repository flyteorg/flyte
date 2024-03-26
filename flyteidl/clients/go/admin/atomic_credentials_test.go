package admin

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAtomicPerRPCCredentials(t *testing.T) {
	a := atomicPerRPCCredentials{}
	assert.True(t, a.Load().RequireTransportSecurity())

	tokenSource := DummyTestTokenSource{}
	chTokenSource := NewCustomHeaderTokenSource(tokenSource, true, "my_custom_header")
	a.Store(chTokenSource)

	assert.False(t, a.Load().RequireTransportSecurity())
}

func TestNewPerRPCCredentialsFuture(t *testing.T) {
	f := NewPerRPCCredentialsFuture()
	assert.False(t, f.RequireTransportSecurity())
	assert.Equal(t, CustomHeaderTokenSource{}, f.Get())

	tokenSource := DummyTestTokenSource{}
	chTokenSource := NewCustomHeaderTokenSource(tokenSource, false, "my_custom_header")
	f.Store(chTokenSource)

	assert.True(t, f.Get().RequireTransportSecurity())
	assert.True(t, f.RequireTransportSecurity())
}

func ExampleNewPerRPCCredentialsFuture() {
	f := NewPerRPCCredentialsFuture()

	// Starts uninitialized
	fmt.Println("Initialized:", f.IsInitialized())

	// Implements credentials.PerRPCCredentials so can be used as one
	m, err := f.GetRequestMetadata(context.TODO(), "")
	fmt.Println("GetRequestMetadata:", m, "Error:", err)

	// Materialize the value later and populate
	tokenSource := DummyTestTokenSource{}
	f.Store(NewCustomHeaderTokenSource(tokenSource, false, "my_custom_header"))

	// Future calls to credentials.PerRPCCredentials methods will use the new instance
	m, err = f.GetRequestMetadata(context.TODO(), "")
	fmt.Println("GetRequestMetadata:", m, "Error:", err)

	// Output:
	// Initialized: false
	// GetRequestMetadata: map[] Error: <nil>
	// GetRequestMetadata: map[my_custom_header:Bearer abc] Error: <nil>
}
