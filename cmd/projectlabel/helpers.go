package projectlabel

import (
	"fmt"

	"planeshift/objects"
	projectlabelresource "planeshift/projectlabels"

	"github.com/spf13/cobra"
)

func projectLabelClient() (*projectlabelresource.Client, error) {
	client, err := clientFactory.New()
	if err != nil {
		return nil, err
	}
	return projectlabelresource.NewClient(client), nil
}

func outputProjectLabel(value any) error {
	if commandConfig == nil {
		return fmt.Errorf("project label command configuration is not initialized")
	}
	output, err := objects.NewRawJSONFromValue(value)
	if err != nil {
		return err
	}
	if commandConfig.OutputFormat == "" {
		commandConfig.OutputFormat = "json"
	}
	commandConfig.OutputData(output)
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

func optionalFloat64Flag(cmd *cobra.Command, name string) (*float64, error) {
	if !cmd.Flags().Changed(name) {
		return nil, nil
	}
	value, err := cmd.Flags().GetFloat64(name)
	if err != nil {
		return nil, err
	}
	return &value, nil
}
