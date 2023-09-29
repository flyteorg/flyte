package config

import "fmt"

var (
	ErrStrictModeValidation       = fmt.Errorf("failed strict mode check")
	ErrChildConfigOverridesConfig = fmt.Errorf("child config attempts to override an existing native config property")
)
