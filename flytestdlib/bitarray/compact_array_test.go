package bitarray

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func getRandInt() uint64 {
	c := 10
	b := make([]byte, c)
	_, err := rand.Read(b)
	if err != nil {
		return 0
	}

	return binary.BigEndian.Uint64(b)
}

func TestNewItemArray(t *testing.T) {
	for itemSize := uint(1); itemSize < 60; itemSize++ {
		t.Run(fmt.Sprintf("Item Size %v", itemSize), func(t *testing.T) {
			maxItemValue := 1 << (itemSize - 1)
			itemsCount := uint(math.Min(float64(200), float64(maxItemValue)))
			expected := make([]Item, 0, itemsCount)

			arr, err := NewCompactArray(itemsCount, Item(1)<<(itemSize-1))
			assert.NoError(t, err)

			for i := 0; i < int(itemsCount); i++ {
				// Ensure inserted items is in the accepted range (0 -> 1<<itemSize)
				val := getRandInt() % (1 << (itemSize - 1))
				arr.SetItem(i, val/3)
				arr.SetItem(i, val)
				expected = append(expected, val)
			}

			items := arr.GetItems()
			//t.Log(items)
			assert.Equal(t, expected, items)
		})
	}
}

func TestMarshal(t *testing.T) {
	itemArray, err := NewCompactArray(105, 7)
	assert.NoError(t, err)

	for idx := range itemArray.GetItems() {
		itemArray.SetItem(idx, Item(4))
	}

	raw, err := json.Marshal(itemArray)
	assert.NoError(t, err)

	m := map[string]interface{}{}
	assert.NoError(t, json.Unmarshal(raw, &m))

	raw, err = json.Marshal(m)
	assert.NoError(t, err)

	actualArray, err := NewCompactArray(105, 7)
	assert.NoError(t, err)
	assert.NoError(t, json.Unmarshal(raw, &actualArray))

	for idx, item := range actualArray.GetItems() {
		assert.Equal(t, Item(4), item, "Expecting 4 at idx %v", idx)
	}
}

func TestCompactArray_GetItems(t *testing.T) {
	itemArray, err := NewCompactArray(2, 15)
	assert.NoError(t, err)
	itemArray.SetItem(0, 3)
	itemArray.SetItem(1, 3)
	itemArray.SetItem(0, 8)

	assert.Equal(t, uint64(8), itemArray.GetItem(0))
	assert.Equal(t, uint64(3), itemArray.GetItem(1))
}

func BenchmarkCompactArray_SetItem(b *testing.B) {
	// Number of bits = log b.N (to base 2)
	itemsCount := uint(b.N)
	expected := make([]Item, 0, itemsCount)

	b.ResetTimer()
	arr, err := NewCompactArray(itemsCount, Item(b.N))
	assert.NoError(b, err)

	for i := 0; i < b.N; i++ {
		// Ensure inserted items is in the accepted range (0 -> 1<<itemSize)
		val := getRandInt() % (1 << arr.ItemSize)
		arr.SetItem(i, val)
		expected = append(expected, val)
	}

	b.StopTimer()
	items := arr.GetItems()
	assert.Equal(b, expected, items)
}

func TestCompactArray_String(t *testing.T) {
	itemArray, err := NewCompactArray(2, 15)
	assert.NoError(t, err)
	itemArray.SetItem(0, 3)
	itemArray.SetItem(1, 3)
	itemArray.SetItem(0, 8)

	assert.Equal(t, "[8, 3, ]", itemArray.String())
}

func TestCompactArray_Size(t *testing.T) {
	itemArray, err := NewCompactArray(5000, 7)
	assert.NoError(t, err)
	raw, err := json.Marshal(itemArray)
	assert.NoError(t, err)
	assert.Equal(t, 348, len(raw))
}

func TestPanicOnOutOfBounds(t *testing.T) {
	itemArray, err := NewCompactArray(5, 3)
	assert.NoError(t, err)
	assert.NotPanics(t, func() { itemArray.validateIndex(0) })
	assert.NotPanics(t, func() { itemArray.validateIndex(4) })
	assert.Panics(t, func() { itemArray.validateIndex(-1) })
	assert.Panics(t, func() { itemArray.validateIndex(5) })
}

func TestPanicOnOversizedValue(t *testing.T) {
	itemArray, err := NewCompactArray(5, 3)
	assert.NoError(t, err)
	assert.Panics(t, func() { itemArray.validateValue(4) })
	for i := 0; i < 4; i++ {
		assert.NotPanics(t, func() { itemArray.validateValue(Item(i)) })
	}
}

func TestBitsNeededCalculation(t *testing.T) {
	assert.Equal(t, uint(1), bitsNeeded(1))
	assert.Equal(t, uint(2), bitsNeeded(2))
	assert.Equal(t, uint(2), bitsNeeded(3))
	assert.Equal(t, uint(3), bitsNeeded(4))
	assert.Equal(t, uint(3), bitsNeeded(5))
	assert.Equal(t, uint(3), bitsNeeded(6))
	assert.Equal(t, uint(3), bitsNeeded(7))
	assert.Equal(t, uint(4), bitsNeeded(8))

}
