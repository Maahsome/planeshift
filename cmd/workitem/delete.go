package workitem

import "github.com/spf13/cobra"

func newDeleteCommand(legacy bool) *cobra.Command {
	command := &cobra.Command{Use: "delete workspace_slug project_id work_item_id", Args: cobra.ExactArgs(3), RunE: runDelete}
	if legacy {
		command.RunE = runLegacyDelete
	}
	return command
}

func runDelete(cmd *cobra.Command, args []string) error {
	client, err := workItemClient()
	if err != nil {
		return err
	}
	_, err = client.Delete(cmd.Context(), args[0], args[1], args[2])
	return err
}

func runLegacyDelete(cmd *cobra.Command, args []string) error {
	client, err := workItemClient()
	if err != nil {
		return err
	}
	_, err = client.LegacyDelete(cmd.Context(), args[0], args[1], args[2])
	return err
}
