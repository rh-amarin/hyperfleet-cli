package cmd

import (
	"fmt"

	"github.com/rh-amarin/hyperfleet-cli/internal/plugin"
	"github.com/spf13/cobra"
)

func init() {
	// Give the root command a RunE so Cobra passes unrecognised subcommand
	// args here instead of returning an "unknown command" error itself.
	rootCmd.Args = cobra.ArbitraryArgs
	rootCmd.RunE = func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return cmd.Help()
		}
		name := args[0]
		if path, ok := plugin.Find(name); ok {
			return plugin.Exec(path, args[1:])
		}
		return fmt.Errorf("unknown command %q for %q\nRun 'hf --help' for usage.", name, cmd.Name())
	}
}
