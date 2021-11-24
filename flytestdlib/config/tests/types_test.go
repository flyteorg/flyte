package tests

import (
	"fmt"

	"encoding/json"

	"github.com/flyteorg/flytestdlib/config"
	"github.com/flyteorg/flytestdlib/config/viper"
	"github.com/spf13/pflag"
)

const testNameFormatter = "[%v] %v"

var providers = []config.AccessorProvider{viper.NewAccessor}

type MyComponentConfig struct {
	StringValue  string `json:"str"`
	StringValue2 string `json:"str2"`
	StringValue3 string `json:"str3"`
}

type NamedType int

const (
	NamedTypeA NamedType = iota
	NamedTypeB
)

type OtherComponentConfig struct {
	DurationValue           config.Duration `json:"duration-value"`
	URLValue                config.URL      `json:"url-value"`
	StringValue             string          `json:"string-value"`
	IntValue                int             `json:"int-val"`
	StringArray             []string        `json:"strings"`
	StringArrayWithDefaults []string        `json:"strings-def"`
	MyByteArray             []byte          `json:"myByteArray"`
	NamedType               NamedType       `json:"namedType"`
}

type Item struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	RandomValue string `json:"randomValue"`
}

type ItemArray struct {
	Items     []Item `json:"items"`
	OtherItem Item   `json:"otherItem"`
}

type ItemMap struct {
	Items    map[string]Item            `json:"items"`
	ItemsMap map[string]map[string]Item `json:"itemsMap"`
}

func (MyComponentConfig) GetPFlagSet(prefix string) *pflag.FlagSet {
	cmdFlags := pflag.NewFlagSet("MyComponentConfig", pflag.ExitOnError)
	cmdFlags.String(fmt.Sprintf("%v%v", prefix, "str"), "hello world", "life is short")
	cmdFlags.String(fmt.Sprintf("%v%v", prefix, "str2"), "hello world", "life is short")
	cmdFlags.String(fmt.Sprintf("%v%v", prefix, "str3"), "hello world", "life is short")
	return cmdFlags
}

var (
	DefaultOtherComponentConfig = &OtherComponentConfig{
		NamedType: NamedTypeA,
	}
)

func (OtherComponentConfig) GetPFlagSet(prefix string) *pflag.FlagSet {
	cmdFlags := pflag.NewFlagSet("MyComponentConfig", pflag.ExitOnError)
	cmdFlags.String(fmt.Sprintf("%v%v", prefix, "string-value"), "hello world", "life is short")
	cmdFlags.String(fmt.Sprintf("%v%v", prefix, "duration-value"), "20s", "")
	cmdFlags.Int(fmt.Sprintf("%v%v", prefix, "int-val"), 4, "this is an important flag")
	cmdFlags.String(fmt.Sprintf("%v%v", prefix, "url-value"), "http://blah.com", "Sets the type of storage to configure [s3/minio/local/mem].")
	cmdFlags.StringSlice(fmt.Sprintf("%v%v", prefix, "strings-def"), []string{"default value"}, "Sets the type of storage to configure [s3/minio/local/mem].")
	cmdFlags.Int(fmt.Sprintf("%v%v", prefix, "namedType"), int(DefaultOtherComponentConfig.NamedType), "")
	return cmdFlags
}

type TestConfig struct {
	MyComponentConfig    MyComponentConfig    `json:"my-component"`
	OtherComponentConfig OtherComponentConfig `json:"other-component"`
}

const (
	MyComponentSectionKey    = "my-component"
	OtherComponentSectionKey = "other-component"
)

func init() {
	if _, err := config.RegisterSection(MyComponentSectionKey, &MyComponentConfig{}); err != nil {
		panic(err)
	}

	if _, err := config.RegisterSection(OtherComponentSectionKey, &OtherComponentConfig{}); err != nil {
		panic(err)
	}
}

const _NamedTypeName = "AB"

var _NamedTypeIndex = [...]uint8{0, 1, 2}

func (i NamedType) String() string {
	if i < 0 || i >= NamedType(len(_NamedTypeIndex)-1) {
		return fmt.Sprintf("NamedType(%d)", i)
	}
	return _NamedTypeName[_NamedTypeIndex[i]:_NamedTypeIndex[i+1]]
}

var _NamedTypeValues = []NamedType{0, 1}

var _NamedTypeNameToValueMap = map[string]NamedType{
	_NamedTypeName[0:1]: 0,
	_NamedTypeName[1:2]: 1,
}

// NamedTypeString retrieves an enum value from the enum constants string name.
// Throws an error if the param is not part of the enum.
func NamedTypeString(s string) (NamedType, error) {
	if val, ok := _NamedTypeNameToValueMap[s]; ok {
		return val, nil
	}
	return 0, fmt.Errorf("%s does not belong to NamedType values", s)
}

// NamedTypeValues returns all values of the enum
func NamedTypeValues() []NamedType {
	return _NamedTypeValues
}

// IsANamedType returns "true" if the value is listed in the enum definition. "false" otherwise
func (i NamedType) IsANamedType() bool {
	for _, v := range _NamedTypeValues {
		if i == v {
			return true
		}
	}
	return false
}

// MarshalJSON implements the json.Marshaler interface for NamedType
func (i NamedType) MarshalJSON() ([]byte, error) {
	return json.Marshal(i.String())
}

// UnmarshalJSON implements the json.Unmarshaler interface for NamedType
func (i *NamedType) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("NamedType should be a string, got %s", data)
	}

	var err error
	*i, err = NamedTypeString(s)
	return err
}
