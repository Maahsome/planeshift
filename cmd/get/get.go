package get

import (
	"planeshift/config"
	"planeshift/help"
	"planeshift/plane"

	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get",
	Args:  cobra.MinimumNArgs(1),
	Short: (&help.GetCmd{}).Short(),
	Long:  (&help.GetCmd{}).Long(),
	Run:   func(cmd *cobra.Command, args []string) {},
}

var c *config.Config
var clientFactory plane.ClientFactory

// InitSubCommands registers the get hierarchy below cmd.RootCmd. The factory
// is optional for compatibility with callers that only need get version; no
// command in this package constructs a Plane client during registration.
func InitSubCommands(conf *config.Config, factories ...plane.ClientFactory) *cobra.Command {
	c = conf
	clientFactory = nil
	if len(factories) > 0 {
		clientFactory = factories[0]
	}
	return getCmd
}

// PlaneClientFactory exposes the lazy seam to future resource registrations
// without exposing Viper or transport construction to resource packages.
func PlaneClientFactory() plane.ClientFactory {
	return clientFactory
}
