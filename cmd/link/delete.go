package link

import "github.com/spf13/cobra"

func newDeleteCommand(legacy bool) *cobra.Command {
	parent := "work_item_id"
	if legacy {
		parent = "issue_id"
	}
	command := &cobra.Command{Use: "delete " + parent + " link_id", Args: cobra.ExactArgs(2), RunE: runDelete}
	addContextFlags(command)
	if legacy {
		command.RunE = runLegacyDelete
	}
	return command
}

func runDelete(command *cobra.Command, args []string) error {
	route, err := routeContext(command)
	if err != nil {
		return err
	}
	client, err := linkClient()
	if err != nil {
		return err
	}
	_, err = client.Delete(command.Context(), route.Workspace, route.ProjectID, args[0], args[1])
	return err
}

func runLegacyDelete(command *cobra.Command, args []string) error {
	route, err := routeContext(command)
	if err != nil {
		return err
	}
	client, err := linkClient()
	if err != nil {
		return err
	}
	_, err = client.LegacyDelete(command.Context(), route.Workspace, route.ProjectID, args[0], args[1])
	return err
}
