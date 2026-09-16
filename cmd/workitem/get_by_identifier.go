package workitem

import (
	resource "planeshift/workitems"

	"github.com/spf13/cobra"
)

func newGetByIdentifierCommand(legacy bool) *cobra.Command {
	command := &cobra.Command{Use: "get-by-identifier workspace_slug project_identifier issue_identifier", Args: cobra.ExactArgs(3), RunE: runGetByIdentifier}
	command.Flags().String("expand", "", "Comma-separated related fields to expand")
	if legacy {
		command.RunE = runLegacyGetByIdentifier
	}
	return command
}

func identifierOptions(cmd *cobra.Command) (resource.IdentifierOptions, error) {
	expand, err := cmd.Flags().GetString("expand")
	return resource.IdentifierOptions{Expand: expand}, err
}

func runGetByIdentifier(cmd *cobra.Command, args []string) error {
	options, err := identifierOptions(cmd)
	if err != nil {
		return err
	}
	client, err := workItemClient()
	if err != nil {
		return err
	}
	item, _, err := client.GetByIdentifier(cmd.Context(), args[0], args[1], args[2], options)
	if err != nil {
		return err
	}
	return outputWorkItem(item)
}

func runLegacyGetByIdentifier(cmd *cobra.Command, args []string) error {
	options, err := identifierOptions(cmd)
	if err != nil {
		return err
	}
	client, err := workItemClient()
	if err != nil {
		return err
	}
	item, _, err := client.LegacyGetByIdentifier(cmd.Context(), args[0], args[1], args[2], options)
	if err != nil {
		return err
	}
	return outputWorkItem(item)
}
