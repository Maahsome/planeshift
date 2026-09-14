package project

import (
	"encoding/json"
	"fmt"

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
	output, err := projectresource.RawOutput(value)
	if err != nil {
		return err
	}
	if c.OutputFormat == "" {
		c.OutputFormat = "json"
	}
	c.OutputData(output)
	return nil
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
