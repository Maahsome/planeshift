package link

import "github.com/spf13/cobra"

func newLegacyCommand() *cobra.Command {
	command := &cobra.Command{
		Use:    "legacy",
		Hidden: true,
		Args:   cobra.NoArgs,
	}
	command.AddCommand(
		newCreateCommand(true),
		newListCommand(true),
		newGetCommand(true),
		newUpdateCommand(true),
		newDeleteCommand(true),
	)
	return command
}
