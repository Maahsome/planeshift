package project

import "github.com/spf13/cobra"

func newArchiveCommand() *cobra.Command {
	return &cobra.Command{
		Use:  "archive workspace_slug project_id",
		Args: cobra.ExactArgs(2),
		RunE: runArchive,
	}
}

func runArchive(cmd *cobra.Command, args []string) error {
	client, err := projectClient()
	if err != nil {
		return err
	}
	_, err = client.Archive(cmd.Context(), args[0], args[1])
	return err
}
