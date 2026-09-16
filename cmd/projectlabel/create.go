package projectlabel

import (
	projectlabelresource "planeshift/projectlabels"

	"github.com/spf13/cobra"
)

func newCreateCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "create workspace_slug",
		Short: "Create a project label",
		Args:  cobra.ExactArgs(1),
		RunE:  runCreate,
	}
	command.Flags().String("name", "", "Project label name")
	command.Flags().String("description", "", "Project label description")
	command.Flags().String("color", "", "Project label color")
	command.Flags().Float64("sort-order", 0, "Project label sort order")
	_ = command.MarkFlagRequired("name")
	return command
}

func runCreate(cmd *cobra.Command, args []string) error {
	name, err := cmd.Flags().GetString("name")
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
	label, _, err := client.Create(cmd.Context(), args[0], projectlabelresource.CreateProjectLabelRequest{
		Name: name, Description: description, Color: color, SortOrder: sortOrder,
	})
	if err != nil {
		return err
	}
	return outputProjectLabel(label)
}
