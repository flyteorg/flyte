package config

import (
	"encoding/json"
	"errors"
	"regexp"
)

// A regexp.Regexp wrapper that can marshal and unmarshal into simple regexp string.
type Regexp struct {
	regexp.Regexp
}

func (r Regexp) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.String())
}

func (r *Regexp) UnmarshalJSON(b []byte) error {
	if len(b) == 0 {
		r.Regexp = regexp.Regexp{}
		return nil
	}

	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}

	switch value := v.(type) {
	case string:
		rc, err := regexp.Compile(value)
		if err != nil {
			return err
		}

		r.Regexp = *rc
		return nil
	default:
		return errors.New("invalid regexp")
	}
}
