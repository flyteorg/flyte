package config

import (
	"encoding/json"
	"errors"
	"net/url"
)

// A url.URL wrapper that can marshal and unmarshal into simple URL strings.
type URL struct {
	url.URL
}

func (d URL) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

func (d *URL) UnmarshalJSON(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}

	switch value := v.(type) {
	case string:
		u, err := url.Parse(value)
		if err != nil {
			return err
		}

		d.URL = *u
		return nil
	default:
		return errors.New("invalid url")
	}
}

// Set implements PFlag's Value interface's set method to set the value of duration from string.
func (d *URL) Set(val string) error {
	u, err := url.Parse(val)
	if err != nil {
		return err
	}

	d.URL = *u
	return nil
}

// Type implements PFlag's Value interface's Type method to return the name of the type.
func (URL) Type() string {
	return "URL"
}
