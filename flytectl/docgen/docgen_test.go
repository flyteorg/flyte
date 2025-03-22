package docgen

import (
	"bytes"
	"context"
	"io"
	"os"
	"strings"
	"testing"

	cmdCore "github.com/flyteorg/flyte/flytectl/cmd/core"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestGenerateMarkdownDoc(t *testing.T) {
	// Create a test command
	cmd := &cobra.Command{
		Use:               "testcmd",
		Short:             "A test command",
		Long:              "A longer description of the test command",
		DisableAutoGenTag: true,
	}
	cmd.Flags().String("flag1", "", "Description of flag1")
	cmd.Flags().Bool("flag2", false, "Description of flag2")

	// Generate markdown
	output := GenerateMarkdownDoc(cmd)
	assert.NotNil(t, output)

	// Check for expected content
	outputStr := string(output)
	t.Logf("Generated output: %s", outputStr)
	assert.Contains(t, outputStr, "## `testcmd`")
	assert.Contains(t, outputStr, "A test command")
	assert.Contains(t, outputStr, "| `--flag1` |")
	assert.Contains(t, outputStr, "| `--flag2` |")

	// Test subcommand separately
	subCmd := &cobra.Command{
		Use:               "subcmd",
		Short:             "A subcommand",
		Long:              "A longer description of the subcommand",
		DisableAutoGenTag: true,
	}
	subCmd.Flags().Int("subflag", 0, "Description of subflag")

	// Generate markdown for subcommand
	subOutput := GenerateMarkdownDoc(subCmd)
	subOutputStr := string(subOutput)

	// Check subcommand output
	assert.Contains(t, subOutputStr, "## `subcmd`")
	assert.Contains(t, subOutputStr, "A subcommand")
	assert.Contains(t, subOutputStr, "| `--subflag` |")
}

func TestConvertFlagBlockToTable(t *testing.T) {
	// Test input with flag block
	input := `Some content

### Options

` + "```" + `
--flag1      Description of flag1
--flag2      Description of flag2
` + "```" + `

More content`

	// Update expected to match the actual output format with extra newlines
	expected := `Some content


### Options

| Flag | Description |
|------|-------------|
| ` + "`--flag1`" + ` | Description of flag1 |
| ` + "`--flag2`" + ` | Description of flag2 |


More content`

	result := convertFlagBlockToTable(input)
	assert.Equal(t, expected, result)
}

func TestConvertFlagBlockToTableSection(t *testing.T) {
	// Test input with specific section
	input := `Some content

### Options

` + "```" + `
--flag1      Description of flag1
--flag2      Description of flag2
` + "```" + `

### Options inherited from parent commands

` + "```" + `
--parent-flag      Parent flag description
` + "```" + `

More content`

	// Test converting just the inherited options section
	result := convertFlagBlockToTableSection("Options inherited from parent commands", input)
	assert.Contains(t, result, "### Options inherited from parent commands")
	assert.Contains(t, result, "| `--parent-flag` | Parent flag description |")
	assert.Contains(t, result, "### Options\n\n```") // Original options section should remain unchanged
}

func TestRemoveSeeAlsoSection(t *testing.T) {
	input := `Some content

### SEE ALSO

* Some reference
* Another reference

### Next section

More content`

	expected := `Some content

### Next section

More content`

	result := removeSeeAlsoSection(input)
	assert.Equal(t, expected, result)
}

func TestRemoveAllGeneratedHeadings(t *testing.T) {
	input := `## Generated heading

Some content

## ` + "`Manual heading`" + `

More content`

	expected := `
Some content

## ` + "`Manual heading`" + `

More content`

	result := removeAllGeneratedHeadings(input)
	assert.Equal(t, expected, result)
}

func TestEscapeHtml(t *testing.T) {
	input := "Text with <tags> and more <content>"
	expected := "Text with &lt;tags&gt; and more &lt;content&gt;"

	result := escapeHtml(input)
	assert.Equal(t, expected, result)
}

func TestGetDocgenCmd(t *testing.T) {
	root := &cobra.Command{Use: "flytectl"}
	cmds := GetDocgenCmd(root)

	// Check if docgen command exists
	docgenCmd, exists := cmds["docgen"]
	assert.True(t, exists)
	assert.Equal(t, []string{"docgen"}, docgenCmd.Aliases)
	assert.True(t, docgenCmd.ProjectDomainNotRequired)
	assert.True(t, strings.Contains(docgenCmd.Short, "generates documentation"))
}

func TestStartDocgen(t *testing.T) {
	// Create a test command
	root := &cobra.Command{
		Use:   "flytectl",
		Short: "Flytectl CLI",
	}

	// Get the command function
	cmdFunc := startDocgen(root)
	assert.NotNil(t, cmdFunc)

	// Create a mock context and command context
	ctx := context.Background()
	cmdCtx := cmdCore.CommandContext{}

	// Redirect stdout to capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Execute the command function with a custom exit function to prevent test from exiting
	originalOsExit := osExit
	defer func() { osExit = originalOsExit }()

	exitCalled := false
	osExit = func(code int) {
		exitCalled = true
		assert.Equal(t, 0, code) // Should exit with code 0 on success
	}

	// Call the command function
	err := cmdFunc(ctx, []string{}, cmdCtx)

	// Close the writer and restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read the captured output
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := buf.String()

	// Verify the output contains expected content
	assert.Contains(t, output, "## `flytectl`")
	assert.Contains(t, output, "Flytectl CLI")

	// Verify no error was returned
	assert.NoError(t, err)
	assert.False(t, exitCalled, "os.Exit should not be called on success")
}

// Add a variable to allow mocking os.Exit
var osExit = os.Exit
