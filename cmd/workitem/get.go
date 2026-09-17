package workitem

import "github.com/spf13/cobra"

func newGetCommand(legacy bool) *cobra.Command {
	command := &cobra.Command{Use: "get work_item_id", Args: cobra.ExactArgs(1), RunE: runGet}
	addContextFlags(command)
	addDetailFlags(command)
	if legacy {
		command.RunE = runLegacyGet
	}
	return command
}

func runGet(cmd *cobra.Command, args []string) error {
	options, err := detailOptions(cmd)
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
	item, _, err := client.Get(cmd.Context(), route.Workspace, route.ProjectID, args[0], options)
	if err != nil {
		return err
	}
	return outputWorkItem(item)
}

func runLegacyGet(cmd *cobra.Command, args []string) error {
	options, err := detailOptions(cmd)
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
	item, _, err := client.LegacyGet(cmd.Context(), route.Workspace, route.ProjectID, args[0], options)
	if err != nil {
		return err
	}
	return outputWorkItem(item)
}
