package project

import "github.com/spf13/cobra"

func newUnarchiveCommand() *cobra.Command {
	return &cobra.Command{
		Use:  "unarchive workspace_slug project_id",
		Args: cobra.ExactArgs(2),
		RunE: runUnarchive,
	}
}

func runUnarchive(cmd *cobra.Command, args []string) error {
	client, err := projectClient()
	if err != nil {
		return err
	}
	_, err = client.Unarchive(cmd.Context(), args[0], args[1])
	return err
}
