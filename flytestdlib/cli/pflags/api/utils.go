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

func jsonUnmarshaler(t types.Type) bool {
	mset := types.NewMethodSet(t)
	jsonUnmarshaler := mset.Lookup(nil, "UnmarshalJSON")
	if jsonUnmarshaler == nil {
		mset = types.NewMethodSet(types.NewPointer(t))
		jsonUnmarshaler = mset.Lookup(nil, "UnmarshalJSON")
	}

	return jsonUnmarshaler != nil
}
