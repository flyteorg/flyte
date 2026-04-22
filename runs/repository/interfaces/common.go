package interfaces

// ListResourceInput contains parameters for querying collections of resources.
type ListResourceInput struct {
	Limit int

	// CursorToken is a keyset pagination cursor encoded as a RFC3339Nano timestamp.
	// When set, the query returns rows with created_at strictly greater than the cursor value.
	// Mutually exclusive with Offset.
	CursorToken string

	// Offset is an integer offset for offset-based pagination.
	Offset int

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
