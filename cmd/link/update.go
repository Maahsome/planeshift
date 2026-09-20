package link

import (
	resource "planeshift/links"

	"github.com/spf13/cobra"
)

func newUpdateCommand(legacy bool) *cobra.Command {
	parent := "work_item_id"
	if legacy {
		parent = "issue_id"
	}
	command := &cobra.Command{Use: "update " + parent + " link_id", Args: cobra.ExactArgs(2), RunE: runUpdate}
	addContextFlags(command)
	command.Flags().String("url", "", "External link URL")
	command.Flags().String("title", "", "Link title")
	if legacy {
		command.RunE = runLegacyUpdate
	}
	return command
}

func runUpdate(command *cobra.Command, args []string) error {
	request, err := updateRequest(command)
	if err != nil {
		return err
	}
	route, err := routeContext(command)
	if err != nil {
		return err
	}
	client, err := linkClient()
	if err != nil {
		return err
	}
	linkValue, _, err := client.Update(command.Context(), route.Workspace, route.ProjectID, args[0], args[1], request)
	if err != nil {
		return err
	}
	return outputLink(linkValue)
}

func runLegacyUpdate(command *cobra.Command, args []string) error {
	request, err := updateRequest(command)
	if err != nil {
		return err
	}
	route, err := routeContext(command)
	if err != nil {
		return err
	}
	client, err := linkClient()
	if err != nil {
		return err
	}
	linkValue, _, err := client.LegacyUpdate(command.Context(), route.Workspace, route.ProjectID, args[0], args[1], request)
	if err != nil {
		return err
	}
	return outputLink(linkValue)
}

func updateRequest(command *cobra.Command) (resource.UpdateLinkRequest, error) {
	urlValue, err := optionalStringFlag(command, "url")
	if err != nil {
		return resource.UpdateLinkRequest{}, err
	}
	title, err := optionalStringFlag(command, "title")
	if err != nil {
		return resource.UpdateLinkRequest{}, err
	}
	return resource.UpdateLinkRequest{URL: urlValue, Title: title}, nil
}
