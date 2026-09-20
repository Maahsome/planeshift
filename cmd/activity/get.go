package activity

import "github.com/spf13/cobra"

func newGetCommand(legacy bool) *cobra.Command {
	identifier := "work_item_id"
	if legacy {
		identifier = "issue_id"
	}
	command := &cobra.Command{Use: "get " + identifier + " activity_id", Args: cobra.ExactArgs(2), RunE: runGet}
	addContextFlags(command)
	addDetailFlags(command)
	if legacy {
		command.RunE = runLegacyGet
	}
	return command
}

func runGet(command *cobra.Command, args []string) error {
	options, err := detailOptions(command)
	if err != nil {
		return err
	}
	route, err := routeContext(command)
	if err != nil {
		return err
	}
	client, err := activityClient()
	if err != nil {
		return err
	}
	result, _, err := client.Get(command.Context(), route.Workspace, route.ProjectID, args[0], args[1], options)
	if err != nil {
		return err
	}
	return outputActivity(result)
}

func runLegacyGet(command *cobra.Command, args []string) error {
	options, err := detailOptions(command)
	if err != nil {
		return err
	}
	route, err := routeContext(command)
	if err != nil {
		return err
	}
	client, err := activityClient()
	if err != nil {
		return err
	}
	result, _, err := client.LegacyGet(command.Context(), route.Workspace, route.ProjectID, args[0], args[1], options)
	if err != nil {
		return err
	}
	return outputActivity(result)
}
