package project

import (
	"encoding/json"
	"fmt"

	"planeshift/config"
	"planeshift/objects"
	projectresource "planeshift/projects"

	"github.com/spf13/cobra"
)

func projectClient() (*projectresource.Client, error) {
	client, err := clientFactory.New()
	if err != nil {
		return nil, err
	}
	return projectresource.NewClient(client), nil
}

func outputProject(value any) error {
	if c == nil {
		return fmt.Errorf("project command configuration is not initialized")
	}
	output, err := objects.NewProject(value)
	if err != nil {
		return err
	}
	if c.OutputFormat == "" {
		c.OutputFormat = "json"
	}
	c.OutputData(output)
	return nil
}

func addContextFlags(command *cobra.Command, project bool) {
	command.Flags().String("workspace", "", "Workspace slug; defaults to context.workspace")
	if project {
		command.Flags().String("project-id", "", "Project ID; defaults to context.project.id")
	}
}

func routeContext(command *cobra.Command, project bool) (config.RouteContext, error) {
	if c == nil {
		return config.RouteContext{}, fmt.Errorf("project command configuration is not initialized")
	}
	workspace, err := optionalStringFlag(command, "workspace")
	if err != nil {
		return config.RouteContext{}, err
	}
	var projectID *string
	if project {
		projectID, err = optionalStringFlag(command, "project-id")
		if err != nil {
			return config.RouteContext{}, err
		}
	}
	return config.ResolveRouteContext(c.Context, workspace, projectID, project)
}

func optionalStringFlag(cmd *cobra.Command, name string) (*string, error) {
	if !cmd.Flags().Changed(name) {
		return nil, nil
	}
	value, err := cmd.Flags().GetString(name)
	if err != nil {
		return nil, err
	}
	return &value, nil
}

func optionalBoolFlag(cmd *cobra.Command, name string) (*bool, error) {
	if !cmd.Flags().Changed(name) {
		return nil, nil
	}
	value, err := cmd.Flags().GetBool(name)
	if err != nil {
		return nil, err
	}
	return &value, nil
}

func optionalIntFlag(cmd *cobra.Command, name string) (*int, error) {
	if !cmd.Flags().Changed(name) {
		return nil, nil
	}
	value, err := cmd.Flags().GetInt(name)
	if err != nil {
		return nil, err
	}
	return &value, nil
}

func optionalJSONFlag(cmd *cobra.Command, name string) (*json.RawMessage, error) {
	if !cmd.Flags().Changed(name) {
		return nil, nil
	}
	value, err := cmd.Flags().GetString(name)
	if err != nil {
		return nil, err
	}
	data := json.RawMessage(value)
	if !json.Valid(data) {
		return nil, fmt.Errorf("--%s must contain valid JSON", name)
	}
	return &data, nil
}
