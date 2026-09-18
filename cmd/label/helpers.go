package label

import (
	"fmt"

	"planeshift/config"
	"planeshift/labels"
	"planeshift/objects"

	"github.com/spf13/cobra"
)

func labelClient() (*labels.Client, error) {
	client, err := clientFactory.New()
	if err != nil {
		return nil, err
	}
	return labels.NewClient(client), nil
}

func outputLabel(value any) error {
	if c == nil {
		return fmt.Errorf("label command configuration is not initialized")
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
		return config.RouteContext{}, fmt.Errorf("label command configuration is not initialized")
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

func optionalFloat64Flag(command *cobra.Command, name string) (*float64, error) {
	if !command.Flags().Changed(name) {
		return nil, nil
	}
	value, err := command.Flags().GetFloat64(name)
	if err != nil {
		return nil, err
	}
	return &value, nil
}

func addLabelFields(command *cobra.Command) {
	command.Flags().String("color", "", "Label color")
	command.Flags().String("description", "", "Label description")
	command.Flags().String("external-source", "", "External source")
	command.Flags().String("external-id", "", "External ID")
	command.Flags().String("parent", "", "Parent label ID")
	command.Flags().Float64("sort-order", 0, "Label sort order")
}
