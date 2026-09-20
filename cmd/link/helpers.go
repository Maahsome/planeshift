package link

import (
	"fmt"

	"planeshift/config"
	resource "planeshift/links"
	"planeshift/objects"

	"github.com/spf13/cobra"
)

func linkClient() (*resource.Client, error) {
	if clientFactory == nil {
		return nil, fmt.Errorf("link client factory is not initialized")
	}
	client, err := clientFactory.New()
	if err != nil {
		return nil, err
	}
	return resource.NewClient(client), nil
}

func outputLink(value any) error {
	if c == nil {
		return fmt.Errorf("link command configuration is not initialized")
	}
	output, err := objects.NewRawJSONFromValue(value)
	if err != nil {
		return err
	}
	if c.OutputFormat == "" {
		c.OutputFormat = "json"
	}
	c.OutputData(output)
	return nil
}

func addContextFlags(command *cobra.Command) {
	command.Flags().String("workspace", "", "Workspace slug; defaults to context.workspace")
	command.Flags().String("project-id", "", "Project ID; defaults to context.project.id")
}

func routeContext(command *cobra.Command) (config.RouteContext, error) {
	if c == nil {
		return config.RouteContext{}, fmt.Errorf("link command configuration is not initialized")
	}
	workspace, err := optionalStringFlag(command, "workspace")
	if err != nil {
		return config.RouteContext{}, err
	}
	projectID, err := optionalStringFlag(command, "project-id")
	if err != nil {
		return config.RouteContext{}, err
	}
	return config.ResolveRouteContext(c.Context, workspace, projectID, true)
}

func optionalStringFlag(command *cobra.Command, name string) (*string, error) {
	if !command.Flags().Changed(name) {
		return nil, nil
	}
	value, err := command.Flags().GetString(name)
	if err != nil {
		return nil, err
	}
	return &value, nil
}

func addListFlags(command *cobra.Command) {
	command.Flags().String("cursor", "", "Cursor for the next or previous link page")
	command.Flags().Int("per-page", 0, "Links per page (Plane default 20; valid range 1-100)")
	command.Flags().String("fields", "", "Comma-separated link fields to return")
	command.Flags().String("expand", "", "Comma-separated related fields to expand")
	command.Flags().String("order-by", "", "Link ordering field; prefix with - for descending")
}

func listOptions(command *cobra.Command) (resource.ListOptions, error) {
	perPage, err := command.Flags().GetInt("per-page")
	if err != nil {
		return resource.ListOptions{}, err
	}
	if command.Flags().Changed("per-page") && perPage == 0 {
		return resource.ListOptions{}, fmt.Errorf("--per-page must be between 1 and 100")
	}
	cursor, err := command.Flags().GetString("cursor")
	if err != nil {
		return resource.ListOptions{}, err
	}
	fields, err := command.Flags().GetString("fields")
	if err != nil {
		return resource.ListOptions{}, err
	}
	expand, err := command.Flags().GetString("expand")
	if err != nil {
		return resource.ListOptions{}, err
	}
	orderBy, err := command.Flags().GetString("order-by")
	if err != nil {
		return resource.ListOptions{}, err
	}
	options := resource.ListOptions{Cursor: cursor, PerPage: perPage, Fields: fields, Expand: expand, OrderBy: orderBy}
	_, err = options.Query()
	return options, err
}

func addDetailFlags(command *cobra.Command) {
	command.Flags().String("cursor", "", "Cursor for the next or previous link page")
	command.Flags().Int("per-page", 0, "Links per page (Plane default 20; valid range 1-100)")
	command.Flags().String("fields", "", "Comma-separated link fields to return")
	command.Flags().String("expand", "", "Comma-separated related fields to expand")
}

func detailOptions(command *cobra.Command) (resource.DetailOptions, error) {
	perPage, err := command.Flags().GetInt("per-page")
	if err != nil {
		return resource.DetailOptions{}, err
	}
	if command.Flags().Changed("per-page") && perPage == 0 {
		return resource.DetailOptions{}, fmt.Errorf("--per-page must be between 1 and 100")
	}
	cursor, err := command.Flags().GetString("cursor")
	if err != nil {
		return resource.DetailOptions{}, err
	}
	fields, err := command.Flags().GetString("fields")
	if err != nil {
		return resource.DetailOptions{}, err
	}
	expand, err := command.Flags().GetString("expand")
	if err != nil {
		return resource.DetailOptions{}, err
	}
	options := resource.DetailOptions{Cursor: cursor, PerPage: perPage, Fields: fields, Expand: expand}
	_, err = options.Query()
	return options, err
}
