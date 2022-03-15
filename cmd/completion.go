/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// completionCmd represents the completion command
var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generates completion script.",
	Long: `To load completion, run the following commands in accordance with the shell you are using:

- Bash
	::

	$ source <(flytectl completion bash)

	To load completions for each session:

	- Linux
	::

	$ flytectl completion bash > /etc/bash_completion.d/flytectl

	- macOS
	::

	$ flytectl completion bash > /usr/local/etc/bash_completion.d/flytectl

- Zsh
	If shell completion is not already enabled in your environment, enable it:

	::

	$ echo "autoload -U compinit; compinit" >> ~/.zshrc

	Once enabled, execute once:

	::

	$ flytectl completion zsh > "${fpath[1]}/_flytectl"

	.. note::
		Start a new shell for this setup to take effect.

- fish
	::

	$ flytectl completion fish | source

	To load completions for each session, run:

	::

	$ flytectl completion fish > ~/.config/fish/completions/flytectl.fish

- PowerShell
	::

	 PS> flytectl completion powershell | Out-String | Invoke-Expression

	To load completions for each session, run:

	::

	 PS> flytectl completion powershell > flytectl.ps1

	and source this file from your PowerShell profile.
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.ExactValidArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		switch args[0] {
		case "bash":
			return cmd.Root().GenBashCompletion(os.Stdout)
		case "zsh":
			return cmd.Root().GenZshCompletion(os.Stdout)
		case "fish":
			return cmd.Root().GenFishCompletion(os.Stdout, true)
		case "powershell":
			return cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
		}
		return nil
	},
}
