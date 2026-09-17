package states

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestStateRoundTripPreservesDocumentedAndUnknownValues(t *testing.T) {
	sample := []byte(`{
  "id":"state-1",
  "created_at":"2024-01-01T00:00:00Z",
  "updated_at":"2024-01-02T00:00:00Z",
  "name":"Ideation",
  "description":null,
  "color":"#eb5757",
  "workspace_slug":"ideation",
  "sequence":130000.0,
  "group":"unstarted",
  "default":false,
  "is_triage":false,
  "external_source":null,
  "external_id":"external-1",
  "created_by":"user-1",
  "updated_by":"user-2",
  "project":"project-1",
  "workspace":"workspace-1",
  "future_scalar":42,
  "future_object":{"keep":true},
  "future_array":["kept"],
  "future_null":null
}`)

	var state State
	if err := json.Unmarshal(sample, &state); err != nil {
		t.Fatal(err)
	}
	if state.Name == nil || *state.Name != "Ideation" || state.Description != nil {
		t.Fatalf("state scalar fields = %#v", state)
	}
	if string(state.Sequence) != "130000.0" || state.Default == nil || *state.Default {
		t.Fatalf("state dynamic fields = %#v", state)
	}
	if string(state.Unknown["future_object"]) != `{"keep":true}` || string(state.Unknown["future_null"]) != "null" {
		t.Fatalf("unknown fields = %#v", state.Unknown)
	}

	encoded, err := json.Marshal(state)
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &fields); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"description", "sequence", "future_scalar", "future_object", "future_array", "future_null"} {
		if _, ok := fields[key]; !ok {
			t.Fatalf("round trip omitted %q: %s", key, encoded)
		}
	}
	if string(fields["description"]) != "null" || string(fields["sequence"]) != "130000.0" || string(fields["future_null"]) != "null" {
		t.Fatalf("round trip changed dynamic/null values: %s", encoded)
	}
}

func TestStateSequencePreservesNumberAndStringAndRejectsOtherJSON(t *testing.T) {
	for _, sample := range []string{`{"sequence":130000.0}`, `{"sequence":"130000.0"}`, `{"sequence":null}`} {
		var state State
		if err := json.Unmarshal([]byte(sample), &state); err != nil {
			t.Fatalf("sample %s: %v", sample, err)
		}
		if string(state.Sequence) == "" {
			t.Fatalf("sequence was not preserved for %s", sample)
		}
	}
	for _, sample := range []string{`{"sequence":true}`, `{"sequence":{}}`, `{"sequence":[]}`, `{"name":false}`} {
		var state State
		if err := json.Unmarshal([]byte(sample), &state); err == nil {
			t.Fatalf("malformed documented scalar accepted: %s", sample)
		}
	}
}

func TestStatePagePreservesCursorGroupingExtraStatsAndUnknown(t *testing.T) {
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
  "results":[{"id":"state-1","name":"Started","sequence":"2","unknown_result":{"kept":true}}],
  "future_page":{"preserve":true}
}`)

	var page StatePage
	if err := json.Unmarshal(sample, &page); err != nil {
		t.Fatal(err)
	}
	if page.GroupedBy == nil || *page.GroupedBy != "state" || page.TotalCount == nil || *page.TotalCount != 150 {
		t.Fatalf("grouping metadata = %#v", page)
	}
	if page.NextCursor == nil || *page.NextCursor != "20:1:0" || page.PrevCursor != nil || string(page.ExtraStats) != "null" {
		t.Fatalf("cursor metadata = %#v", page.CursorPage)
	}
	if len(page.Results) != 1 || string(page.Results[0].Unknown["unknown_result"]) != `{"kept":true}` || string(page.Unknown["future_page"]) != `{"preserve":true}` {
		t.Fatalf("page dynamic values = %#v", page)
	}
	encoded, err := json.Marshal(page)
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"grouped_by", "sub_grouped_by", "total_count", "extra_stats", "future_page"} {
		if !strings.Contains(string(encoded), `"`+key+`"`) {
			t.Fatalf("page output omitted %q: %s", key, encoded)
		}
	}
}

func TestStateRequestPresenceAndValidation(t *testing.T) {
	empty := ""
	falseValue := false
	null := json.RawMessage("null")
	create := CreateStateRequest{
		Name: "State", Description: &empty, Color: "#fff", Sequence: &null,
		Group: &empty, IsTriage: &falseValue, Default: &falseValue,
		ExternalSource: &empty, ExternalID: &empty,
	}
	data, err := json.Marshal(create)
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{`"description":""`, `"sequence":null`, `"group":""`, `"is_triage":false`, `"default":false`, `"external_source":""`, `"external_id":""`} {
		if !strings.Contains(string(data), expected) {
			t.Fatalf("create request omitted %s: %s", expected, data)
		}
	}
	if err := (CreateStateRequest{Color: "#fff"}).validate(); err == nil {
		t.Fatal("create accepted missing name")
	}
	if err := (CreateStateRequest{Name: "State"}).validate(); err == nil {
		t.Fatal("create accepted missing color")
	}
	minimal, err := json.Marshal(CreateStateRequest{Name: "State", Color: "#fff"})
	if err != nil || string(minimal) != `{"name":"State","color":"#fff"}` {
		t.Fatalf("minimal create = %s (%v)", minimal, err)
	}
	update := UpdateStateRequest{Description: &empty, IsTriage: &falseValue, Sequence: &null}
	updateData, err := json.Marshal(update)
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{`"description":""`, `"is_triage":false`, `"sequence":null`} {
		if !strings.Contains(string(updateData), expected) {
			t.Fatalf("update request omitted %s: %s", expected, updateData)
		}
	}
	minimalUpdate, err := json.Marshal(UpdateStateRequest{})
	if err != nil || string(minimalUpdate) != `{}` {
		t.Fatalf("empty update = %s (%v)", minimalUpdate, err)
	}
	if err := (UpdateStateRequest{Sequence: jsonRaw(`{}`)}).validate(); err == nil {
		t.Fatal("update accepted object sequence")
	}
}

func TestListOptionsUsesOnlyDocumentedQueries(t *testing.T) {
	options := ListOptions{Cursor: "20:1:0", PerPage: 100, Fields: "id,name", Expand: "project"}
	query, err := options.Query()
	if err != nil {
		t.Fatal(err)
	}
	if got, want := query.Encode(), "cursor=20%3A1%3A0&expand=project&fields=id%2Cname&per_page=100"; got != want {
		t.Fatalf("query = %q, want %q", got, want)
	}
	if _, err := (ListOptions{PerPage: 101}).Query(); err == nil {
		t.Fatal("list accepted per_page=101")
	}
	if _, err := (ListOptions{PerPage: -1}).Query(); err == nil {
		t.Fatal("list accepted negative per_page")
	}
}

func jsonRaw(value string) *json.RawMessage {
	raw := json.RawMessage(value)
	return &raw
}
