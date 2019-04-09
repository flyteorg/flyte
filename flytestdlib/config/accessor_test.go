package config

import (
	"context"
	"fmt"
	"path/filepath"
)

func Example() {
	// This example demonstrates basic usage of config sections.

	//go:generate pflags OtherComponentConfig

	type OtherComponentConfig struct {
		DurationValue Duration `json:"duration-value"`
		URLValue      URL      `json:"url-value"`
		StringValue   string   `json:"string-value"`
	}

	// Each component should register their section in package init() or as a package var
	section := MustRegisterSection("other-component", &OtherComponentConfig{})

	// Override configpath to look for a custom location.
	configPath := filepath.Join("testdata", "config.yaml")

	// Initialize an accessor.
	var accessor Accessor
	// e.g.
	// accessor = viper.NewAccessor(viper.Options{
	//		StrictMode:   true,
	//		SearchPaths: []string{configPath, configPath2},
	// })

	// Optionally bind to Pflags.
	// accessor.InitializePflags(flags)

	// Parse config from file or pass empty to rely on env variables and PFlags
	err := accessor.UpdateConfig(context.Background())
	if err != nil {
		fmt.Printf("Failed to validate config from [%v], error: %v", configPath, err)
		return
	}

	// Get parsed config value.
	parsedConfig := section.GetConfig().(*OtherComponentConfig)
	fmt.Printf("Config: %v", parsedConfig)
}

func Example_nested() {
	// This example demonstrates registering nested config sections dynamically.

	//go:generate pflags OtherComponentConfig

	type OtherComponentConfig struct {
		DurationValue Duration `json:"duration-value"`
		URLValue      URL      `json:"url-value"`
		StringValue   string   `json:"string-value"`
	}

	// Each component should register their section in package init() or as a package var
	Section := MustRegisterSection("my-component", &MyComponentConfig{})

	// Other packages can register their sections at the root level (like the above line) or as nested sections of other
	// sections (like the below line)
	NestedSection := Section.MustRegisterSection("nested", &OtherComponentConfig{})

	// Override configpath to look for a custom location.
	configPath := filepath.Join("testdata", "nested_config.yaml")

	// Initialize an accessor.
	var accessor Accessor
	// e.g.
	// accessor = viper.NewAccessor(viper.Options{
	//		StrictMode:   true,
	//		SearchPaths: []string{configPath, configPath2},
	// })

	// Optionally bind to Pflags.
	// accessor.InitializePflags(flags)

	// Parse config from file or pass empty to rely on env variables and PFlags
	err := accessor.UpdateConfig(context.Background())
	if err != nil {
		fmt.Printf("Failed to validate config from [%v], error: %v", configPath, err)
		return
	}

	// Get parsed config value.
	parsedConfig := NestedSection.GetConfig().(*OtherComponentConfig)
	fmt.Printf("Config: %v", parsedConfig)
}
