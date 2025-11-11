package atomic

import "sync/atomic"

// This file contains some simplified atomics primitives that Golang default library does not offer
// like, Boolean

// Takes in a uint32 and converts to bool by checking whether the last bit is set to 1
func toBool(n uint32) bool {
	return n&1 == 1
}

// Takes in a bool and returns a uint32 representation
func toInt(b bool) uint32 {
	if b {
		return 1
	}
	return 0
}

// Bool is an atomic Boolean.
// It stores the bool as a uint32 internally. This is to use the uint32 atomic functions available in golang
type Bool struct{ v uint32 }

// NewBool creates a Bool.
func NewBool(initial bool) Bool {
	return Bool{v: toInt(initial)}
}

// Load atomically loads the Boolean.
func (b *Bool) Load() bool {
	return toBool(atomic.LoadUint32(&b.v))
}

// CAS is an atomic compare-and-swap.
func (b *Bool) CompareAndSwap(old, new bool) bool {
	return atomic.CompareAndSwapUint32(&b.v, toInt(old), toInt(new))
}

// Store atomically stores the passed value.
func (b *Bool) Store(new bool) {
	atomic.StoreUint32(&b.v, toInt(new))
}

// Swap sets the given value and returns the previous value.
func (b *Bool) Swap(new bool) bool {
	return toBool(atomic.SwapUint32(&b.v, toInt(new)))
}

// Toggle atomically negates the Boolean and returns the previous value.
func (b *Bool) Toggle() bool {
	return toBool(atomic.AddUint32(&b.v, 1) - 1)
}

type Uint32 struct {
	v uint32
}

// Returns a loaded uint32 value
func (u *Uint32) Load() uint32 {
	return atomic.LoadUint32(&u.v)
}

// CAS is an atomic compare-and-swap.
func (u *Uint32) CompareAndSwap(old, new uint32) bool {
	return atomic.CompareAndSwapUint32(&u.v, old, new)
}

// Add a delta to the number
func (u *Uint32) Add(delta uint32) uint32 {
	return atomic.AddUint32(&u.v, delta)
}

// Increment the value
func (u *Uint32) Inc() uint32 {
	return atomic.AddUint32(&u.v, 1)
}

// Set the value
func (u *Uint32) Store(v uint32) {
	atomic.StoreUint32(&u.v, v)
}

func NewUint32(v uint32) Uint32 {
	return Uint32{v: v}
}

type Int32 struct {
	v int32
}

// Returns a loaded uint32 value
func (i *Int32) Load() int32 {
	return atomic.LoadInt32(&i.v)
}

// CAS is an atomic compare-and-swap.
func (i *Int32) CompareAndSwap(old, new int32) bool {
	return atomic.CompareAndSwapInt32(&i.v, old, new)
}

// Add a delta to the number
func (i *Int32) Add(delta int32) int32 {
	return atomic.AddInt32(&i.v, delta)
}

// Subtract a delta from the number
func (i *Int32) Sub(delta int32) int32 {
	return atomic.AddInt32(&i.v, -delta)
}

// Increment the value
func (i *Int32) Inc() int32 {
	return atomic.AddInt32(&i.v, 1)
}

// Decrement the value
func (i *Int32) Dec() int32 {
	return atomic.AddInt32(&i.v, -1)
}

// Set the value
func (i *Int32) Store(v int32) {
	atomic.StoreInt32(&i.v, v)
}

func NewInt32(v int32) Int32 {
	return Int32{v: v}
}
