package cmd

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
)

// Represents a pflag value that parses a string into a map
type stringMapValue struct {
	value   *map[string]string
	changed bool
}

func newStringMapValue() *stringMapValue {
	return &stringMapValue{
		value:   &map[string]string{},
		changed: false,
	}
}

var entryRegex = regexp.MustCompile("(?P<EntryKey>[^,]+)=(?P<EntryValue>[^,]+)")

// Set parses val into a map. Accepted format: a=1,b=2
func (s *stringMapValue) Set(val string) error {
	matches := entryRegex.FindAllStringSubmatch(val, -1)
	out := make(map[string]string, len(matches))
	for _, entry := range matches {
		if len(entry) != 3 {
			return fmt.Errorf("invalid value for entry. Entries must be formatted as key=value. Found %v",
				entry)
		}

		out[strings.TrimSpace(entry[1])] = strings.TrimSpace(entry[2])
	}

	if !s.changed {
		*s.value = out
	} else {
		for k, v := range out {
			(*s.value)[k] = v
		}
	}
	s.changed = true
	return nil
}

func (s *stringMapValue) Type() string {
	return "stringToString"
}

func (s *stringMapValue) String() string {
	var buf bytes.Buffer
	i := 0
	for k, v := range *s.value {
		if i > 0 {
			_, err := buf.WriteRune(',')
			if err != nil {
				return ""
			}
		}

		_, err := buf.WriteString(k)
		if err != nil {
			return ""
		}

		_, err = buf.WriteRune('=')
		if err != nil {
			return ""
		}

		_, err = buf.WriteString(v)
		if err != nil {
			return ""
		}

		i++
	}

	return "[" + buf.String() + "]"
}
