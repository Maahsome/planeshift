package help

import (
	"fmt"

	"github.com/fatih/color"
)

type GetVersionCmd struct {
}

func (g *GetVersionCmd) Short() string {
	return "Get version information"
}

func (g *GetVersionCmd) Long() string {
	longText := ""
	yellow := color.New(color.FgYellow).SprintFunc()

	longText += `EXAMPLE:
  Get version information about the plane
`
	longText = fmt.Sprintf("%s\n    > %s\n", longText, yellow(`planeshift get version --help`))

	longText += `
      VERSION    COMMIT    BUILD_DATE
`

	return longText
}
