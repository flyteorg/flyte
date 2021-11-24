package api

import (
	"encoding/json"
	"errors"

	"github.com/flyteorg/flytestdlib/storage"
)

var DefaultTestType = &TestType{
	StringValue: "Welcome to defaults",
}

type TestType struct {
	StringValue      string             `json:"str" pflag:"\"hello world\",\"life is short\""`
	BoolValue        bool               `json:"bl" pflag:"true"`
	NestedType       NestedType         `json:"nested"`
	IntArray         []int              `json:"ints" pflag:"[]int{12%2C1}"`
	StringArray      []string           `json:"strs" pflag:"[]string{\"12\"%2C\"1\"}"`
	ComplexJSONArray []ComplexJSONType  `json:"complexArr"`
	StringToJSON     ComplexJSONType    `json:"c" pflag:",I'm a complex type but can be converted from string."`
	IgnoredMap       map[string]string  `json:"ignored-map" pflag:"-,"`
	StorageConfig    storage.Config     `json:"storage"`
	IntValue         *int               `json:"i"`
	StringMap        map[string]string  `json:"m" pflag:",I'm a map of strings"`
	ConstType        NamedType          `json:"constType"`
	AliasType        TestConstTypeAlias `json:"aliasType"`
}

//go:generate enumer --type=NamedType --trimprefix=NamedType -json

type NamedType int

const (
	NamedTypeA NamedType = iota
	NamedTypeB
)

type TestConstTypeAlias = int

type NestedType struct {
	IntValue int `json:"i" pflag:",this is an important flag"`
}

type ComplexJSONType struct {
	StringValue string `json:"str"`
	IntValue    int    `json:"i"`
}

func (c *ComplexJSONType) UnmarshalJSON(b []byte) error {
	if len(b) == 0 {
		c.StringValue = ""
		return nil
	}

	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	switch value := v.(type) {
	case string:
		if len(value) == 0 {
			c.StringValue = ""
		} else {
			c.StringValue = value
		}
	default:
		return errors.New("invalid duration")
	}

	return nil
}
