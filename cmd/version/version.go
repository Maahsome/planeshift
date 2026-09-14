// Package version owns the canonical root version command.
package version

import (
	"encoding/json"

	"planeshift/config"
	"planeshift/help"
	"planeshift/objects"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var commandConfig *config.Config

// Init creates the root version command with the shared configuration. It
// does not construct or invoke a Plane client.
func Init(conf *config.Config) *cobra.Command {
	commandConfig = conf
	return &cobra.Command{
		Use:   "version",
		Short: (&help.VersionCmd{}).Short(),
		Long:  (&help.VersionCmd{}).Long(),
		Run:   run,
	}
}

func run(_ *cobra.Command, _ []string) {
	version, err := expressVersion()
	if err != nil {
		logrus.WithError(err).Error("Failed to express the version")
	}
	if !commandConfig.FormatOverridden {
		commandConfig.OutputFormat = "json"
	}
	commandConfig.OutputData(&version)
}

func expressVersion() (objects.Version, error) {
	var version objects.Version
	err := json.Unmarshal([]byte(commandConfig.VersionJSON), &version)
	if err != nil {
		return version, errors.Wrap(err, "Failed to unmarshal JSON")
	}

	return version, nil
}
