package api

import (
	"bytes"
	"fmt"
	"go/types"
	"unicode"

	"k8s.io/apimachinery/pkg/util/sets"
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

func isPFlagValue(t types.Type) bool {
	return implementsAllOfMethods(t, "String", "Set", "Type")
}

func hasStringConstructor(t *types.Named) bool {
	return t.Obj().Parent().Lookup(fmt.Sprintf("%sString", t.Obj().Name())) != nil
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

func implementsAllOfMethods(t types.Type, methodNames ...string) (found bool) {
	mset := types.NewMethodSet(t)
	foundMethods := sets.NewString()
	for _, name := range methodNames {
		if foundMethod := mset.Lookup(nil, name); foundMethod != nil {
			foundMethods.Insert(name)
		}
	}

	mset = types.NewMethodSet(types.NewPointer(t))
	for _, name := range methodNames {
		if mset.Lookup(nil, name) != nil {
			foundMethods.Insert(name)
		}
	}

	return foundMethods.Len() == len(methodNames)
}
