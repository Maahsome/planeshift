package label

import (
	"fmt"

	labelResource "planeshift/labels"

	"github.com/spf13/cobra"
)

func newListCommand() *cobra.Command {
	command := &cobra.Command{
		Use:  "list",
		Args: cobra.NoArgs,
		RunE: runList,
	}
	addContextFlags(command)
	command.Flags().String("cursor", "", "Cursor for the next or previous label page")
	command.Flags().Int("per-page", 0, "Labels per page (Plane default 20; valid range 1-100)")
	command.Flags().String("fields", "", "Comma-separated label fields to return")
	command.Flags().String("expand", "", "Comma-separated related fields to expand")
	command.Flags().String("order-by", "", "Label ordering field; prefix with - for descending")
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
	orderBy, err := command.Flags().GetString("order-by")
	if err != nil {
		return err
	}
	if command.Flags().Changed("per-page") && perPage == 0 {
		return fmt.Errorf("--per-page must be between 1 and 100")
	}
	options := labelResource.ListOptions{
		Cursor: cursor, PerPage: perPage, Fields: fields, Expand: expand, OrderBy: orderBy,
	}
	if _, err := options.Query(); err != nil {
		return err
	}
	route, err := routeContext(command)
	if err != nil {
		return err
	}
	client, err := labelClient()
	if err != nil {
		return err
	}
	page, _, err := client.List(command.Context(), route.Workspace, route.ProjectID, options)
	if err != nil {
		return err
	}
	return outputLabel(page)
}
