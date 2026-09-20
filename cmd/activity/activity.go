// Package activity owns the Cobra command hierarchy for Plane Work Item
// Activity history.
package activity

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

// Init creates the activity command tree without invoking the lazy client
// factory. The primary command is activity; activities is its plural alias.
func Init(conf *config.Config, factory plane.ClientFactory) *cobra.Command {
	c = conf
	clientFactory = factory
	command := &cobra.Command{
		Use:     config.ActivityCommandName,
		Aliases: []string{config.ActivityCommandAlias},
		Short:   (&help.ActivityCmd{}).Short(),
		Long:    (&help.ActivityCmd{}).Long(),
		Args:    cobra.NoArgs,
	}
	command.AddCommand(newListCommand(false), newGetCommand(false), newLegacyCommand())
	return command
}
