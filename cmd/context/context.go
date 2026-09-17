// Package context owns the local workspace/project context command.
package context

import (
	"planeshift/config"
	"planeshift/help"
	"planeshift/plane"

	"github.com/spf13/cobra"
)

// Saver persists a complete context through the root command's configuration
// boundary.
type Saver func(config.Context) error

// Prompter is the interactive seam used by context set. The production
// implementation delegates to survey/v2; tests can provide deterministic
// answers without a terminal.
type Prompter interface {
	Workspace() (string, error)
	Project(options []string) (string, error)
}

// Init creates the local context command tree without invoking the lazy Plane
// factory. Viper remains owned by cmd/root.go through the injected Saver.
func Init(conf *config.Config, factory plane.ClientFactory, saver Saver, prompt Prompter) *cobra.Command {
	if prompt == nil {
		prompt = surveyPrompter{}
	}
	command := &cobra.Command{
		Use:   config.ContextCommandName,
		Short: (&help.ContextCmd{}).Short(),
		Long:  (&help.ContextCmd{}).Long(),
		Args:  cobra.NoArgs,
	}
	command.AddCommand(
		newSetCommand(conf, factory, saver, prompt),
		newGetCommand(conf),
	)
	return command
}
