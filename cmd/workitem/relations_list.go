package workitem

import "github.com/spf13/cobra"

func newRelationsListCommand() *cobra.Command {
	command := &cobra.Command{Use: "relations-list workspace_slug project_id work_item_id", Args: cobra.ExactArgs(3), RunE: runRelationsList}
	addRelationListFlags(command)
	return command
}

func runRelationsList(cmd *cobra.Command, args []string) error {
	options, err := relationListOptions(cmd)
	if err != nil {
		return err
	}
	client, err := workItemClient()
	if err != nil {
		return err
	}
	page, _, err := client.ListRelations(cmd.Context(), args[0], args[1], args[2], options)
	if err != nil {
		return err
	}
	return outputWorkItem(page)
}
