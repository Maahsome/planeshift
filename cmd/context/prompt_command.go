package context

import (
	"fmt"

	"planeshift/config"
	"planeshift/help"

	"github.com/spf13/cobra"
)

func newPromptCommand(conf *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "prompt",
		Short: (&help.ContextPromptCmd{}).Short(),
		Long:  (&help.ContextPromptCmd{}).Long(),
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			return outputPrompt(conf)
		},
	}
}

func outputPrompt(conf *config.Config) error {
	if conf == nil {
		return fmt.Errorf("context command configuration is not initialized")
	}

	fmt.Printf("%s | %s\n", conf.Context.Workspace, conf.Context.Project.Name)
	return nil
}
