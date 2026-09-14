package objects_test

import (
	"encoding/json"
	"strings"
	"testing"

	projectobject "planeshift/objects"
	projectresource "planeshift/projects"
)

func TestProjectStructuredOutputsPreserveSingleProjectFields(t *testing.T) {
	var value projectresource.Project
	if err := json.Unmarshal([]byte(`{
  "id":"project-1",
  "identifier":"PRJ",
  "name":"Project X",
  "description_text":null,
  "network":2,
  "is_deployed":false,
  "icon_prop":{"name":"rocket","size":2},
  "default_state":null,
  "future_field":{"kept":true}
}`), &value); err != nil {
		t.Fatal(err)
	}

	output, err := projectobject.NewProject(value)
	if err != nil {
		t.Fatal(err)
	}
	for _, rendered := range map[string]string{
		"JSON": output.ToJSON(),
		"YAML": output.ToYAML(),
		"GRON": output.ToGRON(),
		"RAW":  output.ToRAW(),
	} {
		if rendered == "" {
			t.Fatalf("%s output is empty", rendered)
		}
		for _, field := range []string{"future_field", "description_text", "icon_prop", "default_state"} {
			if !strings.Contains(rendered, field) {
				t.Fatalf("%s output omitted %q: %s", rendered, field, rendered)
			}
		}
	}
	if !strings.Contains(output.ToJSON(), `"is_deployed": false`) {
		t.Fatalf("JSON output omitted dynamic boolean: %s", output.ToJSON())
	}
	if !strings.Contains(output.ToRAW(), `"default_state":null`) {
		t.Fatalf("raw output changed explicit null: %s", output.ToRAW())
	}
}

func TestProjectStructuredOutputsPreserveCursorPage(t *testing.T) {
	var value projectresource.ProjectPage
	if err := json.Unmarshal([]byte(`{
  "next_cursor":"20:1:0",
  "prev_cursor":null,
  "count":1,
  "total_results":1,
  "results":[{"id":"project-1","identifier":"PRJ","name":"Project","unknown_result":null}],
  "future_page":{"preserve":true}
}`), &value); err != nil {
		t.Fatal(err)
	}

	output, err := projectobject.NewProject(value)
	if err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{"next_cursor", "results", "unknown_result", "future_page"} {
		if !strings.Contains(output.ToJSON(), field) || !strings.Contains(output.ToRAW(), field) {
			t.Fatalf("structured output omitted page field %q: JSON=%s RAW=%s", field, output.ToJSON(), output.ToRAW())
		}
	}
}

func TestProjectConstructorRejectsUnsupportedValues(t *testing.T) {
	if _, err := projectobject.NewProject(make(chan int)); err == nil {
		t.Fatal("NewProject accepted an unsupported value")
	}
}

func TestProjectTextRendersSingleSummaryRowAndSuppressesHeaders(t *testing.T) {
	output, err := projectobject.NewProject(map[string]any{
		"id":          "project-1",
		"identifier":  "PRJ",
		"name":        "Project X",
		"workspace":   "workspace-1",
		"network":     2,
		"description": "do not render this long field",
		"future_field": map[string]any{
			"secret": "do not render dynamic data",
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	withHeaders := output.ToTEXT(false)
	for _, header := range []string{"ID", "IDENTIFIER", "NAME", "WORKSPACE", "NETWORK"} {
		if !strings.Contains(withHeaders, header) {
			t.Fatalf("text output omitted header %q: %q", header, withHeaders)
		}
	}
	for _, value := range []string{"project-1", "PRJ", "Project X", "workspace-1", "2"} {
		if !strings.Contains(withHeaders, value) {
			t.Fatalf("text output omitted value %q: %q", value, withHeaders)
		}
	}
	if strings.Contains(withHeaders, "do not render") {
		t.Fatalf("text output included non-summary fields: %q", withHeaders)
	}

	withoutHeaders := output.ToTEXT(true)
	for _, header := range []string{"ID", "IDENTIFIER", "NAME", "WORKSPACE", "NETWORK"} {
		if strings.Contains(withoutHeaders, header) {
			t.Fatalf("header %q was not suppressed: %q", header, withoutHeaders)
		}
	}
	if !strings.Contains(withoutHeaders, "project-1") {
		t.Fatalf("headerless output omitted row: %q", withoutHeaders)
	}
}

func TestProjectTextRendersPageInResultOrderAndBlanksNullableFields(t *testing.T) {
	output, err := projectobject.NewProject(map[string]any{
		"next_cursor": "20:1:0",
		"results": []any{
			map[string]any{"id": "project-1", "identifier": "ONE", "name": "First", "workspace": nil, "network": nil},
			map[string]any{"id": "project-2", "identifier": "TWO", "name": "Second", "workspace": "workspace-2", "network": 0},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	textOutput := output.ToTEXT(false)
	first := strings.Index(textOutput, "First")
	second := strings.Index(textOutput, "Second")
	if first == -1 || second == -1 || first >= second {
		t.Fatalf("page result order = %q", textOutput)
	}
	for _, value := range []string{"project-1", "ONE", "project-2", "TWO", "workspace-2", "0"} {
		if !strings.Contains(textOutput, value) {
			t.Fatalf("page text output omitted %q: %q", value, textOutput)
		}
	}
}

func TestProjectTextRejectsNonObjectResultsSafely(t *testing.T) {
	output, err := projectobject.NewProject(map[string]any{"results": []any{"not-an-object"}})
	if err != nil {
		t.Fatal(err)
	}
	if rendered := output.ToTEXT(false); rendered != "" {
		t.Fatalf("invalid page text output = %q, want empty", rendered)
	}
}
