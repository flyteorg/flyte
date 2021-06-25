.. _flytectl_completion:

flytectl completion
-------------------

Generate completion script

Synopsis
~~~~~~~~


To load completions:

Bash:

  $ source <(flytectl completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ flytectl completion bash > /etc/bash_completion.d/flytectl
  # macOS:
  $ flytectl completion bash > /usr/local/etc/bash_completion.d/flytectl

Zsh:

  # If shell completion is not already enabled in your environment,
  # you will need to enable it.  You can execute the following once:

  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ flytectl completion zsh > "${fpath[1]}/_flytectl"

  # You will need to start a new shell for this setup to take effect.

fish:

  $ flytectl completion fish | source

  # To load completions for each session, execute once:
  $ flytectl completion fish > ~/.config/fish/completions/flytectl.fish

PowerShell:

  PS> flytectl completion powershell | Out-String | Invoke-Expression

  # To load completions for every new session, run:
  PS> flytectl completion powershell > flytectl.ps1
  # and source this file from your PowerShell profile.


::

  flytectl completion [bash|zsh|fish|powershell]

Options
~~~~~~~

::

  -h, --help   help for completion

Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

  -c, --config string    config file (default is $HOME/.flyte/config.yaml)
  -d, --domain string    Specifies the Flyte project's domain.
  -o, --output string    Specifies the output type - supported formats [TABLE JSON YAML DOT DOTURL]. NOTE: dot, doturl are only supported for Workflow (default "TABLE")
  -p, --project string   Specifies the Flyte project.

SEE ALSO
~~~~~~~~

* :doc:`flytectl` 	 - flyetcl CLI tool

