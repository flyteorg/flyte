package errors

// Represents error config that can change the behavior of how errors collection/reporting is handled.
type Config struct {
	// Indicates that a panic should be issued as soon as the first error is collected.
	PanicOnError bool

	// Indicates that errors should include source code information when collected. There is an associated performance
	// penalty with this behavior.
	IncludeSource bool
}

var config = Config{}

// Sets global config.
func SetConfig(cfg Config) {
	config = cfg
}

// Gets global config.
func GetConfig() Config {
	return config
}

func SetIncludeSource() {
	config.IncludeSource = true
}
