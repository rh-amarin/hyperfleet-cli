package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:          "completion [bash|zsh|fish|powershell]",
	Short:        "Generate shell completion scripts",
	SilenceUsage: true,
	ValidArgs:    []string{"bash", "zsh", "fish", "powershell"},
	Args:         cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		w := cmd.OutOrStdout()
		switch args[0] {
		case "bash":
			return cmd.Root().GenBashCompletion(w)
		case "zsh":
			return cmd.Root().GenZshCompletion(w)
		case "fish":
			return cmd.Root().GenFishCompletion(w, true)
		case "powershell":
			return cmd.Root().GenPowerShellCompletionWithDesc(w)
		default:
			return fmt.Errorf("unknown shell: %s (supported: bash, zsh, fish, powershell)", args[0])
		}
	},
}

func init() {
	rootCmd.AddCommand(completionCmd)
}
