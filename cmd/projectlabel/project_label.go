// Package projectlabel owns the Cobra hierarchy for Plane Project Labels.
package projectlabel

import (
	"planeshift/config"
	"planeshift/help"
	"planeshift/plane"

	"github.com/spf13/cobra"
)

var (
	commandConfig *config.Config
	clientFactory plane.ClientFactory
)

// Init creates the top-level project-label command and its operations without
// constructing a client or making a network request.
func Init(conf *config.Config, factory plane.ClientFactory) *cobra.Command {
	commandConfig = conf
	clientFactory = factory
	command := &cobra.Command{
		Use:     config.ProjectLabelCommandName,
		Aliases: []string{config.ProjectLabelCommandAlias},
		Short:   (&help.ProjectLabelCmd{}).Short(),
		Long:    (&help.ProjectLabelCmd{}).Long(),
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
