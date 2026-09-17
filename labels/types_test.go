package labels

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestLabelRoundTripPreservesDocumentedAndUnknownValues(t *testing.T) {
	sample := []byte(`{
  "id":"label-1",
  "created_at":"2024-01-01T00:00:00Z",
  "updated_at":"2024-01-02T00:00:00Z",
  "name":"Urgent",
  "description":"Needs attention",
  "color":"#eb5757",
  "sort_order":72416.5,
  "created_by":"user-1",
  "updated_by":{"id":"user-2"},
  "project":"project-1",
  "workspace":null,
  "parent":null,
  "external_source":"github",
  "external_id":"external-1",
  "future_scalar":42,
  "future_object":{"keep":true},
  "future_array":["kept"],
  "future_null":null
}`)

	var label Label
	if err := json.Unmarshal(sample, &label); err != nil {
		t.Fatal(err)
	}
	if label.Name == nil || *label.Name != "Urgent" || label.SortOrder == nil || *label.SortOrder != 72416.5 {
		t.Fatalf("typed label fields = %#v", label)
	}
	if string(label.CreatedBy) != `"user-1"` || string(label.UpdatedBy) != `{"id":"user-2"}` ||
		string(label.Workspace) != "null" || string(label.Parent) != "null" {
		t.Fatalf("dynamic association fields = %#v", label)
	}
	if string(label.Unknown["future_object"]) != `{"keep":true}` || string(label.Unknown["future_null"]) != "null" {
		t.Fatalf("unknown fields = %#v", label.Unknown)
	}

	encoded, err := json.Marshal(label)
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &fields); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{
		"id", "created_at", "updated_at", "name", "description", "color", "sort_order",
		"created_by", "updated_by", "project", "workspace", "parent", "external_source",
		"external_id", "future_scalar", "future_object", "future_array", "future_null",
	} {
		if _, ok := fields[key]; !ok {
			t.Fatalf("round trip omitted %q: %s", key, encoded)
		}
	}
	if string(fields["sort_order"]) != "72416.5" || string(fields["workspace"]) != "null" || string(fields["future_null"]) != "null" {
		t.Fatalf("round trip changed numeric/null values: %s", encoded)
	}
}

func TestLabelOmittedAndExplicitNullFieldsRemainDistinct(t *testing.T) {
	var omitted Label
	if err := json.Unmarshal([]byte(`{"name":"Urgent"}`), &omitted); err != nil {
		t.Fatal(err)
	}
	omittedData, err := json.Marshal(omitted)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(omittedData), `"description"`) || strings.Contains(string(omittedData), `"parent"`) {
		t.Fatalf("omitted fields were added: %s", omittedData)
	}

	var explicit Label
	if err := json.Unmarshal([]byte(`{"description":null,"parent":null,"sort_order":null}`), &explicit); err != nil {
		t.Fatal(err)
	}
	explicitData, err := json.Marshal(explicit)
	if err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{`"description":null`, `"parent":null`, `"sort_order":null`} {
		if !strings.Contains(string(explicitData), field) {
			t.Fatalf("explicit null field %s was lost: %s", field, explicitData)
		}
	}
}

func TestLabelRejectsMalformedStableScalarTypes(t *testing.T) {
	for _, sample := range []string{
		`{"name":false}`,
		`{"sort_order":"72416.5"}`,
		`{"created_at":{}}`,
		`{"external_id":[]}`,
	} {
		var label Label
		if err := json.Unmarshal([]byte(sample), &label); err == nil {
			t.Fatalf("malformed documented scalar accepted: %s", sample)
		}
	}
}

