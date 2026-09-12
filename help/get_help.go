package help

import (
	"fmt"

	"github.com/fatih/color"
)

type GetCmd struct {
}

func (g *GetCmd) Short() string {
	return "Get various plane resources"
}

func (g *GetCmd) Long() string {
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
