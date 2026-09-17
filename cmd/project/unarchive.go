package project

import "github.com/spf13/cobra"

func newUnarchiveCommand() *cobra.Command {
	command := &cobra.Command{
		Use:  "unarchive",
		Args: cobra.NoArgs,
		RunE: runUnarchive,
	}
	addContextFlags(command, true)
	return command
}

func runUnarchive(cmd *cobra.Command, args []string) error {
	route, err := routeContext(cmd, true)
	if err != nil {
		return err
	}
	client, err := projectClient()
	if err != nil {
		return err
	}
	_, err = client.Unarchive(cmd.Context(), route.Workspace, route.ProjectID)
	return err
}
