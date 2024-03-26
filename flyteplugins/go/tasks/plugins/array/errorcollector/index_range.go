/*
 * Copyright (c) 2018 Lyft. All rights reserved.
 */

package errorcollector

import "fmt"

type indexRange struct {
	start int
	end   int
}

func (r indexRange) String() string {
	if r.start == r.end {
		return fmt.Sprintf("[%v]", r.start)
	}

	return fmt.Sprintf("[%v-%v]", r.start, r.end)
}

func (r indexRange) CanMerge(other indexRange) bool {
	return (r.start <= other.start+1 && r.end >= other.start-1) ||
		(r.start <= other.end+1 && r.end >= other.end-1)
}

func (r *indexRange) MergeFrom(other indexRange) {
	if other.start < r.start {
		r.start = other.start
	}

	if other.end > r.end {
		r.end = other.end
	}
}
