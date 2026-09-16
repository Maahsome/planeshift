package workitem

import "github.com/spf13/cobra"

func newUpdateCommand(legacy bool) *cobra.Command {
	command := &cobra.Command{Use: "update workspace_slug project_id work_item_id", Args: cobra.ExactArgs(3), RunE: runUpdate}
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
	client, err := workItemClient()
	if err != nil {
		return err
	}
	item, _, err := client.Update(cmd.Context(), args[0], args[1], args[2], request)
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
	client, err := workItemClient()
	if err != nil {
		return err
	}
	item, _, err := client.LegacyUpdate(cmd.Context(), args[0], args[1], args[2], request)
	if err != nil {
		return err
	}
	return outputWorkItem(item)
}
