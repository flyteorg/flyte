package api

import (
	"bytes"
	"fmt"
	"go/types"
	"unicode"
)

func camelCase(str string) string {
	if len(str) == 0 {
		return str
	}

	firstRune := bytes.Runes([]byte(str))[0]
	if unicode.IsLower(firstRune) {
		return fmt.Sprintf("%v%v", string(unicode.ToUpper(firstRune)), str[1:])
	}

	return str
}

func isJSONUnmarshaler(t types.Type) bool {
	found, _ := implementsAnyOfMethods(t, "UnmarshalJSON")
	return found
}

func isStringer(t types.Type) bool {
	found, _ := implementsAnyOfMethods(t, "String")
	return found
}

func implementsAnyOfMethods(t types.Type, methodNames ...string) (found, implementedByPtr bool) {
	mset := types.NewMethodSet(t)
	for _, name := range methodNames {
		if mset.Lookup(nil, name) != nil {
			return true, false
		}
	}
	mset = types.NewMethodSet(types.NewPointer(t))
	for _, name := range methodNames {
		if mset.Lookup(nil, name) != nil {
			return true, true
		}
	}

	return false, false
}
