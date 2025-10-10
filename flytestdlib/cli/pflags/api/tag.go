package api

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/fatih/structtag"
)

const (
	TagName     = "pflag"
	JSONTagName = "json"
)

// Tag represents parsed PFlag Go-struct tag.
// type Foo struct {
// StringValue string `json:"str" pflag:"\"hello world\",This is a string value"`
// }
// Name will be "str", Default value is "hello world" and Usage is "This is a string value"
type Tag struct {
	Name         string
	DefaultValue string
	Usage        string
}

// ParseTag parses tag. Name is computed from json tag, defaultvalue is the name of the pflag tag and usage is the concatenation
// of all options for pflag tag.
// e.g. `json:"name" pflag:"2,this is a useful param"`
func ParseTag(tag string) (t Tag, err error) {
	tags, err := structtag.Parse(tag)
	if err != nil {
		return Tag{}, err
	}

	t = Tag{}

	jsonTag, err := tags.Get(JSONTagName)
	if err == nil {
		t.Name = jsonTag.Name
	}

	pflagTag, err := tags.Get(TagName)
	if err == nil {
		t.DefaultValue, err = url.QueryUnescape(pflagTag.Name)
		if err != nil {
			fmt.Printf("Failed to Query unescape tag name [%v], will use value as is. Error: %v", pflagTag.Name, err)
			t.DefaultValue = pflagTag.Name
		}

		t.Usage = strings.Join(pflagTag.Options, ", ")
		if len(t.Usage) == 0 {
			t.Usage = `""`
		}

		if t.Usage[0] != '"' {
			t.Usage = fmt.Sprintf(`"%v"`, t.Usage)
		}
	} else {
		// We receive an error when the tag isn't present (or is malformed). Because there is no strongly-typed way to
		// do that, we will just set Usage to empty string and move on.
		t.Usage = `""`
	}

	return t, nil
}
