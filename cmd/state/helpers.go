package state

import (
	"encoding/json"
	"fmt"

	stateResource "planeshift/states"

	"github.com/spf13/cobra"
	"planeshift/objects"
)

func stateClient() (*stateResource.Client, error) {
	client, err := clientFactory.New()
	if err != nil {
		return nil, err
	}
	return stateResource.NewClient(client), nil
}

func outputState(value any) error {
	if c == nil {
		return fmt.Errorf("state command configuration is not initialized")
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

func optionalBoolFlag(command *cobra.Command, name string) (*bool, error) {
	if !command.Flags().Changed(name) {
		return nil, nil
	}
	value, err := command.Flags().GetBool(name)
	if err != nil {
		return nil, err
	}
	return &value, nil
}

func optionalJSONFlag(command *cobra.Command, name string) (*json.RawMessage, error) {
	if !command.Flags().Changed(name) {
		return nil, nil
	}
	value, err := command.Flags().GetString(name)
	if err != nil {
		return nil, err
	}
	raw := json.RawMessage(value)
	if !json.Valid(raw) {
		return nil, fmt.Errorf("--%s must contain valid JSON", name)
	}
	return &raw, nil
}

func addStateFields(command *cobra.Command) {
	command.Flags().String("description", "", "State description")
	command.Flags().String("sequence", "", "State sequence as a JSON number, string, or null")
	command.Flags().String("group", "", "State group")
	command.Flags().Bool("is-triage", false, "Mark the state as triage")
	command.Flags().Bool("default", false, "Make the state the default")
	command.Flags().String("external-source", "", "External source")
	command.Flags().String("external-id", "", "External ID")
}
