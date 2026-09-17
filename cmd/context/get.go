package context

import (
	"fmt"

	"planeshift/config"
	"planeshift/help"
	"planeshift/objects"

	"github.com/spf13/cobra"
)

func newGetCommand(conf *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "get",
		Short: (&help.ContextGetCmd{}).Short(),
		Long:  (&help.ContextGetCmd{}).Long(),
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			return outputContext(conf)
		},
	}
}

func outputContext(conf *config.Config) error {
	if conf == nil {
		return fmt.Errorf("context command configuration is not initialized")
	}
	output, err := objects.NewRawJSONFromValue(conf.Context)
	if err != nil {
		return err
	}
	if conf.OutputFormat == "" {
		conf.OutputFormat = "json"
	}
	conf.OutputData(output)
	return nil
}
