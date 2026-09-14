package projects

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestProjectRoundTripPreservesNullableDynamicAndUnknownFields(t *testing.T) {
	sample := []byte(`{
  "id":"project-1",
  "name":"Project X",
  "identifier":"PRJ",
  "description":"",
  "description_text":null,
  "description_html":"<p>Project</p>",
  "total_members":1,
  "total_cycles":0,
  "total_modules":0,
  "is_member":true,
  "member_role":20,
  "is_deployed":false,
  "created_at":"2024-01-01T00:00:00Z",
  "updated_at":"2024-01-02T00:00:00Z",
  "network":2,
  "emoji":null,
  "icon_prop":{"name":"rocket","size":2},
  "module_view":true,
  "cycle_view":true,
  "inbox_view":false,
  "page_view":true,
  "issue_views_view":true,
  "cover_image":null,
  "archive_in":0,
  "close_in":0,
  "created_by":"user-1",
  "updated_by":"user-1",
  "workspace":"workspace-1",
  "default_assignee":null,
  "project_lead":"user-1",
  "estimate":null,
  "default_state":null,
  "template_id":"template-1",
  "guest_view_all_features":false,
  "external_source":null,
  "external_id":null,
  "is_issue_type_enabled":true,
  "is_time_tracking_enabled":false,
  "future_field":{"kept":true}
}`)

	var project Project
	if err := json.Unmarshal(sample, &project); err != nil {
		t.Fatal(err)
	}
	if project.Description == nil || *project.Description != "" {
		t.Fatalf("description = %#v", project.Description)
	}
	if project.DescriptionText != nil {
		t.Fatalf("description_text = %#v, want nil", project.DescriptionText)
	}
	if string(project.IsDeployed) != "false" {
		t.Fatalf("is_deployed = %s", project.IsDeployed)
	}
	if string(project.IconProp) != `{"name":"rocket","size":2}` {
		t.Fatalf("icon_prop = %s", project.IconProp)
	}
	if string(project.Unknown["future_field"]) != `{"kept":true}` {
		t.Fatalf("unknown fields = %#v", project.Unknown)
	}
	encoded, err := json.Marshal(project)
	if err != nil {
		t.Fatal(err)
	}
	var roundTrip map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &roundTrip); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"description_text", "emoji", "cover_image", "default_state", "future_field"} {
		if _, ok := roundTrip[key]; !ok {
			t.Fatalf("round trip omitted %q: %s", key, encoded)
		}
	}
	if string(roundTrip["description_text"]) != "null" || string(roundTrip["future_field"]) != `{"kept":true}` {
		t.Fatalf("round trip changed null/unknown values: %s", encoded)
	}
}

func TestProjectSupportsBothDocumentedIsDeployedTypes(t *testing.T) {
	for _, sample := range []string{`{"is_deployed":false}`, `{"is_deployed":1}`, `{"is_deployed":null}`} {
		var project Project
		if err := json.Unmarshal([]byte(sample), &project); err != nil {
			t.Fatalf("sample %s: %v", sample, err)
		}
		var fields map[string]json.RawMessage
		if err := json.Unmarshal([]byte(sample), &fields); err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(project.IsDeployed, fields["is_deployed"]) || !json.Valid(project.IsDeployed) {
			t.Fatalf("is_deployed = %s, want raw %s", project.IsDeployed, fields["is_deployed"])
		}
	}
}

func TestProjectPageRoundTripPreservesSharedAndProjectMetadata(t *testing.T) {
	sample := []byte(`{
  "grouped_by":"state",
  "sub_grouped_by":"priority",
  "total_count":150,
  "next_cursor":"20:1:0",
  "prev_cursor":null,
  "next_page_results":true,
  "prev_page_results":false,
  "count":20,
  "total_pages":8,
  "total_results":150,
  "extra_stats":null,
  "results":[{"id":"project-1","name":"Project","identifier":"PRJ","unknown_result":null}],
  "future_page":{"preserve":true}
}`)
	var page ProjectPage
	if err := json.Unmarshal(sample, &page); err != nil {
		t.Fatal(err)
	}
	if page.GroupedBy == nil || *page.GroupedBy != "state" || page.TotalCount == nil || *page.TotalCount != 150 {
		t.Fatalf("project page metadata = %#v", page)
	}
	if page.NextCursor == nil || *page.NextCursor != "20:1:0" || page.PrevCursor != nil {
		t.Fatalf("cursor metadata = %#v", page.CursorPage)
	}
	if len(page.Results) != 1 || string(page.Results[0].Unknown["unknown_result"]) != "null" || string(page.Unknown["future_page"]) != `{"preserve":true}` {
		t.Fatalf("page dynamic values = %#v", page)
	}
	encoded, err := json.Marshal(page)
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"grouped_by", "sub_grouped_by", "total_count", "next_cursor", "future_page"} {
		if !strings.Contains(string(encoded), `"`+key+`"`) {
			t.Fatalf("page output omitted %q: %s", key, encoded)
		}
	}
}

func TestRequestPresenceAndWireNames(t *testing.T) {
	falseValue := false
	zeroValue := 0
	create := CreateProjectRequest{
		Name: "Project", Identifier: "PRJ", IntakeView: &falseValue, ArchiveIn: &zeroValue,
	}
	data, err := json.Marshal(create)
	if err != nil {
		t.Fatal(err)
	}
	output := string(data)
	for _, expected := range []string{`"intake_view":false`, `"archive_in":0`} {
		if !strings.Contains(output, expected) {
			t.Fatalf("request omitted explicit value %s: %s", expected, output)
		}
	}
	if strings.Contains(output, `"inbox_view"`) {
		t.Fatalf("create request used response-only inbox_view name: %s", output)
	}

	minimal, err := json.Marshal(CreateProjectRequest{Name: "Project", Identifier: "PRJ"})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(minimal), `"description"`) || strings.Contains(string(minimal), `"module_view"`) {
		t.Fatalf("minimal request did not omit optional fields: %s", minimal)
	}
	template, err := json.Marshal(CreateProjectFromTemplateRequest{TemplateID: "template"})
	if err != nil {
		t.Fatal(err)
	}
	if string(template) != `{"template_id":"template"}` {
		t.Fatalf("template request = %s", template)
	}
}

func TestListOptionsUsesDocumentedQueryAndValidation(t *testing.T) {
	options := ListOptions{Cursor: "cursor:1", PerPage: 100, Fields: "id,name", Expand: "members", OrderBy: "-created_at"}
	query, err := options.Query()
	if err != nil {
		t.Fatal(err)
	}
	want := "cursor=cursor%3A1&expand=members&fields=id%2Cname&order_by=-created_at&per_page=100"
	if got := query.Encode(); got != want {
		t.Fatalf("query = %q, want %q", got, want)
	}
	for _, perPage := range []int{-1, 101} {
		if _, err := (ListOptions{PerPage: perPage}).Query(); err == nil {
			t.Fatalf("ListOptions accepted per_page=%d", perPage)
		}
	}
}
