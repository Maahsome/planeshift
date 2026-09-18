package link

import (
	resource "planeshift/links"

	"github.com/spf13/cobra"
)

func newCreateCommand(legacy bool) *cobra.Command {
	identifier := "work_item_id"
	if legacy {
		identifier = "issue_id"
	}
	command := &cobra.Command{Use: "create " + identifier, Args: cobra.ExactArgs(1), RunE: runCreate}
	addContextFlags(command)
	command.Flags().String("url", "", "External link URL")
	command.Flags().String("title", "", "Link title")
	_ = command.MarkFlagRequired("url")
	if legacy {
		command.RunE = runLegacyCreate
	}
	return command
}

func runCreate(command *cobra.Command, args []string) error {
	urlValue, err := command.Flags().GetString("url")
	if err != nil {
		return err
	}
	title, err := optionalStringFlag(command, "title")
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
	linkValue, _, err := client.Create(command.Context(), route.Workspace, route.ProjectID, args[0], resource.CreateLinkRequest{URL: urlValue, Title: title})
	if err != nil {
		return err
	}
	return outputLink(linkValue)
}

func runLegacyCreate(command *cobra.Command, args []string) error {
	urlValue, err := command.Flags().GetString("url")
	if err != nil {
		return err
	}
	title, err := optionalStringFlag(command, "title")
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
	linkValue, _, err := client.LegacyCreate(command.Context(), route.Workspace, route.ProjectID, args[0], resource.CreateLinkRequest{URL: urlValue, Title: title})
	if err != nil {
		return err
	}
	return outputLink(linkValue)
}
