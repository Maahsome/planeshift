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
  Get various information about plane resources
`
	longText = fmt.Sprintf("%s\n    > %s\n", longText, yellow(`planeshift get users --help`))

	longText += `
      ID    NAME    EMAIL
`

	return longText
}
