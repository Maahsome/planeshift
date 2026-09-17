package workitems

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestWorkItemContractsPreserveDynamicNullableAndUnknownFields(t *testing.T) {
	data := []byte(`{"id":"item-1","name":"Item","estimate_point":null,"parent":null,"state":{"id":"state-1"},"assignees":[{"id":"user-1"}],"future":{"keep":true}}`)
	var item WorkItem
	if err := json.Unmarshal(data, &item); err != nil {
		t.Fatal(err)
	}
	if item.Unknown["future"] == nil || string(item.Parent) != "null" || string(item.EstimatePoint) != "null" {
		t.Fatalf("lossless item = %#v", item)
	}
	roundTrip, err := json.Marshal(item)
	if err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{`"parent":null`, `"estimate_point":null`, `"future":{"keep":true}`} {
		if !strings.Contains(string(roundTrip), field) {
			t.Fatalf("round trip omitted %s: %s", field, roundTrip)
		}
	}
}

func TestWorkItemPageSearchAndRelationsRetainUnknowns(t *testing.T) {
	var page WorkItemPage
	if err := json.Unmarshal([]byte(`{"next_cursor":"next","total_count":2,"results":[{"id":"one","name":"One","expanded":[]}],"page_future":false}`), &page); err != nil {
		t.Fatal(err)
	}
	if page.TotalCount == nil || *page.TotalCount != 2 || page.Unknown["page_future"] == nil || page.Results[0].Unknown["expanded"] == nil {
		t.Fatalf("page = %#v", page)
	}
	var search WorkItemSearchResponse
	if err := json.Unmarshal([]byte(`{"issues":[{"id":"one","project_id":{"id":"project"},"future":null}],"search_future":true}`), &search); err != nil {
		t.Fatal(err)
	}
	if search.Unknown["search_future"] == nil || search.Issues[0].Unknown["future"] == nil {
		t.Fatalf("search = %#v", search)
	}
	var relations WorkItemRelationPage
	if err := json.Unmarshal([]byte(`{"results":[{"id":"relation","future":{"value":1}}],"relation_future":"yes"}`), &relations); err != nil {
		t.Fatal(err)
	}
	if relations.Unknown["relation_future"] == nil || relations.Results[0].Unknown["future"] == nil {
		t.Fatalf("relations = %#v", relations)
	}
}

func TestWorkItemRelationCreateResponseUsesFlatLosslessArray(t *testing.T) {
	data := []byte(`[{"id":"relation-1","project_id":"project-1","sequence_id":2,"relation_type":"relates_to","name":"Related item","state_id":"state-1","priority":"none","created_by":"user-1","created_at":"2026-09-17T02:25:15.676423Z","updated_at":"2026-09-17T02:25:15.676434Z","updated_by":"user-1","future":{"keep":true}}]`)
	var response WorkItemRelationCreateResponse
	if err := json.Unmarshal(data, &response); err != nil {
		t.Fatal(err)
	}
	if len(response) != 1 {
		t.Fatalf("relation count = %d, want 1", len(response))
	}
	relation := response[0]
	if relation.ID != "relation-1" || relation.RelationType == nil || *relation.RelationType != "relates_to" {
		t.Fatalf("relation identity = %#v", relation)
	}
	if relation.CreatedAt == nil || *relation.CreatedAt != "2026-09-17T02:25:15.676423Z" || relation.UpdatedAt == nil || *relation.UpdatedAt != "2026-09-17T02:25:15.676434Z" {
		t.Fatalf("relation timestamps = %#v", relation)
	}
	if string(relation.ProjectID) != `"project-1"` || string(relation.SequenceID) != "2" || relation.Unknown["future"] == nil {
		t.Fatalf("relation dynamic/unknown fields = %#v", relation)
	}

	roundTrip, err := json.Marshal(response)
	if err != nil {
		t.Fatal(err)
	}
	var flat []json.RawMessage
	if err := json.Unmarshal(roundTrip, &flat); err != nil {
		t.Fatal(err)
	}
	if len(flat) != 1 || len(flat[0]) == 0 || flat[0][0] != '{' || strings.Contains(string(roundTrip), "[[") || !strings.Contains(string(roundTrip), `"future":{"keep":true}`) {
		t.Fatalf("flat lossless round trip = %s", roundTrip)
	}
}

func TestRequestPresenceAndWireNames(t *testing.T) {
	falseValue := false
	zero := 0
	empty := []string{}
	null := json.RawMessage("null")
	request := UpdateWorkItemRequest{
		Assignees: &empty, IsDraft: &falseValue, Point: &zero, EstimatePoint: &null,
	}
	request.SetNull("parent")
	data, err := json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	for _, field := range []string{`"assignees":[]`, `"is_draft":false`, `"point":0`, `"estimate_point":null`, `"parent":null`} {
		if !strings.Contains(text, field) {
			t.Fatalf("request omitted %s: %s", field, text)
		}
	}
	if strings.Contains(text, "description_html") {
		t.Fatalf("untouched field was emitted: %s", text)
	}
	if err := (CreateWorkItemRequest{}).validate(); err == nil {
		t.Fatal("empty create name accepted")
	}
	if _, err := (SearchOptions{}).Query(); err == nil {
		t.Fatal("empty search accepted")
	}
	if _, err := (ListOptions{PerPage: 101}).Query(); err == nil {
		t.Fatal("invalid per_page accepted")
	}
}
