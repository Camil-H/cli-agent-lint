package cmd

import (
	"github.com/spf13/cobra"
)

func newCompletionCmd(opts *GlobalOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate shell completion scripts",
		Long: `Generate shell completion scripts for cli-agent-lint.

To load completions:

  bash:
    source <(cli-agent-lint completion bash)

  zsh:
    cli-agent-lint completion zsh > "${fpath[1]}/_cli-agent-lint"

  fish:
    cli-agent-lint completion fish | source

  powershell:
    cli-agent-lint completion powershell | Out-String | Invoke-Expression`,
		Args:      cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "bash":
				return cmd.Root().GenBashCompletion(opts.Out.DataWriter())
			case "zsh":
				return cmd.Root().GenZshCompletion(opts.Out.DataWriter())
			case "fish":
				return cmd.Root().GenFishCompletion(opts.Out.DataWriter(), true)
			case "powershell":
				return cmd.Root().GenPowerShellCompletionWithDesc(opts.Out.DataWriter())
			}
			return nil
		},
	}
	return cmd
}
