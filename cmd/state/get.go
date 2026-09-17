package state

import "github.com/spf13/cobra"

func newGetCommand() *cobra.Command {
	command := &cobra.Command{
		Use:  "get state_id",
		Args: cobra.ExactArgs(1),
		RunE: runGet,
	}
	addContextFlags(command)
	return command
}

func runGet(command *cobra.Command, args []string) error {
	route, err := routeContext(command)
	if err != nil {
		return err
	}
	client, err := stateClient()
	if err != nil {
		return err
	}
	state, _, err := client.Get(command.Context(), route.Workspace, route.ProjectID, args[0])
	if err != nil {
		return err
	}
	return outputState(state)
}
