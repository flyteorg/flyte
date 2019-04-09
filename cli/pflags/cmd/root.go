package cmd

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"strings"

	"github.com/lyft/flytestdlib/cli/pflags/api"
	"github.com/lyft/flytestdlib/logger"
	"github.com/spf13/cobra"
)

var (
	pkg = flag.String("pkg", ".", "what package to get the interface from")
)

var root = cobra.Command{
	Use:  "pflags MyStructName --package myproject/mypackage",
	Args: cobra.ExactArgs(1),
	RunE: generatePflagsProvider,
	Example: `
// go:generate pflags MyStruct
type MyStruct struct {
	BoolValue        bool              ` + "`json:\"bl\" pflag:\"true\"`" + `
	NestedType       NestedType        ` + "`json:\"nested\"`" + `
	IntArray         []int             ` + "`json:\"ints\" pflag:\"[]int{12%2C1}\"`" + `
}
	`,
}

func init() {
	root.Flags().StringP("package", "p", ".", "Determines the source/destination package.")
}

func Execute() error {
	return root.Execute()
}

func generatePflagsProvider(cmd *cobra.Command, args []string) error {
	structName := args[0]
	if structName == "" {
		return fmt.Errorf("need to specify a struct name")
	}

	ctx := context.Background()
	gen, err := api.NewGenerator(*pkg, structName)
	if err != nil {
		return err
	}

	provider, err := gen.Generate(ctx)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	defer buf.Reset()

	logger.Infof(ctx, "Generating PFlags for type [%v.%v.%v]\n", gen.GetTargetPackage().Path(), gen.GetTargetPackage().Name(), structName)

	outFilePath := fmt.Sprintf("%s_flags.go", strings.ToLower(structName))
	err = provider.WriteCodeFile(outFilePath)
	if err != nil {
		return err
	}

	tOutFilePath := fmt.Sprintf("%s_flags_test.go", strings.ToLower(structName))
	return provider.WriteTestFile(tOutFilePath)
}
