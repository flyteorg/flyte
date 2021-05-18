package cmd

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/flyteorg/flytestdlib/cli/pflags/api"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/spf13/cobra"
)

var (
	pkg                       string
	defaultValuesVariable     string
	shouldBindDefaultVariable bool
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
	root.Flags().StringVarP(&pkg, "package", "p", ".", "Determines the source/destination package.")
	root.Flags().StringVar(&defaultValuesVariable, "default-var", "defaultConfig", "Points to a variable to use to load default configs. If specified & found, it'll be used instead of the values specified in the tag.")
	root.Flags().BoolVar(&shouldBindDefaultVariable, "bind-default-var", false, "The generated PFlags Set will bind fields to the default variable.")
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
	gen, err := api.NewGenerator(pkg, structName, defaultValuesVariable, shouldBindDefaultVariable)
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
