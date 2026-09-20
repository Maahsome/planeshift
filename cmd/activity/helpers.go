package activity

import (
	"fmt"

	"planeshift/activities"
	"planeshift/config"
	"planeshift/objects"

	"github.com/spf13/cobra"
)

func activityClient() (*activities.Client, error) {
	if clientFactory == nil {
		return nil, fmt.Errorf("activity client factory is not initialized")
	}
	client, err := clientFactory.New()
	if err != nil {
		return nil, err
	}
	return activities.NewClient(client), nil
}

func outputActivity(value any) error {
	if c == nil {
		return fmt.Errorf("activity command configuration is not initialized")
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
		return config.RouteContext{}, fmt.Errorf("activity command configuration is not initialized")
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
	command.Flags().String("cursor", "", "Cursor for the next or previous activity page")
	command.Flags().Int("per-page", 0, "Activities per page (Plane default 20; valid range 1-100)")
	command.Flags().String("fields", "", "Comma-separated activity fields to return")
	command.Flags().String("expand", "", "Comma-separated related fields to expand")
	command.Flags().String("order-by", "", "Activity ordering field; prefix with - for descending")
}

func listOptions(command *cobra.Command) (activities.ListOptions, error) {
	perPage, err := command.Flags().GetInt("per-page")
	if err != nil {
		return activities.ListOptions{}, err
	}
	if command.Flags().Changed("per-page") && perPage == 0 {
		return activities.ListOptions{}, fmt.Errorf("--per-page must be between 1 and 100")
	}
	cursor, err := command.Flags().GetString("cursor")
	if err != nil {
		return activities.ListOptions{}, err
	}
	fields, err := command.Flags().GetString("fields")
	if err != nil {
		return activities.ListOptions{}, err
	}
	expand, err := command.Flags().GetString("expand")
	if err != nil {
		return activities.ListOptions{}, err
	}
	orderBy, err := command.Flags().GetString("order-by")
	if err != nil {
		return activities.ListOptions{}, err
	}
	options := activities.ListOptions{Cursor: cursor, PerPage: perPage, Fields: fields, Expand: expand, OrderBy: orderBy}
	_, err = options.Query()
	return options, err
}

func addDetailFlags(command *cobra.Command) {
	command.Flags().String("cursor", "", "Cursor for the next or previous activity page")
	command.Flags().Int("per-page", 0, "Activities per page (Plane default 20; valid range 1-100)")
	command.Flags().String("fields", "", "Comma-separated activity fields to return")
	command.Flags().String("expand", "", "Comma-separated related fields to expand")
	command.Flags().String("order-by", "", "Activity ordering field; prefix with - for descending")
}

func detailOptions(command *cobra.Command) (activities.DetailOptions, error) {
	perPage, err := command.Flags().GetInt("per-page")
	if err != nil {
		return activities.DetailOptions{}, err
	}
	if command.Flags().Changed("per-page") && perPage == 0 {
		return activities.DetailOptions{}, fmt.Errorf("--per-page must be between 1 and 100")
	}
	cursor, err := command.Flags().GetString("cursor")
	if err != nil {
		return activities.DetailOptions{}, err
	}
	fields, err := command.Flags().GetString("fields")
	if err != nil {
		return activities.DetailOptions{}, err
	}
	expand, err := command.Flags().GetString("expand")
	if err != nil {
		return activities.DetailOptions{}, err
	}
	orderBy, err := command.Flags().GetString("order-by")
	if err != nil {
		return activities.DetailOptions{}, err
	}
	options := activities.DetailOptions{Cursor: cursor, PerPage: perPage, Fields: fields, Expand: expand, OrderBy: orderBy}
	_, err = options.Query()
	return options, err
}
