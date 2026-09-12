package objects

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/maahsome/gron"
	"github.com/olekukonko/tablewriter"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type Version struct {
	SemVer    string `json:"SemVer"`
	GitCommit string `json:"GitCommit"`
	BuildDate string `json:"BuildDate"`
	GitRef    string `json:"GitRef"`
}

// ToJSON - Write the output as JSON
func (v *Version) ToJSON() string {
	versionJSON, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		logrus.WithError(err).Error("Error extracting JSON")
		return ""
	}
	return string(versionJSON[:])
}

func (v *Version) ToGRON() string {
	versionJSON, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		logrus.WithError(err).Error("Error extracting JSON for GRON")
	}
	subReader := strings.NewReader(string(versionJSON[:]))
	subValues := &bytes.Buffer{}
	ges := gron.NewGron(subReader, subValues)
	ges.SetMonochrome(false)
	if serr := ges.ToGron(); serr != nil {
		logrus.WithError(serr).Error("Problem generating GRON syntax")
		return ""
	}
	return subValues.String()
}

func (v *Version) ToYAML() string {
	versionYAML, err := yaml.Marshal(v)
	if err != nil {
		logrus.WithError(err).Error("Error extracting YAML")
		return ""
	}
	return string(versionYAML[:])
}

func (v *Version) ToTEXT(noHeaders bool) string {
	buf := new(bytes.Buffer)
	var row []string

	// ************************** TableWriter ******************************
	table := tablewriter.NewWriter(buf)
	if !noHeaders {
		table.SetHeader([]string{"COMPONENT", "VERSION"})
		table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	}

	table.SetAutoWrapText(false)
	table.SetAutoFormatHeaders(true)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetHeaderLine(false)
	table.SetBorder(false)
	table.SetTablePadding("\t") // pad with tabs
	table.SetNoWhiteSpace(true)

	row = []string{"Version", v.SemVer}
	table.Append(row)
	row = []string{"BuildDate", v.BuildDate}
	table.Append(row)
	row = []string{"GitCommit", v.GitCommit}
	table.Append(row)
	row = []string{"GitRef", v.GitRef}
	table.Append(row)

	table.Render()

	return buf.String()
}
