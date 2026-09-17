package workitem

import "github.com/spf13/cobra"

func newGetCommand(legacy bool) *cobra.Command {
	command := &cobra.Command{Use: "get workspace_slug project_id work_item_id", Args: cobra.ExactArgs(3), RunE: runGet}
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
	client, err := workItemClient()
	if err != nil {
		return err
	}
	item, _, err := client.Get(cmd.Context(), args[0], args[1], args[2], options)
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
	client, err := workItemClient()
	if err != nil {
		return err
	}
	item, _, err := client.LegacyGet(cmd.Context(), args[0], args[1], args[2], options)
	if err != nil {
		return err
	}
	return outputWorkItem(item)
}
