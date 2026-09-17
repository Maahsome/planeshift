package workitem

import "github.com/spf13/cobra"

func newRelationsListCommand() *cobra.Command {
	command := &cobra.Command{Use: "relations-list work_item_id", Args: cobra.ExactArgs(1), RunE: runRelationsList}
	addContextFlags(command)
	addRelationListFlags(command)
	return command
}

func runRelationsList(cmd *cobra.Command, args []string) error {
	options, err := relationListOptions(cmd)
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
	page, _, err := client.ListRelations(cmd.Context(), route.Workspace, route.ProjectID, args[0], options)
	if err != nil {
		return err
	}
	return outputWorkItem(page)
}
