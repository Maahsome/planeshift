package workitem

import "github.com/spf13/cobra"

func newCreateCommand(legacy bool) *cobra.Command {
	command := &cobra.Command{Use: "create", Args: cobra.NoArgs, RunE: runCreate}
	addContextFlags(command)
	addRequestFlags(command, false)
	if legacy {
		command.RunE = runLegacyCreate
	}
	return command
}

func runCreate(cmd *cobra.Command, args []string) error {
	request, err := requestValues(cmd)
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
	item, _, err := client.Create(cmd.Context(), route.Workspace, route.ProjectID, request)
	if err != nil {
		return err
	}
	return outputWorkItem(item)
}

func runLegacyCreate(cmd *cobra.Command, args []string) error {
	request, err := requestValues(cmd)
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
	item, _, err := client.LegacyCreate(cmd.Context(), route.Workspace, route.ProjectID, request)
	if err != nil {
		return err
	}
	return outputWorkItem(item)
}
