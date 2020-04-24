package common

// Common configuration parameters for initializing back-end cloud clients.

type CloudProvider = string

const (
	AWS   CloudProvider = "aws"
	GCP   CloudProvider = "gcp"
	Local CloudProvider = "local"
)
