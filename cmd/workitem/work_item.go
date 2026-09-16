// Package workitem owns the resource-oriented Cobra hierarchy for Plane Work
// Items. Route compatibility commands are intentionally nested and hidden.
package workitem

import (
	"planeshift/config"
	"planeshift/help"
	"planeshift/plane"

	"github.com/spf13/cobra"
)

var (
	c             *config.Config
	clientFactory plane.ClientFactory
)

// Init builds the Work Item command without invoking the lazy client factory.
func Init(conf *config.Config, factory plane.ClientFactory) *cobra.Command {
	c = conf
	clientFactory = factory
	command := &cobra.Command{
		Use:     config.WorkItemCommandName,
		Aliases: []string{config.WorkItemCommandAlias},
		Short:   (&help.WorkItemCmd{}).Short(),
		Long:    (&help.WorkItemCmd{}).Long(),
		Args:    cobra.NoArgs,
	}
	command.AddCommand(
		newSearchCommand(false),
		newGetByIdentifierCommand(false),
		newListCommand(false),
		newCreateCommand(false),
		newGetCommand(false),
		newUpdateCommand(false),
		newDeleteCommand(false),
		newRelationsListCommand(),
		newRelationsCreateCommand(),
		newLegacyCommand(),
	)
	return command
}
