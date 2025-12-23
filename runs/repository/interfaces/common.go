package interfaces

// ListResourceInput contains parameters for querying collections of resources.
type ListResourceInput struct {
	Limit  int
	Offset int

	Filter            Filter
	// The filter set by scopeBy in the query
	ScopeByFilter  Filter
	SortParameters []SortParameter
}

func (l ListResourceInput) WithFilter(filter Filter) ListResourceInput {
	if l.Filter != nil {
		l.Filter = l.Filter.And(filter)
	}

	l.Filter = filter
	return l
}

func (l ListResourceInput) WithSortParameters(sortParameters ...SortParameter) ListResourceInput {
	if l.SortParameters == nil {
		l.SortParameters = make([]SortParameter, 0, len(sortParameters))
	}

	l.SortParameters = append(l.SortParameters, sortParameters...)
	return l
}
