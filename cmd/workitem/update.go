package workitem

import "github.com/spf13/cobra"

func newUpdateCommand(legacy bool) *cobra.Command {
	command := &cobra.Command{Use: "update work_item_id", Args: cobra.ExactArgs(1), RunE: runUpdate}
	addContextFlags(command)
	addRequestFlags(command, true)
	if legacy {
		command.RunE = runLegacyUpdate
	}
	return command
}

func runUpdate(cmd *cobra.Command, args []string) error {
	request, err := updateRequestValues(cmd)
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
	item, _, err := client.Update(cmd.Context(), route.Workspace, route.ProjectID, args[0], request)
	if err != nil {
		return err
	}
	return outputWorkItem(item)
}

func runLegacyUpdate(cmd *cobra.Command, args []string) error {
	request, err := updateRequestValues(cmd)
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
	item, _, err := client.LegacyUpdate(cmd.Context(), route.Workspace, route.ProjectID, args[0], request)
	if err != nil {
		return err
	}
	return outputWorkItem(item)
}
