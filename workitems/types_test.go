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
