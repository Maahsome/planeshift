package project

import "github.com/spf13/cobra"

func newGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:  "get workspace_slug project_id",
		Args: cobra.ExactArgs(2),
		RunE: runGet,
	}
}

func runGet(cmd *cobra.Command, args []string) error {
	client, err := projectClient()
	if err != nil {
		return err
	}
	project, _, err := client.Get(cmd.Context(), args[0], args[1])
	if err != nil {
		return err
	}
	return outputProject(project)
}
