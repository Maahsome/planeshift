package projectlabel

import "github.com/spf13/cobra"

func newDeleteCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "delete workspace_slug label_id",
		Short: "Delete a project label",
		Args:  cobra.ExactArgs(2),
		RunE:  runDelete,
	}
}

func runDelete(cmd *cobra.Command, args []string) error {
	client, err := projectLabelClient()
	if err != nil {
		return err
	}
	_, err = client.Delete(cmd.Context(), args[0], args[1])
	return err
}
