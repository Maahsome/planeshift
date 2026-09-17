package workitem

import "github.com/spf13/cobra"

func newDeleteCommand(legacy bool) *cobra.Command {
	command := &cobra.Command{Use: "delete work_item_id", Args: cobra.ExactArgs(1), RunE: runDelete}
	addContextFlags(command)
	if legacy {
		command.RunE = runLegacyDelete
	}
	return command
}

func runDelete(cmd *cobra.Command, args []string) error {
	route, err := routeContext(cmd, true)
	if err != nil {
		return err
	}
	client, err := workItemClient()
	if err != nil {
		return err
	}
	_, err = client.Delete(cmd.Context(), route.Workspace, route.ProjectID, args[0])
	return err
}

func runLegacyDelete(cmd *cobra.Command, args []string) error {
	route, err := routeContext(cmd, true)
	if err != nil {
		return err
	}
	client, err := workItemClient()
	if err != nil {
		return err
	}
	_, err = client.LegacyDelete(cmd.Context(), route.Workspace, route.ProjectID, args[0])
	return err
}
