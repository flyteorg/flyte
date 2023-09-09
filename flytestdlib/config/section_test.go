package config

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/util/rand"

	"github.com/flyteorg/flytestdlib/internal/utils"
	"github.com/ghodss/yaml"
	"github.com/spf13/pflag"

	"github.com/stretchr/testify/assert"
)

// Make sure existing config file(s) parse correctly before overriding them with this flag!
var update = flag.Bool("update", false, "Updates testdata")

type MyComponentConfig struct {
	StringValue string `json:"str"`
}

type OtherComponentConfig struct {
	DurationValue Duration `json:"duration-value"`
	URLValue      URL      `json:"url-value"`
	StringValue   string   `json:"string-value"`
	IntValue      int      `json:"int-val"`
	StringArray   []string `json:"strings"`
}

func (MyComponentConfig) GetPFlagSet(prefix string) *pflag.FlagSet {
	cmdFlags := pflag.NewFlagSet("MyComponentConfig", pflag.ExitOnError)
	cmdFlags.String(fmt.Sprintf("%v%v", prefix, "str"), "hello world", "life is short")
	return cmdFlags
}

func (OtherComponentConfig) GetPFlagSet(prefix string) *pflag.FlagSet {
	cmdFlags := pflag.NewFlagSet("MyComponentConfig", pflag.ExitOnError)
	cmdFlags.String(fmt.Sprintf("%v%v", prefix, "string-value"), "hello world", "life is short")
	cmdFlags.String(fmt.Sprintf("%v%v", prefix, "duration-value"), "20s", "")
	cmdFlags.Int(fmt.Sprintf("%v%v", prefix, "int-val"), 4, "this is an important flag")
	cmdFlags.String(fmt.Sprintf("%v%v", prefix, "url-value"), "http://blah.com", "Sets the type of storage to configure [s3/minio/local/mem].")
	return cmdFlags
}

type TestConfig struct {
	MyComponentConfig    MyComponentConfig    `json:"my-component"`
	OtherComponentConfig OtherComponentConfig `json:"other-component"`
}

func TestMarshal(t *testing.T) {
	expected := TestConfig{
		MyComponentConfig: MyComponentConfig{
			StringValue: "Hello World",
		},
		OtherComponentConfig: OtherComponentConfig{
			StringValue:   "Hey there!",
			IntValue:      4,
			URLValue:      URL{URL: utils.MustParseURL("http://something.com")},
			DurationValue: Duration{Duration: time.Second * 20},
			StringArray:   []string{"hello", "world", "!"},
		},
	}

	configPath := filepath.Join("testdata", "config.yaml")
	if *update {
		t.Log("Updating config file.")
		raw, err := yaml.Marshal(expected)
		assert.NoError(t, err)
		assert.NoError(t, ioutil.WriteFile(configPath, raw, os.ModePerm))
	}

	r := TestConfig{}
	raw, err := ioutil.ReadFile(configPath)
	assert.NoError(t, err)
	assert.NoError(t, yaml.Unmarshal(raw, &r))
	assert.True(t, reflect.DeepEqual(expected, r))
}

func TestRegisterSection(t *testing.T) {
	t.Run("New Section", func(t *testing.T) {
		_, err := RegisterSection(rand.String(6), &TestConfig{})
		assert.NoError(t, err)
	})

	t.Run("Duplicate", func(t *testing.T) {
		s := rand.String(6)
		_, err := RegisterSection(s, &TestConfig{})
		assert.NoError(t, err)
		_, err = RegisterSection(s, &TestConfig{})
		assert.Error(t, err)
	})

	t.Run("Register Nested", func(t *testing.T) {
		root := NewRootSection()
		s := rand.String(6)
		_, err := root.RegisterSection(s, &TestConfig{})
		assert.NoError(t, err)
		_, err = root.RegisterSection(s, &TestConfig{})
		assert.Error(t, err)
	})
}

func TestGetSection(t *testing.T) {
	sectionName := rand.String(6)
	actual1, err := RegisterSection(sectionName, &TestConfig{})
	assert.NoError(t, err)
	assert.Equal(t, reflect.TypeOf(&TestConfig{}), reflect.TypeOf(actual1.GetConfig()))

	actual2 := GetSection(sectionName)
	assert.NotNil(t, actual2)
	assert.Equal(t, reflect.TypeOf(&TestConfig{}), reflect.TypeOf(actual2.GetConfig()))
}
