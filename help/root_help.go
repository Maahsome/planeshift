package help

import (
	"fmt"

	"github.com/fatih/color"
)

type RootCmd struct {
}

func (r *RootCmd) Short() string {
	return "Root command for planeshift"
}

func (r *RootCmd) Long() string {
	longText := ""
	yellow := color.New(color.FgYellow).SprintFunc()

	longText += `EXAMPLE:
  Manage Plane resources
`
	longText = fmt.Sprintf("%s\n    > %s\n", longText, yellow(`planeshift project --help`))

	longText += `
  The Project command is also available as planeshift projects.
`

	return longText
}