func TestLabelPagePreservesCursorGroupingExtraStatsAndUnknown(t *testing.T) {
	sample := []byte(`{
  "grouped_by":"label",
  "sub_grouped_by":"color",
  "total_count":150,
  "next_cursor":"20:1:0",
  "prev_cursor":null,
  "next_page_results":true,
  "prev_page_results":false,
  "count":20,
  "total_pages":8,
  "total_results":150,
  "extra_stats":null,
  "results":[{"id":"label-1","name":"Urgent","sort_order":2.0,"future_result":{"kept":true}}],
  "future_page":{"preserve":true}
}`)

	var page LabelPage
	if err := json.Unmarshal(sample, &page); err != nil {
		t.Fatal(err)
	}
	if page.GroupedBy == nil || *page.GroupedBy != "label" || page.SubGroupedBy == nil || *page.SubGroupedBy != "color" ||
		page.TotalCount == nil || *page.TotalCount != 150 {
		t.Fatalf("grouping metadata = %#v", page)
	}
	if page.NextCursor == nil || *page.NextCursor != "20:1:0" || page.PrevCursor != nil ||
		page.NextPageResults == nil || !*page.NextPageResults || page.Count == nil || *page.Count != 20 ||
		string(page.ExtraStats) != "null" {
		t.Fatalf("cursor metadata = %#v", page.CursorPage)
	}
	if len(page.Results) != 1 || string(page.Results[0].Unknown["future_result"]) != `{"kept":true}` ||
		string(page.Unknown["future_page"]) != `{"preserve":true}` {
		t.Fatalf("page dynamic values = %#v", page)
	}
	encoded, err := json.Marshal(page)
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"grouped_by", "sub_grouped_by", "total_count", "extra_stats", "future_page", "future_result"} {
		if !strings.Contains(string(encoded), `"`+key+`"`) {
			t.Fatalf("page output omitted %q: %s", key, encoded)
		}
	}
}

func TestLabelRequestsPreservePresenceAndValidateOnlyName(t *testing.T) {
	empty := ""
	zero := 0.0
	create := CreateLabelRequest{
		Name: "Urgent", Color: &empty, Description: &empty,
		ExternalSource: &empty, ExternalID: &empty, Parent: &empty, SortOrder: &zero,
	}
	data, err := json.Marshal(create)
	if err != nil {
		t.Fatal(err)
	}
	expected := `{"name":"Urgent","color":"","description":"","external_source":"","external_id":"","parent":"","sort_order":0}`
	if string(data) != expected {
		t.Fatalf("create request = %s, want %s", data, expected)
	}
	if err := (CreateLabelRequest{}).validate(); err == nil {
		t.Fatal("create accepted missing name")
	}
	if err := (CreateLabelRequest{Name: "   "}).validate(); err == nil {
		t.Fatal("create accepted blank name")
	}
	minimal, err := json.Marshal(CreateLabelRequest{Name: "Urgent"})
	if err != nil || string(minimal) != `{"name":"Urgent"}` {
		t.Fatalf("minimal create = %s (%v)", minimal, err)
	}

	update := UpdateLabelRequest{Name: &empty, Description: &empty, Parent: &empty, SortOrder: &zero}
	updateData, err := json.Marshal(update)
	if err != nil {
		t.Fatal(err)
	}
	expectedUpdate := `{"name":"","description":"","parent":"","sort_order":0}`
	if string(updateData) != expectedUpdate {
		t.Fatalf("update request = %s, want %s", updateData, expectedUpdate)
	}
	minimalUpdate, err := json.Marshal(UpdateLabelRequest{})
	if err != nil || string(minimalUpdate) != `{}` {
		t.Fatalf("empty update = %s (%v)", minimalUpdate, err)
	}
}

func TestListOptionsUsesOnlyDocumentedQueries(t *testing.T) {
	options := ListOptions{
		Cursor: "20:1:0", PerPage: 100, Fields: "id,name", Expand: "project,parent", OrderBy: "-sort_order",
	}
	query, err := options.Query()
	if err != nil {
		t.Fatal(err)
	}
	if got, want := query.Encode(), "cursor=20%3A1%3A0&expand=project%2Cparent&fields=id%2Cname&order_by=-sort_order&per_page=100"; got != want {
		t.Fatalf("query = %q, want %q", got, want)
	}
	if query.Get("name") != "" || query.Get("color") != "" {
		t.Fatal("list exposed undocumented name/color filters")
	}
	if _, err := (ListOptions{PerPage: 101}).Query(); err == nil {
		t.Fatal("list accepted per_page=101")
	}
	if _, err := (ListOptions{PerPage: -1}).Query(); err == nil {
		t.Fatal("list accepted negative per_page")
	}
	if query, err := (ListOptions{}).Query(); err != nil || query.Get("per_page") != "" {
		t.Fatalf("zero options query = %v (%v)", query, err)
	}
}
