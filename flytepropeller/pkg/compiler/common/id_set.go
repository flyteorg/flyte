package common

import (
	"sort"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
)

type Empty struct{}
type Identifier = core.Identifier
type IdentifierSet map[string]Identifier

// NewString creates a String from a list of values.
func NewIdentifierSet(items ...Identifier) IdentifierSet {
	ss := IdentifierSet{}
	ss.Insert(items...)
	return ss
}

// Insert adds items to the set.
func (s IdentifierSet) Insert(items ...Identifier) {
	for _, item := range items {
		s[item.String()] = item
	}
}

// Delete removes all items from the set.
func (s IdentifierSet) Delete(items ...Identifier) {
	for _, item := range items {
		delete(s, item.String())
	}
}

// Has returns true if and only if item is contained in the set.
func (s IdentifierSet) Has(item Identifier) bool {
	_, contained := s[item.String()]
	return contained
}

// HasAll returns true if and only if all items are contained in the set.
func (s IdentifierSet) HasAll(items ...Identifier) bool {
	for _, item := range items {
		if !s.Has(item) {
			return false
		}
	}
	return true
}

// HasAny returns true if any items are contained in the set.
func (s IdentifierSet) HasAny(items ...Identifier) bool {
	for _, item := range items {
		if s.Has(item) {
			return true
		}
	}
	return false
}

type sortableSliceOfString []Identifier

func (s sortableSliceOfString) Len() int { return len(s) }
func (s sortableSliceOfString) Less(i, j int) bool {
	first, second := s[i], s[j]
	if first.ResourceType != second.ResourceType {
		return first.ResourceType < second.ResourceType
	}

	if first.Project != second.Project {
		return first.Project < second.Project
	}

	if first.Domain != second.Domain {
		return first.Domain < second.Domain
	}

	if first.Name != second.Name {
		return first.Name < second.Name
	}

	if first.Version != second.Version {
		return first.Version < second.Version
	}

	return false
}

func (s sortableSliceOfString) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

// List returns the contents as a sorted Identifier slice.
func (s IdentifierSet) List() []Identifier {
	res := make(sortableSliceOfString, 0, len(s))
	for _, value := range s {
		res = append(res, value)
	}

	sort.Sort(res)
	return []Identifier(res)
}
