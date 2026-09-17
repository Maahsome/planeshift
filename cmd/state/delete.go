package state

import "github.com/spf13/cobra"

func newDeleteCommand() *cobra.Command {
	command := &cobra.Command{
		Use:  "delete state_id",
		Args: cobra.ExactArgs(1),
		RunE: runDelete,
	}
	addContextFlags(command)
	return command
}

func runDelete(command *cobra.Command, args []string) error {
	route, err := routeContext(command)
	if err != nil {
		return err
	}
	client, err := stateClient()
	if err != nil {
		return err
	}
	_, err = client.Delete(command.Context(), route.Workspace, route.ProjectID, args[0])
	return err
}
