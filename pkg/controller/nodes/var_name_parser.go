package nodes

import (
	"fmt"
	"regexp"
	"strconv"
)

var arrayVarMatcher = regexp.MustCompile(`(\[(?P<index>\d+)\]\.)?(?P<var>\w+)`)

// Parses var names
func ParseVarName(varName string) (idx *int, name string, err error) {
	allMatches := arrayVarMatcher.FindAllStringSubmatch(varName, -1)
	if len(allMatches) != 1 {
		return nil, "", fmt.Errorf("unexpected number of matches [%v]", len(allMatches))
	}

	if len(allMatches[0]) != 4 {
		return nil, "", fmt.Errorf("unexpected number of groups [%v]", len(allMatches[0]))
	}

	if len(allMatches[0][2]) > 0 {
		index, convErr := strconv.Atoi(allMatches[0][2])
		err = convErr
		idx = &index
	}

	name = allMatches[0][3]
	return
}
