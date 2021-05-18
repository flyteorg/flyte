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
	return implementsAnyOfMethods(t, "UnmarshalJSON")
}

func isJSONMarshaler(t types.Type) bool {
	return implementsAnyOfMethods(t, "MarshalJSON")
}

func isStringer(t types.Type) bool {
	return implementsAnyOfMethods(t, "String")
}

func implementsAnyOfMethods(t types.Type, methodNames ...string) (found bool) {
	mset := types.NewMethodSet(t)
	for _, name := range methodNames {
		if mset.Lookup(nil, name) != nil {
			return true
		}
	}

	mset = types.NewMethodSet(types.NewPointer(t))
	for _, name := range methodNames {
		if mset.Lookup(nil, name) != nil {
			return true
		}
	}

	return false
}
