package config

import (
	"context"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

const (
	PathFlag        = "file"
	StrictModeFlag  = "strict"
	CommandValidate = "validate"
	CommandDiscover = "discover"
)

type AccessorProvider func(options Options) Accessor

type printer interface {
	Printf(format string, i ...interface{})
	Println(i ...interface{})
}

func NewConfigCommand(accessorProvider AccessorProvider) *cobra.Command {
	opts := Options{}
	rootCmd := &cobra.Command{
		Use:       "config",
		Short:     "Runs various config commands, look at the help of this command to get a list of available commands..",
		ValidArgs: []string{CommandValidate, CommandDiscover},
	}

	validateCmd := &cobra.Command{
		Use:   "validate",
		Short: "Validates the loaded config.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return validate(accessorProvider(opts), cmd)
		},
	}

	discoverCmd := &cobra.Command{
		Use:   "discover",
		Short: "Searches for a config in one of the default search paths.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return validate(accessorProvider(opts), cmd)
		},
	}

	// Configure Root Command
	rootCmd.PersistentFlags().StringArrayVar(&opts.SearchPaths, PathFlag, []string{}, `Passes the config file to load.
If empty, it'll first search for the config file path then, if found, will load config from there.`)

	rootCmd.AddCommand(validateCmd)
	rootCmd.AddCommand(discoverCmd)

	// Configure Validate Command
	validateCmd.Flags().BoolVar(&opts.StrictMode, StrictModeFlag, false, `Validates that all keys in loaded config
map to already registered sections.`)

	return rootCmd
}

// Redirects Stdout to a string buffer until context is cancelled.
func redirectStdOut() (old, new *os.File) {
	old = os.Stdout // keep backup of the real stdout
	var err error
	_, new, err = os.Pipe()
	if err != nil {
		panic(err)
	}

	os.Stdout = new

	return
}

func validate(accessor Accessor, p printer) error {
	// Redirect stdout
	old, n := redirectStdOut()
	defer func() {
		err := n.Close()
		if err != nil {
			panic(err)
		}
	}()
	defer func() { os.Stdout = old }()

	err := accessor.UpdateConfig(context.Background())

	printInfo(p, accessor)
	if err == nil {
		green := color.New(color.FgGreen).SprintFunc()
		p.Println(green("Validated config file successfully."))
	} else {
		red := color.New(color.FgRed).SprintFunc()
		p.Println(red("Failed to validate config file."))
	}

	return err
}

func printInfo(p printer, v Accessor) {
	cfgFile := v.ConfigFilesUsed()
	if len(cfgFile) != 0 {
		green := color.New(color.FgGreen).SprintFunc()

		p.Printf("Config file(s) found at: %v\n", green(strings.Join(cfgFile, "\n")))
	} else {
		red := color.New(color.FgRed).SprintFunc()
		p.Println(red("Couldn't find a config file."))
	}
}
