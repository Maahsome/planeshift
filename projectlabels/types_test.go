package projectlabels

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestProjectLabelRoundTripPreservesOfficialFieldsUnknownsAndNulls(t *testing.T) {
	sample := []byte(`{"id":"550e8400-e29b-41d4-a716-446655440000","name":"Example Name","color":"#f39c12","description":null,"sort_order":65535.5,"workspace":"workspace-1","created_at":"2024-01-01T00:00:00Z","updated_at":"2024-01-01T00:00:00Z","created_by":null,"updated_by":"user-1","future_number":12,"future_string":"value","future_null":null}`)
	var label ProjectLabel
	if err := json.Unmarshal(sample, &label); err != nil {
		t.Fatal(err)
	}
	if label.ID == "" || label.Name != "Example Name" || label.SortOrder == nil || *label.SortOrder != 65535.5 {
		t.Fatalf("decoded label = %#v", label)
	}
	if label.Description != nil || label.CreatedBy != nil || label.Workspace == nil || *label.Workspace != "workspace-1" {
		t.Fatalf("decoded nullable fields = %#v", label)
	}
	encoded, err := json.Marshal(label)
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &fields); err != nil {
		t.Fatal(err)
	}
	for _, name := range projectLabelKnownFields {
		if _, ok := fields[name]; !ok {
			t.Fatalf("round trip omitted %q: %s", name, encoded)
		}
	}
	for _, expected := range []string{`"description":null`, `"created_by":null`, `"future_number":12`, `"future_string":"value"`, `"future_null":null`} {
		if !strings.Contains(string(encoded), expected) {
			t.Fatalf("round trip omitted %s: %s", expected, encoded)
		}
	}
}

func TestProjectLabelPageRoundTripPreservesGroupingPaginationAndUnknowns(t *testing.T) {
	sample := []byte(`{"grouped_by":"state","sub_grouped_by":"priority","total_count":150,"next_cursor":"20:1:0","prev_cursor":null,"next_page_results":true,"prev_page_results":false,"count":1,"total_pages":1,"total_results":1,"extra_stats":null,"results":[{"id":"label-1","name":"Example Name","description":"Description","color":"#f39c12","sort_order":1,"workspace":"workspace-1","future":null}],"future_page":{"enabled":true}}`)
	var page ProjectLabelPage
	if err := json.Unmarshal(sample, &page); err != nil {
		t.Fatal(err)
	}
	if page.GroupedBy == nil || *page.GroupedBy != "state" || page.TotalCount == nil || *page.TotalCount != 150 || len(page.Results) != 1 {
		t.Fatalf("decoded page = %#v", page)
	}
	if string(page.ExtraStats) != "null" || string(page.Unknown["future_page"]) != `{"enabled":true}` {
		t.Fatalf("page metadata = %#v unknown=%#v", page.CursorPage, page.Unknown)
	}
	encoded, err := json.Marshal(page)
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{`"grouped_by":"state"`, `"sub_grouped_by":"priority"`, `"total_count":150`, `"prev_cursor":null`, `"extra_stats":null`, `"future_page":{"enabled":true}`} {
		if !strings.Contains(string(encoded), expected) {
			t.Fatalf("page round trip omitted %s: %s", expected, encoded)
		}
	}
}

func TestProjectLabelRejectsMalformedDocumentedScalars(t *testing.T) {
	for _, sample := range []string{
		`{"id":1,"name":"Label"}`,
		`{"id":"id","name":true}`,
		`{"id":"id","name":"Label","description":1}`,
		`{"id":"id","name":"Label","color":false}`,
		`{"id":"id","name":"Label","sort_order":"first"}`,
		`{"id":"id","name":"Label","workspace":[]}`,
	} {
		var label ProjectLabel
		if err := json.Unmarshal([]byte(sample), &label); err == nil {
			t.Fatalf("accepted malformed label %s", sample)
		}
	}
}

func TestProjectLabelRequestsPreservePresenceAndValidateOnlyCreateName(t *testing.T) {
	empty := ""
	zero := float64(0)
	create := CreateProjectLabelRequest{Name: "Label", Description: &empty, Color: &empty, SortOrder: &zero}
	data, err := json.Marshal(create)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `{"name":"Label","description":"","color":"","sort_order":0}` {
		t.Fatalf("create request = %s", data)
	}
	update := UpdateProjectLabelRequest{Name: &empty, Description: &empty, Color: &empty, SortOrder: &zero}
	data, err = json.Marshal(update)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `{"name":"","description":"","color":"","sort_order":0}` {
		t.Fatalf("update request = %s", data)
	}
	if data, err := json.Marshal(UpdateProjectLabelRequest{}); err != nil || string(data) != `{}` {
		t.Fatalf("empty update = %s / %v", data, err)
	}
	for _, name := range []string{"", "  "} {
		if err := (CreateProjectLabelRequest{Name: name}).validate(); err == nil {
			t.Fatalf("accepted empty name %q", name)
		}
	}
	if err := (CreateProjectLabelRequest{Name: "Label", Description: &empty}).validate(); err != nil {
		t.Fatal(err)
	}
}

func TestProjectLabelOmittedFieldsStayOmitted(t *testing.T) {
	var label ProjectLabel
	if err := json.Unmarshal([]byte(`{"id":"id","name":"Label","future":null}`), &label); err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(label)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `{"future":null,"id":"id","name":"Label"}` {
		t.Fatalf("omitted fields were manufactured: %s", data)
	}
}
