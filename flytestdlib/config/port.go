package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
)

// A common port struct that supports Json marshal/unmarshal into/from simple strings/floats.
type Port struct {
	Port int `json:"port,omitempty"`
}

func (p Port) String() string {
	return strconv.Itoa(p.Port)
}

func (p Port) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.Port)
}

func (p *Port) UnmarshalJSON(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}

	switch value := v.(type) {
	case string:
		u, err := parsePortString(value)
		if err != nil {
			return err
		}

		p.Port = u
		return nil
	case float64:
		if !validPortRange(value) {
			return fmt.Errorf("port must be a valid number between 0 and 65535, inclusive")
		}

		p.Port = int(value)
		return nil
	default:
		return errors.New("invalid port")
	}
}

func parsePortString(port string) (int, error) {
	if portInt, err := strconv.Atoi(port); err == nil && validPortRange(float64(portInt)) {
		return portInt, nil
	}

	return 0, fmt.Errorf("port must be a valid number between 1 and 65535, inclusive")
}

func validPortRange(port float64) bool {
	return 0 <= port && port <= 65535
}

// Set implements PFlag's Value interface's set method to set the value of duration from string.
func (p *Port) Set(val string) error {
	u, err := parsePortString(val)
	if err != nil {
		return err
	}

	p.Port = u
	return nil
}

// Type implements PFlag's Value interface's Type method to return the name of the type.
func (Port) Type() string {
	return "URL"
}
