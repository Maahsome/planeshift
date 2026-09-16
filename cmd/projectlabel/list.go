package projectlabel

import (
	"fmt"

	projectlabelresource "planeshift/projectlabels"

	"github.com/spf13/cobra"
)

func newListCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "list workspace_slug",
		Short: "List workspace project labels",
		Args:  cobra.ExactArgs(1),
		RunE:  runList,
	}
	command.Flags().String("cursor", "", "Cursor for the next or previous project-label page")
	command.Flags().Int("per-page", 0, "Project labels per page (Plane default 20; valid range 1-100)")
	command.Flags().String("fields", "", "Comma-separated project-label fields to return")
	command.Flags().String("expand", "", "Comma-separated related fields to expand")
	command.Flags().String("order-by", "", "Project-label ordering field; prefix with - for descending")
	return command
}

func runList(cmd *cobra.Command, args []string) error {
	cursor, err := cmd.Flags().GetString("cursor")
	if err != nil {
		return err
	}
	perPage, err := cmd.Flags().GetInt("per-page")
	if err != nil {
		return err
	}
	fields, err := cmd.Flags().GetString("fields")
	if err != nil {
		return err
	}
	expand, err := cmd.Flags().GetString("expand")
	if err != nil {
		return err
	}
	orderBy, err := cmd.Flags().GetString("order-by")
	if err != nil {
		return err
	}
	if cmd.Flags().Changed("per-page") && perPage == 0 {
		return fmt.Errorf("--per-page must be between 1 and 100")
	}
	options := projectlabelresource.ListOptions{
		Cursor: cursor, PerPage: perPage, Fields: fields, Expand: expand, OrderBy: orderBy,
	}
	if _, err := options.Query(); err != nil {
		return err
	}
	client, err := projectLabelClient()
	if err != nil {
		return err
	}
	page, _, err := client.List(cmd.Context(), args[0], options)
	if err != nil {
		return err
	}
	return outputProjectLabel(page)
}
