package state

import "github.com/spf13/cobra"

func newDeleteCommand() *cobra.Command {
	return &cobra.Command{
		Use:  "delete workspace_slug project_id state_id",
		Args: cobra.ExactArgs(3),
		RunE: runDelete,
	}
}

func runDelete(command *cobra.Command, args []string) error {
	client, err := stateClient()
	if err != nil {
		return err
	}
	_, err = client.Delete(command.Context(), args[0], args[1], args[2])
	return err
}
