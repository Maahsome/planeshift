package state

import "github.com/spf13/cobra"

func newGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:  "get workspace_slug project_id state_id",
		Args: cobra.ExactArgs(3),
		RunE: runGet,
	}
}

func runGet(command *cobra.Command, args []string) error {
	client, err := stateClient()
	if err != nil {
		return err
	}
	state, _, err := client.Get(command.Context(), args[0], args[1], args[2])
	if err != nil {
		return err
	}
	return outputState(state)
}
