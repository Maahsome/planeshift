package activity

import "github.com/spf13/cobra"

func newListCommand(legacy bool) *cobra.Command {
	identifier := "work_item_id"
	if legacy {
		identifier = "issue_id"
	}
	command := &cobra.Command{Use: "list " + identifier, Args: cobra.ExactArgs(1), RunE: runList}
	addContextFlags(command)
	addListFlags(command)
	if legacy {
		command.RunE = runLegacyList
	}
	return command
}

func runList(command *cobra.Command, args []string) error {
	options, err := listOptions(command)
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
	page, _, err := client.List(command.Context(), route.Workspace, route.ProjectID, args[0], options)
	if err != nil {
		return err
	}
	return outputActivity(page)
}

func runLegacyList(command *cobra.Command, args []string) error {
	options, err := listOptions(command)
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
	page, _, err := client.LegacyList(command.Context(), route.Workspace, route.ProjectID, args[0], options)
	if err != nil {
		return err
	}
	return outputActivity(page)
}
