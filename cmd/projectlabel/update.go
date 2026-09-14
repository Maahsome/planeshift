package projectlabel

import (
	projectlabelresource "planeshift/projectlabels"

	"github.com/spf13/cobra"
)

func newUpdateCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "update workspace_slug label_id",
		Short: "Update a project label",
		Args:  cobra.ExactArgs(2),
		RunE:  runUpdate,
	}
	command.Flags().String("name", "", "Project label name")
	command.Flags().String("description", "", "Project label description")
	command.Flags().String("color", "", "Project label color")
	command.Flags().Float64("sort-order", 0, "Project label sort order")
	return command
}

func runUpdate(cmd *cobra.Command, args []string) error {
	name, err := optionalStringFlag(cmd, "name")
	if err != nil {
		return err
	}
	description, err := optionalStringFlag(cmd, "description")
	if err != nil {
		return err
	}
	color, err := optionalStringFlag(cmd, "color")
	if err != nil {
		return err
	}
	sortOrder, err := optionalFloat64Flag(cmd, "sort-order")
	if err != nil {
		return err
	}
	client, err := projectLabelClient()
	if err != nil {
		return err
	}
	label, _, err := client.Update(cmd.Context(), args[0], args[1], projectlabelresource.UpdateProjectLabelRequest{
		Name: name, Description: description, Color: color, SortOrder: sortOrder,
	})
	if err != nil {
		return err
	}
	return outputProjectLabel(label)
}
