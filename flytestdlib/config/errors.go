package config

import "fmt"

var (
	ErrStrictModeValidation       = fmt.Errorf("failed strict mode check")
	ErrChildConfigOverridesConfig = fmt.Errorf("child config attempts to override an existing native config property")
)

// A helper object that collects errors.
type ErrorCollection []error

func (e ErrorCollection) Error() string {
	res := ""
	for _, err := range e {
		res = fmt.Sprintf("%v\n%v", res, err.Error())
	}

	return res
}

func (e ErrorCollection) ErrorOrDefault() error {
	if len(e) == 0 {
		return nil
	}

	return e
}

func (e *ErrorCollection) Append(err error) bool {
	if err != nil {
		*e = append(*e, err)
		return true
	}

	return false
}
