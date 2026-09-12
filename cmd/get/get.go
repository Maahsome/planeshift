package get

import (
	"planeshift/config"

	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get",
	Args:  cobra.MinimumNArgs(1),
	Short: c.CmdHelp.GetCmd.Short(),
	Long:  c.CmdHelp.GetCmd.Long(),
	Run:   func(cmd *cobra.Command, args []string) {},
}

var c *config.Config

func InitSubCommands(conf *config.Config) *cobra.Command {
	c = conf
	return getCmd
}
