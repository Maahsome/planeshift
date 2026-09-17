package label

import "github.com/spf13/cobra"

func newGetCommand() *cobra.Command {
	command := &cobra.Command{
		Use:  "get label_id",
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
	client, err := labelClient()
	if err != nil {
		return err
	}
	label, _, err := client.Get(command.Context(), route.Workspace, route.ProjectID, args[0])
	if err != nil {
		return err
	}
	return outputLabel(label)
}
