package label

import (
	labelResource "planeshift/labels"

	"github.com/spf13/cobra"
)

func newUpdateCommand() *cobra.Command {
	command := &cobra.Command{
		Use:  "update label_id",
		Args: cobra.ExactArgs(1),
		RunE: runUpdate,
	}
	addContextFlags(command)
	command.Flags().String("name", "", "Label name")
	addLabelFields(command)
	return command
}

func runUpdate(command *cobra.Command, args []string) error {
	name, err := optionalStringFlag(command, "name")
	if err != nil {
		return err
	}
	color, err := optionalStringFlag(command, "color")
	if err != nil {
		return err
	}
	description, err := optionalStringFlag(command, "description")
	if err != nil {
		return err
	}
	externalSource, err := optionalStringFlag(command, "external-source")
	if err != nil {
		return err
	}
	externalID, err := optionalStringFlag(command, "external-id")
	if err != nil {
		return err
	}
	parent, err := optionalStringFlag(command, "parent")
	if err != nil {
		return err
	}
	sortOrder, err := optionalFloat64Flag(command, "sort-order")
	if err != nil {
		return err
	}
	request := labelResource.UpdateLabelRequest{
		Name: name, Color: color, Description: description,
		ExternalSource: externalSource, ExternalID: externalID,
		Parent: parent, SortOrder: sortOrder,
	}
	route, err := routeContext(command)
	if err != nil {
		return err
	}
	client, err := labelClient()
	if err != nil {
		return err
	}
	label, _, err := client.Update(command.Context(), route.Workspace, route.ProjectID, args[0], request)
	if err != nil {
		return err
	}
	return outputLabel(label)
}
