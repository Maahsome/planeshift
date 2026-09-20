package activities

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestActivityPreservesDynamicNullableAndUnknownFields(t *testing.T) {
	data := []byte(`{
        "id":"activity-1",
        "created_at":"2024-01-01T00:00:00Z",
        "updated_at":"2024-01-01T00:00:01Z",
        "deleted_at":null,
        "verb":"updated",
        "field":null,
        "old_value":{"state":"old"},
        "new_value":["new",2,null],
        "comment":"changed state",
        "attachments":["https://example.test/a",{"name":"b"}],
        "old_identifier":null,
        "new_identifier":"state-2",
        "epoch":1704067200.5,
        "project":{"id":"project-1"},
        "workspace":"workspace-1",
        "issue":"work-item-1",
        "issue_comment":null,
        "actor":{"id":"actor-1","name":"A"},
        "future_activity":{"kept":true}
    }`)

	var activity Activity
	if err := json.Unmarshal(data, &activity); err != nil {
		t.Fatal(err)
	}
	if activity.DeletedAt != nil || activity.Field != nil || string(activity.IssueComment) != "null" {
		t.Fatalf("nullable fields = %#v/%#v/%s", activity.DeletedAt, activity.Field, activity.IssueComment)
	}
	if string(activity.OldValue) != `{"state":"old"}` || string(activity.NewValue) != `["new",2,null]` {
		t.Fatalf("dynamic values = %s / %s", activity.OldValue, activity.NewValue)
	}
	if len(activity.Attachments) == 0 || string(activity.OldIdentifier) != "null" || string(activity.NewIdentifier) != `"state-2"` {
		t.Fatalf("dynamic activity fields were lost: %#v", activity)
	}
	if activity.Epoch == nil || *activity.Epoch != 1704067200.5 {
		t.Fatalf("epoch = %v", activity.Epoch)
	}
	if string(activity.Unknown["future_activity"]) != `{"kept":true}` {
		t.Fatalf("unknown activity field = %s", activity.Unknown["future_activity"])
	}

	encoded, err := json.Marshal(activity)
	if err != nil {
		t.Fatal(err)
	}
	var got, want map[string]any
	if err := json.Unmarshal(encoded, &got); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &want); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("round trip changed activity:\n got %#v\nwant %#v", got, want)
	}
}

func TestActivityPageAndDetailPreserveEnvelope(t *testing.T) {
	data := []byte(`{
        "grouped_by":"field",
        "sub_grouped_by":null,
        "total_count":2,
        "next_cursor":"next",
        "prev_cursor":null,
        "next_page_results":true,
        "prev_page_results":false,
        "count":1,
        "total_pages":2,
        "total_results":2,
        "extra_stats":{"changes":2},
        "results":[{"id":"activity-1","old_value":"old","new_value":42}],
        "future_page":{"enabled":true}
    }`)

	var page ActivityPage
	if err := json.Unmarshal(data, &page); err != nil {
		t.Fatal(err)
	}
	if len(page.Results) != 1 || page.Results[0].ID == nil || *page.Results[0].ID != "activity-1" {
		t.Fatalf("results = %#v", page.Results)
	}
	if page.NextCursor == nil || *page.NextCursor != "next" || page.TotalCount == nil || *page.TotalCount != 2 {
		t.Fatalf("page metadata = %#v", page)
	}
	if string(page.ExtraStats) != `{"changes":2}` || string(page.Unknown["future_page"]) != `{"enabled":true}` {
		t.Fatalf("page dynamic fields = %s / %s", page.ExtraStats, page.Unknown["future_page"])
	}

	encoded, err := json.Marshal(ActivityDetailResponse{ActivityPage: page})
	if err != nil {
		t.Fatal(err)
	}
	var got, want map[string]any
	if err := json.Unmarshal(encoded, &got); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &want); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("detail round trip changed envelope:\n got %#v\nwant %#v", got, want)
	}
}

func TestActivityPreservesEveryDocumentedDynamicValueShape(t *testing.T) {
	for _, raw := range []string{`"string"`, `42.5`, `{"key":true}`, `["array"]`, `null`} {
		t.Run(raw, func(t *testing.T) {
			data := []byte(`{"old_value":` + raw + `,"new_value":` + raw + `}`)
			var activity Activity
			if err := json.Unmarshal(data, &activity); err != nil {
				t.Fatal(err)
			}
			if string(activity.OldValue) != raw || string(activity.NewValue) != raw {
				t.Fatalf("dynamic values = %s / %s, want %s", activity.OldValue, activity.NewValue, raw)
			}
			encoded, err := json.Marshal(activity)
			if err != nil {
				t.Fatal(err)
			}
			var got map[string]json.RawMessage
			if err := json.Unmarshal(encoded, &got); err != nil {
				t.Fatal(err)
			}
			if string(got["old_value"]) != raw || string(got["new_value"]) != raw {
				t.Fatalf("round trip values = %s / %s, want %s", got["old_value"], got["new_value"], raw)
			}
		})
	}
}

func TestActivityOptionsAllowOnlyDocumentedQueryValues(t *testing.T) {
	options := DetailOptions{Cursor: "next:1", PerPage: 20, Fields: "id,old_value", Expand: "actor", OrderBy: "-created_at"}
	query, err := options.Query()
	if err != nil {
		t.Fatal(err)
	}
	want := "cursor=next%3A1&expand=actor&fields=id%2Cold_value&order_by=-created_at&per_page=20"
	if query.Encode() != want {
		t.Fatalf("query = %q, want %q", query.Encode(), want)
	}
	if _, err := (ListOptions{PerPage: 101}).Query(); err == nil {
		t.Fatal("invalid per_page was accepted")
	}
}
