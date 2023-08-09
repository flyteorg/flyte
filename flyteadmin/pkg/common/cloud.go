package common

// Common configuration parameters for initializing back-end cloud clients.

type CloudProvider = string

const (
	AWS     CloudProvider = "aws"
	GCP     CloudProvider = "gcp"
	Sandbox CloudProvider = "sandbox"
	Local   CloudProvider = "local"
	None    CloudProvider = "none"
)
