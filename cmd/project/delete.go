package project

import "github.com/spf13/cobra"

func newDeleteCommand() *cobra.Command {
	command := &cobra.Command{
		Use:  "delete",
		Args: cobra.NoArgs,
		RunE: runDelete,
	}
	addContextFlags(command, true)
	return command
}

func runDelete(cmd *cobra.Command, args []string) error {
	route, err := routeContext(cmd, true)
	if err != nil {
		return err
	}
	client, err := projectClient()
	if err != nil {
		return err
	}
	_, err = client.Delete(cmd.Context(), route.Workspace, route.ProjectID)
	return err
}
