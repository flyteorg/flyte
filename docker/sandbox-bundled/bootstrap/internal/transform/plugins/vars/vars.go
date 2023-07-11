package vars

import (
	"strings"
)

type ValueGetter func() (string, error)

type Vars struct {
	internal map[string]ValueGetter
}

func NewVars(values map[string]ValueGetter) *Vars {
	return &Vars{internal: values}
}

func (v *Vars) Transform(data []byte) ([]byte, error) {
	var tokens []string
	for key, valueGetter := range v.internal {
		value, err := valueGetter()
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, []string{key, value}...)
	}
	replacer := strings.NewReplacer(tokens...)
	return []byte(replacer.Replace(string(data))), nil
}
