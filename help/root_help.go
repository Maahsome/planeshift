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

  The Work Item State command is also available as planeshift states.

  Use planeshift context set or planeshift context get to manage the local
  workspace and project context. Use planeshift context prompt for a
  shell-friendly label such as:

    my-workspace | Demo Project

  Project, state, and project-scoped work-item commands consume the saved
  context when --workspace or --project-id is omitted. Use those flags for a
  one-command route override; overrides are not saved.
`

	return longText
}
