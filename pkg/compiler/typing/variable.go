package typing

import (
	"fmt"
	"regexp"
	"strconv"
)

var arrayVarMatcher = regexp.MustCompile(`(\[(?P<index>\d+)\]\.)?(?P<var>\w+)`)

type Variable struct {
	Name  string
	Index *int
}

// ParseVarName parses var names
func ParseVarName(varName string) (v Variable, err error) {
	allMatches := arrayVarMatcher.FindAllStringSubmatch(varName, -1)
	if len(allMatches) != 1 {
		return Variable{}, fmt.Errorf("unexpected number of matches [%v]", len(allMatches))
	}

	if len(allMatches[0]) != 4 {
		return Variable{}, fmt.Errorf("unexpected number of groups [%v]", len(allMatches[0]))
	}

	res := Variable{}
	if len(allMatches[0][2]) > 0 {
		index, convErr := strconv.Atoi(allMatches[0][2])
		err = convErr
		res.Index = &index
	}

	res.Name = allMatches[0][3]
	return res, err
}
