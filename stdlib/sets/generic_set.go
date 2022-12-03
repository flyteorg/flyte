package sets

import (
	"sort"
)

type SetObject interface {
	GetID() string
}

type Generic map[string]SetObject

// New creates a Generic from a list of values.
func NewGeneric(items ...SetObject) Generic {
	gs := Generic{}
	gs.Insert(items...)
	return gs
}

// Insert adds items to the set.
func (g Generic) Insert(items ...SetObject) {
	for _, item := range items {
		g[item.GetID()] = item
	}
}

// Delete removes all items from the set.
func (g Generic) Delete(items ...SetObject) {
	for _, item := range items {
		delete(g, item.GetID())
	}
}

// Has returns true if and only if item is contained in the set.
func (g Generic) Has(item SetObject) bool {
	_, contained := g[item.GetID()]
	return contained
}

// HasAll returns true if and only if all items are contained in the set.
func (g Generic) HasAll(items ...SetObject) bool {
	for _, item := range items {
		if !g.Has(item) {
			return false
		}
	}
	return true
}

// HasAny returns true if any items are contained in the set.
func (g Generic) HasAny(items ...SetObject) bool {
	for _, item := range items {
		if g.Has(item) {
			return true
		}
	}
	return false
}

// Difference returns a set of objects that are not in s2
// For example:
// s1 = {a1, a2, a3}
// s2 = {a1, a2, a4, a5}
// s1.Difference(s2) = {a3}
// s2.Difference(s1) = {a4, a5}
func (g Generic) Difference(g2 Generic) Generic {
	result := NewGeneric()
	for _, v := range g {
		if !g2.Has(v) {
			result.Insert(v)
		}
	}
	return result
}

// Union returns a new set which includes items in either s1 or s2.
// For example:
// s1 = {a1, a2}
// s2 = {a3, a4}
// s1.Union(s2) = {a1, a2, a3, a4}
// s2.Union(s1) = {a1, a2, a3, a4}
func (g Generic) Union(s2 Generic) Generic {
	result := NewGeneric()
	for _, v := range g {
		result.Insert(v)
	}
	for _, v := range s2 {
		result.Insert(v)
	}
	return result
}

// Intersection returns a new set which includes the item in BOTH s1 and s2
// For example:
// s1 = {a1, a2}
// s2 = {a2, a3}
// s1.Intersection(s2) = {a2}
func (g Generic) Intersection(s2 Generic) Generic {
	var walk, other Generic
	result := NewGeneric()
	if g.Len() < s2.Len() {
		walk = g
		other = s2
	} else {
		walk = s2
		other = g
	}
	for _, v := range walk {
		if other.Has(v) {
			result.Insert(v)
		}
	}
	return result
}

// IsSuperset returns true if and only if s1 is a superset of s2.
func (g Generic) IsSuperset(s2 Generic) bool {
	for _, v := range s2 {
		if !g.Has(v) {
			return false
		}
	}
	return true
}

// Equal returns true if and only if s1 is equal (as a set) to s2.
// Two sets are equal if their membership is identical.
// (In practice, this means same elements, order doesn't matter)
func (g Generic) Equal(s2 Generic) bool {
	return len(g) == len(s2) && g.IsSuperset(s2)
}

type sortableSliceOfGeneric []string

func (s sortableSliceOfGeneric) Len() int           { return len(s) }
func (s sortableSliceOfGeneric) Less(i, j int) bool { return lessString(s[i], s[j]) }
func (s sortableSliceOfGeneric) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

// List returns the contents as a sorted string slice.
func (g Generic) ListKeys() []string {
	res := make(sortableSliceOfGeneric, 0, len(g))
	for key := range g {
		res = append(res, key)
	}
	sort.Sort(res)
	return []string(res)
}

// List returns the contents as a sorted string slice.
func (g Generic) List() []SetObject {
	keys := g.ListKeys()
	res := make([]SetObject, 0, len(keys))
	for _, k := range keys {
		s := g[k]
		res = append(res, s)
	}
	return res
}

// UnsortedList returns the slice with contents in random order.
func (g Generic) UnsortedListKeys() []string {
	res := make([]string, 0, len(g))
	for key := range g {
		res = append(res, key)
	}
	return res
}

// UnsortedList returns the slice with contents in random order.
func (g Generic) UnsortedList() []SetObject {
	res := make([]SetObject, 0, len(g))
	for _, v := range g {
		res = append(res, v)
	}
	return res
}

// Returns a single element from the set.
func (g Generic) PopAny() (SetObject, bool) {
	for _, v := range g {
		g.Delete(v)
		return v, true
	}
	var zeroValue SetObject
	return zeroValue, false
}

// Len returns the size of the set.
func (g Generic) Len() int {
	return len(g)
}

func lessString(lhs, rhs string) bool {
	return lhs < rhs
}
