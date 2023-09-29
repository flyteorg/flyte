/*
 * Copyright (c) 2018 Lyft. All rights reserved.
 */

package errorcollector

type rangeStack struct {
	top *rangeNode
}

type rangeNode struct {
	*indexRange
	next *rangeNode
}

func (s *rangeStack) Push(new *indexRange) {
	s.top = &rangeNode{indexRange: new, next: s.top}
}

func (s *rangeStack) Pop() *indexRange {
	top := s.top
	if top == nil {
		return nil
	}

	s.top = s.top.next
	return top.indexRange
}

func (s *rangeStack) Top() *indexRange {
	if s.top == nil {
		return nil
	}

	return s.top.indexRange
}

func (s *rangeStack) List() []*indexRange {
	res := make([]*indexRange, 0)
	for {
		top := s.Pop()
		if top == nil {
			break
		}

		res = append(res, top)
	}

	return res
}
