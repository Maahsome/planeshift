package project

import "github.com/spf13/cobra"

func newArchiveCommand() *cobra.Command {
	command := &cobra.Command{
		Use:  "archive",
		Args: cobra.NoArgs,
		RunE: runArchive,
	}
	addContextFlags(command, true)
	return command
}

func runArchive(cmd *cobra.Command, args []string) error {
	route, err := routeContext(cmd, true)
	if err != nil {
		return err
	}
	client, err := projectClient()
	if err != nil {
		return err
	}
	_, err = client.Archive(cmd.Context(), route.Workspace, route.ProjectID)
	return err
}
