package state

import (
	stateResource "planeshift/states"

	"github.com/spf13/cobra"
)

func newCreateCommand() *cobra.Command {
	command := &cobra.Command{
		Use:  "create",
		Args: cobra.NoArgs,
		RunE: runCreate,
	}
	addContextFlags(command)
	command.Flags().String("name", "", "State name")
	command.Flags().String("color", "", "State color")
	addStateFields(command)
	_ = command.MarkFlagRequired("name")
	_ = command.MarkFlagRequired("color")
	return command
}

func runCreate(command *cobra.Command, args []string) error {
	name, err := command.Flags().GetString("name")
	if err != nil {
		return err
	}
	color, err := command.Flags().GetString("color")
	if err != nil {
		return err
	}
	description, err := optionalStringFlag(command, "description")
	if err != nil {
		return err
	}
	sequence, err := optionalJSONFlag(command, "sequence")
	if err != nil {
		return err
	}
	group, err := optionalStringFlag(command, "group")
	if err != nil {
		return err
	}
	isTriage, err := optionalBoolFlag(command, "is-triage")
	if err != nil {
		return err
	}
	defaultState, err := optionalBoolFlag(command, "default")
	if err != nil {
		return err
	}
	externalSource, err := optionalStringFlag(command, "external-source")
	if err != nil {
		return err
	}
	externalID, err := optionalStringFlag(command, "external-id")
	if err != nil {
		return err
	}
	request := stateResource.CreateStateRequest{
		Name: name, Description: description, Color: color, Sequence: sequence,
		Group: group, IsTriage: isTriage, Default: defaultState,
		ExternalSource: externalSource, ExternalID: externalID,
	}
	route, err := routeContext(command)
	if err != nil {
		return err
	}
	client, err := stateClient()
	if err != nil {
		return err
	}
	state, _, err := client.Create(command.Context(), route.Workspace, route.ProjectID, request)
	if err != nil {
		return err
	}
	return outputState(state)
}
