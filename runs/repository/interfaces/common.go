package interfaces

import "time"

// ListResourceInput contains parameters for querying collections of resources.
type ListResourceInput struct {
	Limit int

	// CursorToken is reserved for an opaque keyset-cursor pagination mode. It is currently
	// unused — the ListRuns/ListActions RPC paginates by Offset (see below) — and no repo
	// reads it. Kept as a placeholder so a future keyset cursor can be wired in without an
	// interface change.
	CursorToken string

	// Offset is an integer offset for offset-based pagination. actionRepo.ListActions
	// applies it as SQL OFFSET (mutually exclusive with KeysetAfter) and rejects a negative
	// value; the ListRuns/ListActions RPC encodes the running offset as the page token. The
	// other repos' List methods use it too.
	Offset int

	// KeysetAfterCreatedAt/KeysetAfterName do ascending composite keyset pagination:
	// when KeysetAfterCreatedAt is non-nil, the query returns rows ordered strictly
	// after (created_at, name) = (*KeysetAfterCreatedAt, KeysetAfterName). The caller
	// must sort by (created_at ASC, name ASC). Mutually exclusive with Offset.
	// Used by the WatchActions snapshot to page a run's actions in O(n).
	KeysetAfterCreatedAt *time.Time
	KeysetAfterName      string

	Filter Filter
	// The filter set by scopeBy in the query
	ScopeByFilter  Filter
	SortParameters []SortParameter
}

func (l ListResourceInput) WithFilter(filter Filter) ListResourceInput {
	if l.Filter != nil {
		l.Filter = l.Filter.And(filter)
	} else {
		l.Filter = filter
	}

	return l
}

func (l ListResourceInput) WithSortParameters(sortParameters ...SortParameter) ListResourceInput {
	if l.SortParameters == nil {
		l.SortParameters = make([]SortParameter, 0, len(sortParameters))
	}

	l.SortParameters = append(l.SortParameters, sortParameters...)
	return l
}
