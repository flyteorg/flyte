package common

// Common configuration parameters for initializing back-end cloud clients.

type CloudProvider = string

const (
	AWS     CloudProvider = "aws"
	GCP     CloudProvider = "gcp"
	Azure   CloudProvider = "azure"
	Sandbox CloudProvider = "sandbox"
	Local   CloudProvider = "local"
	None    CloudProvider = "none"
)
