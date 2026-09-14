package project

import (
	"planeshift/help"

	"github.com/spf13/cobra"
)

func newFeaturesCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "features",
		Short: (&help.ProjectFeaturesCmd{}).Short(),
		Long:  (&help.ProjectFeaturesCmd{}).Long(),
		Args:  cobra.NoArgs,
	}
	command.AddCommand(newFeaturesGetCommand(), newFeaturesUpdateCommand())
	return command
}
