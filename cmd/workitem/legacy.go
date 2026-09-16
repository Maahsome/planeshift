package workitem

import "github.com/spf13/cobra"

func newLegacyCommand() *cobra.Command {
	command := &cobra.Command{
		Use:    "legacy",
		Short:  "Deprecated /issues/ compatibility operations",
		Long:   "Compatibility-only access to the seven inventoried legacy /issues/ core routes. Current /work-items/ routes remain primary; relation commands are intentionally unavailable here.",
		Hidden: true,
		Args:   cobra.NoArgs,
	}
	command.AddCommand(
		newSearchCommand(true),
		newGetByIdentifierCommand(true),
		newListCommand(true),
		newCreateCommand(true),
		newGetCommand(true),
		newUpdateCommand(true),
		newDeleteCommand(true),
	)
	return command
}
