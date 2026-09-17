// Package state owns the Cobra command hierarchy for Plane Work Item States.
package state

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

// Init creates the state command tree without invoking the lazy client
// factory. Resource operations resolve the factory only when they run.
func Init(conf *config.Config, factory plane.ClientFactory) *cobra.Command {
	c = conf
	clientFactory = factory
	command := &cobra.Command{
		Use:     config.StateCommandName,
		Aliases: []string{config.StateCommandAlias},
		Short:   (&help.StateCmd{}).Short(),
		Long:    (&help.StateCmd{}).Long(),
		Args:    cobra.NoArgs,
	}
	command.AddCommand(
		newCreateCommand(),
		newListCommand(),
		newGetCommand(),
		newUpdateCommand(),
		newDeleteCommand(),
	)
	return command
}
