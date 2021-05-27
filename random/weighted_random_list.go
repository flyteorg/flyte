package random

import (
	"context"
	"fmt"
	"math/rand"
	"sort"
	"time"

	"github.com/flyteorg/flytestdlib/logger"
)

//go:generate mockery -all -case=underscore

// Interface to use the Weighted Random
type WeightedRandomList interface {
	Get() Comparable
	GetWithSeed(seed rand.Source) (Comparable, error)
	List() []Comparable
	Len() int
}

// Interface for items that can be used along with WeightedRandomList
type Comparable interface {
	Compare(to Comparable) bool
}

// Structure of each entry to select from
type Entry struct {
	Item   Comparable
	Weight float32
}

type internalEntry struct {
	entry        Entry
	currentTotal float32
}

// WeightedRandomList selects elements randomly from the list taking into account individual weights.
// Weight has to be assigned between 0 and 1.
// Support deterministic results when given a particular seed source
type weightedRandomListImpl struct {
	entries     []internalEntry
	totalWeight float32
}

func validateEntries(entries []Entry) error {
	if len(entries) == 0 {
		return fmt.Errorf("entries is empty")
	}
	for index, entry := range entries {
		if entry.Item == nil {
			return fmt.Errorf("invalid entry: nil, index %d", index)
		}
		if entry.Weight < 0 || entry.Weight > float32(1) {
			return fmt.Errorf("invalid weight %f, index %d", entry.Weight, index)
		}
	}
	return nil
}

// Given a list of entries with weights, returns WeightedRandomList
func NewWeightedRandom(ctx context.Context, entries []Entry) (WeightedRandomList, error) {
	err := validateEntries(entries)
	if err != nil {
		return nil, err
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Item.Compare(entries[j].Item)
	})
	var internalEntries []internalEntry
	numberOfEntries := len(entries)
	totalWeight := float32(0)
	for _, e := range entries {
		totalWeight += e.Weight
	}

	currentTotal := float32(0)
	for _, e := range entries {
		if totalWeight == 0 {
			// This indicates that none of the entries have weight assigned.
			// We will assign equal weights to everyone
			currentTotal += 1.0 / float32(numberOfEntries)
		} else if e.Weight == 0 {
			// Entries which have zero weight are ignored
			logger.Debug(ctx, "ignoring entry due to empty weight %v", e)
			continue
		}

		currentTotal += e.Weight
		internalEntries = append(internalEntries, internalEntry{
			entry:        e,
			currentTotal: currentTotal,
		})
	}

	return &weightedRandomListImpl{
		entries:     internalEntries,
		totalWeight: currentTotal,
	}, nil
}

func (w *weightedRandomListImpl) get(generator *rand.Rand) Comparable {
	randomWeight := generator.Float32() * w.totalWeight
	for _, e := range w.entries {
		if e.currentTotal >= randomWeight && e.currentTotal > 0 {
			return e.entry.Item
		}
	}
	return w.entries[len(w.entries)-1].entry.Item
}

// Returns a random entry based on the weights
func (w *weightedRandomListImpl) Get() Comparable {
	/* #nosec */
	randGenerator := rand.New(rand.NewSource(time.Now().UTC().UnixNano()))
	return w.get(randGenerator)
}

// For a given seed, the same entry will be returned all the time.
func (w *weightedRandomListImpl) GetWithSeed(seed rand.Source) (Comparable, error) {
	/* #nosec */
	randGenerator := rand.New(seed)
	return w.get(randGenerator), nil
}

// Lists all the entries that are eligible for selection
func (w *weightedRandomListImpl) List() []Comparable {
	entries := make([]Comparable, len(w.entries))
	for index, indexedItem := range w.entries {
		entries[index] = indexedItem.entry.Item
	}
	return entries
}

// Gets the number of items that are being considered for selection.
func (w *weightedRandomListImpl) Len() int {
	return len(w.entries)
}
