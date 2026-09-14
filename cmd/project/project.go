// Package project owns the Cobra command hierarchy for Plane projects.
package project

import (
	"planeshift/config"
	"planeshift/help"
	"planeshift/plane"

	"github.com/spf13/cobra"
)

var (
	c             *config.Config
	clientFactory plane.ClientFactory
	projectCmd    *cobra.Command
)

// Init creates the Project command and its operation subcommands. The factory
// is retained as an injected seam and is not invoked while the tree is built.
func Init(conf *config.Config, factory plane.ClientFactory) *cobra.Command {
	c = conf
	clientFactory = factory
	projectCmd = &cobra.Command{
		Use:     config.ProjectCommandName,
		Aliases: []string{config.ProjectCommandAlias},
		Short:   (&help.ProjectCmd{}).Short(),
		Long:    (&help.ProjectCmd{}).Long(),
		Args:    cobra.NoArgs,
	}

	projectCmd.AddCommand(
		newListCommand(),
		newCreateCommand(),
		newCreateTemplateCommand(),
		newGetCommand(),
		newUpdateCommand(),
		newArchiveCommand(),
		newUnarchiveCommand(),
		newDeleteCommand(),
		newFeaturesCommand(),
	)
	return projectCmd
}
