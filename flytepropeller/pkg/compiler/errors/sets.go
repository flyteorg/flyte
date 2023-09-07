package errors

import (
	"sort"
	"strings"
)

var keyExists = struct{}{}

type compileErrorSet map[CompileError]struct{}

func (s compileErrorSet) Put(key CompileError) {
	s[key] = keyExists
}

func (s compileErrorSet) Contains(key CompileError) bool {
	_, ok := s[key]
	return ok
}

func (s compileErrorSet) Remove(key CompileError) {
	delete(s, key)
}

func refCompileError(x CompileError) *CompileError {
	return &x
}

func (s compileErrorSet) List() []*CompileError {
	res := make([]*CompileError, 0, len(s))
	for key := range s {
		res = append(res, refCompileError(key))
	}

	sort.SliceStable(res, func(i, j int) bool {
		if res[i].Code() == res[j].Code() {
			return res[i].Error() < res[j].Error()
		}

		return strings.Compare(string(res[i].Code()), string(res[j].Code())) < 0
	})

	return res
}
