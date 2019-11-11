/*
 * Copyright (c) 2018 Lyft. All rights reserved.
 */

package errorcollector

import "sort"

type indexRangeCollection []*indexRange

func (c *indexRangeCollection) Len() int {
	return len(*c)
}

func (c *indexRangeCollection) Less(i, j int) bool {
	first := (*c)[i]
	second := (*c)[j]
	if first.start != second.start {
		return first.start < second.start
	}

	return first.end < second.end
}

func (c *indexRangeCollection) Swap(i, j int) {
	temp := (*c)[i]
	(*c)[i] = (*c)[j]
	(*c)[j] = temp
}

func (c *indexRangeCollection) Add(idx int) {
	newRange := &indexRange{start: idx, end: idx}
	for _, r := range *c {
		if r.CanMerge(*newRange) {
			r.MergeFrom(*newRange)
			return
		}
	}

	*c = append(*c, newRange)
}

func (c *indexRangeCollection) simplify() {
	hasImproved := false
	s := rangeStack{}

	sort.Sort(c)

	// push the first interval to stack
	s.Push((*c)[0])

	// Start from the next interval and merge if necessary
	for i := 1; i < len(*c); i++ {
		// get interval from stack top
		top := s.Top()
		if top.CanMerge(*(*c)[i]) {
			top.MergeFrom(*(*c)[i])
			hasImproved = true
		} else {
			s.Push((*c)[i])
		}
	}

	if hasImproved {
		*c = s.List()

		// Keep simplifying until we have nothing to simplify
		c.simplify()
	}
}

func (c indexRangeCollection) String() string {
	c.simplify()
	res := ""
	for _, r := range c {
		res += r.String()
	}

	return res
}
