package activity

import "github.com/spf13/cobra"

func newLegacyCommand() *cobra.Command {
	command := &cobra.Command{Use: "legacy", Hidden: true, Args: cobra.NoArgs}
	command.AddCommand(newListCommand(true), newGetCommand(true))
	return command
}
