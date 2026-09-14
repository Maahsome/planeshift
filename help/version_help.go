package help

import (
	"fmt"

	"github.com/fatih/color"
)

type VersionCmd struct{}

func (v *VersionCmd) Short() string {
	return "Get version information"
}

func (v *VersionCmd) Long() string {
	longText := ""
	yellow := color.New(color.FgYellow).SprintFunc()

	longText += `EXAMPLE:
  Get planeshift version information
`
	longText = fmt.Sprintf("%s\n    > %s\n", longText, yellow(`planeshift version --help`))

	longText += `
      VERSION    COMMIT    BUILD_DATE
`

	return longText
}
