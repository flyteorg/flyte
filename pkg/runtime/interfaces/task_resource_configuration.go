package interfaces

type TaskResourceSet struct {
	CPU              string `json:"cpu"`
	GPU              string `json:"gpu"`
	Memory           string `json:"memory"`
	Storage          string `json:"storage"`
	EphemeralStorage string `json:"ephemeralStorage"`
}

// Provides default values for task resource limits and defaults.
type TaskResourceConfiguration interface {
	GetDefaults() TaskResourceSet
	GetLimits() TaskResourceSet
}
