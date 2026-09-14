package objects

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/olekukonko/tablewriter"
	"github.com/sirupsen/logrus"
)

// Project is the output object for Project responses. It retains a defensive
// JSON document so resource-specific fields remain owned by the projects
// package while every configured output format has an intentional object.
type Project struct {
	document RawJSON
}

// NewProject marshals a Project response or cursor page into a preserved
// output document without coupling objects to the route-specific projects
// package.
func NewProject(value any) (Project, error) {
	document, err := NewRawJSONFromValue(value)
	if err != nil {
		return Project{}, err
	}
	return Project{document: document}, nil
}

// MarshalJSON returns the preserved Project document.
func (p Project) MarshalJSON() ([]byte, error) {
	return p.document.MarshalJSON()
}

// ToJSON renders the preserved Project document as indented JSON.
func (p Project) ToJSON() string {
	return p.document.ToJSON()
}

// ToYAML renders the preserved Project document as YAML.
func (p Project) ToYAML() string {
	return p.document.ToYAML()
}

// ToGRON renders the preserved Project document as GRON.
func (p Project) ToGRON() string {
	return p.document.ToGRON()
}

// ToRAW returns the preserved compact Project document.
func (p Project) ToRAW() string {
	return p.document.ToRAW()
}

// ToTEXT renders one summary row for a Project or one row per result for a
// cursor page. Only stable identity fields are decoded for text output.
func (p Project) ToTEXT(noHeaders bool) string {
	rows, err := p.textRows()
	if err != nil {
		logrus.WithError(err).Error("Error extracting Project text")
		return ""
	}

	buf := new(bytes.Buffer)
	table := tablewriter.NewWriter(buf)
	if !noHeaders {
		table.SetHeader([]string{"ID", "Identifier", "Name", "Workspace", "Network"})
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
	table.SetTablePadding("\t")
	table.SetNoWhiteSpace(true)

	for _, row := range rows {
		table.Append(row)
	}
	table.Render()
	return buf.String()
}

func (p Project) textRows() ([][]string, error) {
	value, err := p.rawValue()
	if err != nil {
		return nil, err
	}
	projectValue, ok := value.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("project output is not a JSON object")
	}

	resultsValue, hasResults := projectValue["results"]
	if !hasResults {
		return [][]string{projectTextRow(projectValue)}, nil
	}
	if resultsValue == nil {
		return nil, nil
	}
	results, ok := resultsValue.([]any)
	if !ok {
		return nil, fmt.Errorf("project page results is not an array")
	}
	rows := make([][]string, 0, len(results))
	for index, result := range results {
		projectValue, ok := result.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("project page result %d is not a JSON object", index)
		}
		rows = append(rows, projectTextRow(projectValue))
	}
	return rows, nil
}

func projectTextRow(value map[string]any) []string {
	return []string{
		projectTextField(value["id"]),
		projectTextField(value["identifier"]),
		projectTextField(value["name"]),
		projectTextField(value["workspace"]),
		projectTextField(value["network"]),
	}
}

func projectTextField(value any) string {
	switch value := value.(type) {
	case nil:
		return ""
	case string:
		return value
	case json.Number:
		return value.String()
	default:
		return ""
	}
}

// rawValue decodes the preserved document for specialized renderers.
func (p Project) rawValue() (any, error) {
	return p.document.value()
}

var _ interface {
	ToJSON() string
	ToYAML() string
	ToGRON() string
	ToTEXT(bool) string
	ToRAW() string
} = Project{}

var _ json.Marshaler = Project{}
