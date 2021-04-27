package tests

import (
	"fmt"

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

type OtherComponentConfig struct {
	DurationValue           config.Duration `json:"duration-value"`
	URLValue                config.URL      `json:"url-value"`
	StringValue             string          `json:"string-value"`
	IntValue                int             `json:"int-val"`
	StringArray             []string        `json:"strings"`
	StringArrayWithDefaults []string        `json:"strings-def"`
}

type Item struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type ItemArray struct {
	Items     []Item `json:"items"`
	OtherItem Item   `json:"otherItem"`
}

func (MyComponentConfig) GetPFlagSet(prefix string) *pflag.FlagSet {
	cmdFlags := pflag.NewFlagSet("MyComponentConfig", pflag.ExitOnError)
	cmdFlags.String(fmt.Sprintf("%v%v", prefix, "str"), "hello world", "life is short")
	cmdFlags.String(fmt.Sprintf("%v%v", prefix, "str2"), "hello world", "life is short")
	cmdFlags.String(fmt.Sprintf("%v%v", prefix, "str3"), "hello world", "life is short")
	return cmdFlags
}

func (OtherComponentConfig) GetPFlagSet(prefix string) *pflag.FlagSet {
	cmdFlags := pflag.NewFlagSet("MyComponentConfig", pflag.ExitOnError)
	cmdFlags.String(fmt.Sprintf("%v%v", prefix, "string-value"), "hello world", "life is short")
	cmdFlags.String(fmt.Sprintf("%v%v", prefix, "duration-value"), "20s", "")
	cmdFlags.Int(fmt.Sprintf("%v%v", prefix, "int-val"), 4, "this is an important flag")
	cmdFlags.String(fmt.Sprintf("%v%v", prefix, "url-value"), "http://blah.com", "Sets the type of storage to configure [s3/minio/local/mem].")
	cmdFlags.StringSlice(fmt.Sprintf("%v%v", prefix, "strings-def"), []string{"default value"}, "Sets the type of storage to configure [s3/minio/local/mem].")
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
