package workitem

import "github.com/spf13/cobra"

func newListCommand(legacy bool) *cobra.Command {
	command := &cobra.Command{Use: "list", Args: cobra.NoArgs, RunE: runList}
	addContextFlags(command)
	addListFlags(command)
	if legacy {
		command.RunE = runLegacyList
	}
	return command
}

func runList(cmd *cobra.Command, args []string) error {
	options, err := listOptions(cmd)
	if err != nil {
		return err
	}
	route, err := routeContext(cmd, true)
	if err != nil {
		return err
	}
	client, err := workItemClient()
	if err != nil {
		return err
	}
	page, _, err := client.List(cmd.Context(), route.Workspace, route.ProjectID, options)
	if err != nil {
		return err
	}
	return outputWorkItem(page)
}

func runLegacyList(cmd *cobra.Command, args []string) error {
	options, err := listOptions(cmd)
	if err != nil {
		return err
	}
	route, err := routeContext(cmd, true)
	if err != nil {
		return err
	}
	client, err := workItemClient()
	if err != nil {
		return err
	}
	page, _, err := client.LegacyList(cmd.Context(), route.Workspace, route.ProjectID, options)
	if err != nil {
		return err
	}
	return outputWorkItem(page)
}
