// Package label owns the Cobra command hierarchy for project-scoped Plane
// labels.
package label

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

// Init creates the label command tree without invoking the lazy client
// factory. Resource operations resolve the factory only when they run.
func Init(conf *config.Config, factory plane.ClientFactory) *cobra.Command {
	c = conf
	clientFactory = factory
	command := &cobra.Command{
		Use:     config.LabelCommandName,
		Aliases: []string{config.LabelCommandAlias},
		Short:   (&help.LabelCmd{}).Short(),
		Long:    (&help.LabelCmd{}).Long(),
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
