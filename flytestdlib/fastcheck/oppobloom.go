// Copyright 2012 Jeff Hodges. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Copyright FlyteAuthors.
// This code is liberally copied from the original implementation at https://github.com/jmhodges/opposite_of_a_bloom_filter/blob/master/go/oppobloom/oppobloom.go
// Package oppobloom implements a filter data structure that may report false negatives but no false positives.

// The fastcheck.OppoBloomFilter provides separate Contains, Add, and Remove methods.
// This makes it possible to check and then optionally add the value. It is possible
// that two threads may race and add it multiple times.
package fastcheck

import (
	"bytes"
	"context"
	"crypto/md5" //nolint:gosec
	"errors"
	"hash"
	"math"
	"sync/atomic"
	"unsafe"

	"go.opentelemetry.io/otel/metric"
)

var ErrSizeTooLarge = errors.New("oppobloom: size given too large to round to a power of 2")
var ErrSizeTooSmall = errors.New("oppobloom: filter cannot have a zero or negative size")
var MaxFilterSize = 1 << 30

// validate that it conforms to the interface
var _ Filter = &OppoBloomFilter{}

type md5UintHash struct {
	hash.Hash // a hack with knowledge of how md5 works
}

func (m md5UintHash) Sum32() uint32 {
	sum := m.Sum(nil)
	x := uint32(sum[0])
	for _, val := range sum[1:3] {
		x = x << 3
		x += uint32(val)
	}
	return x
}

// Implementation of the oppoFilter proposed in https://github.com/jmhodges/opposite_of_a_bloom_filter/ and
// the related blog https://www.somethingsimilar.com/2012/05/21/the-opposite-of-a-bloom-filter/
type OppoBloomFilter struct {
	array    []*[]byte
	sizeMask uint32
	metrics  Metrics
}

// getIndex calculates the hashindex of the given id
func (f *OppoBloomFilter) getIndex(id []byte) int32 {
	//nolint:gosec
	h := md5UintHash{md5.New()}
	h.Write(id)
	uindex := h.Sum32() & f.sizeMask
	return int32(uindex) // #nosec G115
}

func (f *OppoBloomFilter) Add(_ context.Context, id []byte) bool {
	oldID := getAndSet(f.array, f.getIndex(id), id)
	return !bytes.Equal(oldID, id)
}

func (f *OppoBloomFilter) Remove(_ context.Context, id []byte) bool {
	index := f.getIndex(id)
	indexPtr := (*unsafe.Pointer)(unsafe.Pointer(&f.array[index]))
	for {
		curr := get(f.array, index)
		if curr == nil || !bytes.Equal(id, *curr) {
			return false
		}
		// Only clear the value we checked; a concurrent Add may replace it.
		if atomic.CompareAndSwapPointer(indexPtr, unsafe.Pointer(curr), nil) {
			return true
		}
	}
}

func (f *OppoBloomFilter) Contains(ctx context.Context, id []byte) bool {
	curr := get(f.array, f.getIndex(id))
	if curr != nil {
		if bytes.Equal(id, *curr) {
			f.metrics.Hit.Add(ctx, 1, f.metrics.attributes)
			return true
		}
	}
	f.metrics.Miss.Add(ctx, 1, f.metrics.attributes)
	return false
}

// Helper methods

// get returns the actual value stored at the given index. If not found the value can be nil
func get(arr []*[]byte, index int32) *[]byte {
	indexPtr := (*unsafe.Pointer)(unsafe.Pointer(&arr[index]))
	return (*[]byte)(atomic.LoadPointer(indexPtr))
}

// getAndSet Returns the id that was in the slice at the given index after putting the
// new id in the slice at that index, atomically.
func getAndSet(arr []*[]byte, index int32, id []byte) []byte {
	indexPtr := (*unsafe.Pointer)(unsafe.Pointer(&arr[index]))
	idUnsafe := unsafe.Pointer(&id)
	var oldID []byte
	for {
		oldIDUnsafe := atomic.LoadPointer(indexPtr)
		if atomic.CompareAndSwapPointer(indexPtr, oldIDUnsafe, idUnsafe) {
			oldIDPtr := (*[]byte)(oldIDUnsafe)
			if oldIDPtr != nil {
				oldID = *oldIDPtr
			}
			break
		}
	}
	return oldID
}

// NewOppoBloomFilter creates a new Opposite of Bloom filter proposed in https://github.com/jmhodges/opposite_of_a_bloom_filter/ and
// the related blog https://www.somethingsimilar.com/2012/05/21/the-opposite-of-a-bloom-filter/.
// Scope identifies the filter in OTEL metrics. A nil provider disables metrics.
func NewOppoBloomFilter(size int, scope string, provider metric.MeterProvider) (*OppoBloomFilter, error) {
	if size > MaxFilterSize {
		return nil, ErrSizeTooLarge
	}
	if size <= 0 {
		return nil, ErrSizeTooSmall
	}
	// round to the next largest power of two
	size = int(math.Pow(2, math.Ceil(math.Log2(float64(size)))))
	slice := make([]*[]byte, size)
	sizeMask := uint32(size - 1) // #nosec G115
	f := &OppoBloomFilter{array: slice, sizeMask: sizeMask}
	metrics, err := newMetrics(scope, "oppobloom", f.size, provider)
	if err != nil {
		return nil, err
	}
	f.metrics = metrics
	return f, nil
}

// size counts occupied slots at collection time without slowing down mutations.
// Atomic reads keep collection safe during concurrent Add and Remove operations;
// like Contains, the observation can change while operations are in flight.
func (f *OppoBloomFilter) size() int64 {
	var count int64
	for i := range f.array {
		if get(f.array, int32(i)) != nil { // #nosec G115 -- capacity is bounded by MaxFilterSize.
			count++
		}
	}
	return count
}
