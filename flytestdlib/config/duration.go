package config

import (
	"encoding/json"
	"errors"
	"time"
)

// Duration is a wrapper around time.Duration that enables Json Marshalling capabilities
type Duration struct {
	time.Duration
}

func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

func (d *Duration) UnmarshalJSON(b []byte) error {
	if len(b) == 0 {
		d.Duration = time.Duration(0)
		return nil
	}

	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	switch value := v.(type) {
	case float64:
		d.Duration = time.Duration(value)
		return nil
	case string:
		if len(value) == 0 {
			d.Duration = time.Duration(0)
		} else {
			var err error
			d.Duration, err = time.ParseDuration(value)
			if err != nil {
				return err
			}
		}
	default:
		return errors.New("invalid duration")
	}

	return nil
}

// Set implements PFlag's Value interface's set method to set the value of duration from string.
func (d *Duration) Set(val string) error {
	var err error
	d.Duration, err = time.ParseDuration(val)
	return err
}

// Type implements PFlag's Value interface's Type method to return the name of the type.
func (Duration) Type() string {
	return "Duration"
}
