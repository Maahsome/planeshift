package project

import "github.com/spf13/cobra"

func newGetCommand() *cobra.Command {
	command := &cobra.Command{
		Use:  "get",
		Args: cobra.NoArgs,
		RunE: runGet,
	}
	addContextFlags(command, true)
	return command
}

func runGet(cmd *cobra.Command, args []string) error {
	route, err := routeContext(cmd, true)
	if err != nil {
		return err
	}
	client, err := projectClient()
	if err != nil {
		return err
	}
	project, _, err := client.Get(cmd.Context(), route.Workspace, route.ProjectID)
	if err != nil {
		return err
	}
	return outputProject(project)
}
