package links

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestLinkPreservesNullableDynamicAndUnknownFields(t *testing.T) {
	input := `{"id":"link-1","title":null,"url":"https://example.test/a","metadata":null,"created_at":"2024-01-01T00:00:00Z","created_by":{"id":"user-1"},"project":"project-1","workspace":["workspace-1"],"issue":null,"future":{"nested":[1,null]}}`
	var link Link
	if err := json.Unmarshal([]byte(input), &link); err != nil {
		t.Fatal(err)
	}
	if link.Title != nil || string(link.Metadata) != "null" || string(link.Issue) != "null" {
		t.Fatalf("nullable values were not preserved: %#v", link)
	}
	if string(link.Unknown["future"]) != `{"nested":[1,null]}` {
		t.Fatalf("unknown field = %s", link.Unknown["future"])
	}
	data, err := json.Marshal(link)
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]json.RawMessage
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{"title", "metadata", "issue", "future"} {
		if _, ok := got[field]; !ok {
			t.Fatalf("round trip omitted %q: %s", field, data)
		}
	}
}

func TestLinkPagePreservesEnvelopeAndDetailShape(t *testing.T) {
	input := `{"grouped_by":"state","sub_grouped_by":"priority","total_count":1,"next_cursor":"next","prev_cursor":null,"next_page_results":true,"prev_page_results":false,"count":1,"total_pages":1,"total_results":1,"extra_stats":null,"results":[{"id":"link-1","url":"https://example.test"}],"future_page":{"keep":true}}`
	var page LinkPage
	if err := json.Unmarshal([]byte(input), &page); err != nil {
		t.Fatal(err)
	}
	if page.GroupedBy == nil || *page.GroupedBy != "state" || page.TotalCount == nil || *page.TotalCount != 1 {
		t.Fatalf("grouping metadata = %#v", page)
	}
	if page.PrevCursor != nil || string(page.ExtraStats) != "null" {
		t.Fatalf("nullable page values = %#v / %s", page.PrevCursor, page.ExtraStats)
	}
	if string(page.Unknown["future_page"]) != `{"keep":true}` {
		t.Fatalf("unknown page field = %s", page.Unknown["future_page"])
	}
	var detail LinkDetailResponse
	if err := json.Unmarshal([]byte(input), &detail); err != nil {
		t.Fatal(err)
	}
	if len(detail.Results) != 1 || detail.Results[0].ID == nil || *detail.Results[0].ID != "link-1" {
		t.Fatalf("detail response = %#v", detail)
	}
	data, err := json.Marshal(detail)
	if err != nil || !strings.Contains(string(data), `"future_page"`) {
		t.Fatalf("detail round trip = %s, err=%v", data, err)
	}
}

func TestLinkRequestsPreservePresenceAndValidateRequiredURL(t *testing.T) {
	title := ""
	createData, err := json.Marshal(CreateLinkRequest{URL: "https://example.test", Title: &title})
	if err != nil {
		t.Fatal(err)
	}
	if string(createData) != `{"url":"https://example.test","title":""}` {
		t.Fatalf("create body = %s", createData)
	}
	urlValue := ""
	updateData, err := json.Marshal(UpdateLinkRequest{URL: &urlValue})
	if err != nil {
		t.Fatal(err)
	}
	if string(updateData) != `{"url":""}` {
		t.Fatalf("update body = %s", updateData)
	}
	if err := (CreateLinkRequest{}).validate(); err == nil {
		t.Fatal("blank URL was accepted")
	}
}

func TestLinkOptionsAllowOnlyDocumentedQueries(t *testing.T) {
	list, err := (ListOptions{Cursor: "20:1:0", PerPage: 100, Fields: "id,url", Expand: "project", OrderBy: "-created_at"}).Query()
	if err != nil {
		t.Fatal(err)
	}
	if got := list.Encode(); got != "cursor=20%3A1%3A0&expand=project&fields=id%2Curl&order_by=-created_at&per_page=100" {
		t.Fatalf("list query = %s", got)
	}
	detail, err := (DetailOptions{Cursor: "next", PerPage: 20, Fields: "id", Expand: "workspace"}).Query()
	if err != nil {
		t.Fatal(err)
	}
	if got := detail.Encode(); got != "cursor=next&expand=workspace&fields=id&per_page=20" {
		t.Fatalf("detail query = %s", got)
	}
	if _, err := (DetailOptions{PerPage: 101}).Query(); err == nil {
		t.Fatal("detail accepted per_page > 100")
	}
}
