package state

import (
	"fmt"

	stateResource "planeshift/states"

	"github.com/spf13/cobra"
)

func newListCommand() *cobra.Command {
	command := &cobra.Command{
		Use:  "list workspace_slug project_id",
		Args: cobra.ExactArgs(2),
		RunE: runList,
	}
	command.Flags().String("cursor", "", "Cursor for the next or previous state page")
	command.Flags().Int("per-page", 0, "States per page (Plane default 20; valid range 1-100)")
	command.Flags().String("fields", "", "Comma-separated state fields to return")
	command.Flags().String("expand", "", "Comma-separated related fields to expand")
	return command
}

func runList(command *cobra.Command, args []string) error {
	cursor, err := command.Flags().GetString("cursor")
	if err != nil {
		return err
	}
	perPage, err := command.Flags().GetInt("per-page")
	if err != nil {
		return err
	}
	fields, err := command.Flags().GetString("fields")
	if err != nil {
		return err
	}
	expand, err := command.Flags().GetString("expand")
	if err != nil {
		return err
	}
	if command.Flags().Changed("per-page") && perPage == 0 {
		return fmt.Errorf("--per-page must be between 1 and 100")
	}
	options := stateResource.ListOptions{Cursor: cursor, PerPage: perPage, Fields: fields, Expand: expand}
	if _, err := options.Query(); err != nil {
		return err
	}
	client, err := stateClient()
	if err != nil {
		return err
	}
	page, _, err := client.List(command.Context(), args[0], args[1], options)
	if err != nil {
		return err
	}
	return outputState(page)
}
