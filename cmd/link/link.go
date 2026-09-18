// Package link owns the Cobra command hierarchy for Plane Work Item Links.
package link

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

// Init creates the link command tree without invoking the lazy client factory.
func Init(conf *config.Config, factory plane.ClientFactory) *cobra.Command {
	c = conf
	clientFactory = factory
	command := &cobra.Command{
		Use:     config.LinkCommandName,
		Aliases: []string{config.LinkCommandAlias},
		Short:   (&help.LinkCmd{}).Short(),
		Long:    (&help.LinkCmd{}).Long(),
		Args:    cobra.NoArgs,
	}
	command.AddCommand(
		newCreateCommand(false),
		newListCommand(false),
		newGetCommand(false),
		newUpdateCommand(false),
		newDeleteCommand(false),
		newLegacyCommand(),
	)
	return command
}
