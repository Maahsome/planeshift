package activities

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"planeshift/internal/testsupport"
	"planeshift/plane"
)

const (
	testWorkspace = "team/slash"
	testProject   = "project/slash"
	testWorkItem  = "work/item"
	testIssue     = "issue/item"
	testActivity  = "activity/item"
)

func assertActivityRequest(t *testing.T, request *http.Request, method, escapedPath string, query url.Values, authMode string) []byte {
	t.Helper()
	if request.Method != method {
		t.Errorf("method = %q, want %q", request.Method, method)
	}
	if request.URL.EscapedPath() != escapedPath {
		t.Errorf("escaped path = %q, want %q", request.URL.EscapedPath(), escapedPath)
	}
	if request.URL.Query().Encode() != query.Encode() {
		t.Errorf("query = %v, want %v", request.URL.Query(), query)
	}
	switch authMode {
	case plane.AuthModeAPIKey:
		if got := request.Header.Get("X-API-Key"); got != "secret-key" {
			t.Errorf("X-API-Key = %q, want secret-key", got)
		}
		if got := request.Header.Get("Authorization"); got != "" {
			t.Errorf("Authorization = %q, want absent", got)
		}
	case plane.AuthModeBearer:
		if got := request.Header.Get("Authorization"); got != "Bearer bearer-token" {
			t.Errorf("Authorization = %q, want bearer token", got)
		}
		if got := request.Header.Get("X-API-Key"); got != "" {
			t.Errorf("X-API-Key = %q, want absent", got)
		}
	}
	body, err := io.ReadAll(request.Body)
	if err != nil {
		t.Fatal(err)
	}
	return body
}

func activityPageFixture() map[string]any {
	return map[string]any{
		"grouped_by": "field", "sub_grouped_by": "verb", "total_count": 1,
		"next_cursor": "next", "prev_cursor": nil, "next_page_results": true,
		"prev_page_results": false, "count": 1, "total_pages": 1, "total_results": 1,
		"extra_stats": map[string]any{"changed": 1}, "future_page": map[string]any{"kept": true},
		"results": []any{map[string]any{
			"id": "activity-1", "created_at": "2024-01-01T00:00:00Z", "updated_at": "2024-01-01T00:00:01Z",
			"deleted_at": nil, "verb": "updated", "field": nil, "old_value": map[string]any{"state": "old"},
			"new_value": []any{"new", 2, nil}, "comment": "changed", "attachments": []any{"https://example.test/a"},
			"old_identifier": nil, "new_identifier": "state-2", "epoch": 1704067200.5,
			"project": map[string]any{"id": "project-1"}, "workspace": "workspace-1", "issue": "work-item-1",
			"issue_comment": nil, "actor": map[string]any{"id": "actor-1"}, "future_activity": true,
		}},
	}
}

func TestListUsesPrimaryRouteQueryAuthPageAndMetadata(t *testing.T) {
	query := url.Values{"cursor": {"20:1:0"}, "expand": {"actor"}, "fields": {"id,old_value"}, "order_by": {"-created_at"}, "per_page": {"20"}}
	server := testsupport.NewServer(t, func(writer http.ResponseWriter, request *http.Request) {
		body := assertActivityRequest(t, request, http.MethodGet, "/api/v1/workspaces/team%2Fslash/projects/project%2Fslash/work-items/work%2Fitem/activities/", query, plane.AuthModeAPIKey)
		if len(body) != 0 {
			t.Errorf("list body = %q, want empty", body)
		}
		writer.Header().Set("X-RateLimit-Remaining", "7")
		writer.Header().Set("X-RateLimit-Reset", "1234")
		testsupport.JSON(writer, http.StatusOK, activityPageFixture())
	})
	shared, err := plane.NewClient(plane.Options{BaseURL: server.URL, AuthMode: plane.AuthModeAPIKey, APIKey: "secret-key"})
	if err != nil {
		t.Fatal(err)
	}
	client := NewClient(shared)
	page, response, err := client.List(context.Background(), testWorkspace, testProject, testWorkItem, ListOptions{
		Cursor: "20:1:0", PerPage: 20, Fields: "id,old_value", Expand: "actor", OrderBy: "-created_at",
	})
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != http.StatusOK || response.RateLimit.Remaining != 7 || response.RateLimit.Reset != 1234 {
		t.Fatalf("response metadata = %#v", response)
	}
	if len(page.Results) != 1 || page.Results[0].ID == nil || *page.Results[0].ID != "activity-1" || string(page.Results[0].OldValue) != `{"state":"old"}` {
		t.Fatalf("page = %#v", page)
	}
	if string(page.Unknown["future_page"]) != `{"kept":true}` || page.NextCursor == nil || *page.NextCursor != "next" {
		t.Fatalf("page metadata = %#v", page)
	}
}

func TestGetUsesPrimaryDetailRouteQueryAndDecodesPageEnvelope(t *testing.T) {
	query := url.Values{"cursor": {"next"}, "expand": {"actor"}, "fields": {"id"}, "order_by": {"created_at"}, "per_page": {"10"}}
	server := testsupport.NewServer(t, func(writer http.ResponseWriter, request *http.Request) {
		body := assertActivityRequest(t, request, http.MethodGet, "/api/v1/workspaces/team%2Fslash/projects/project%2Fslash/work-items/work%2Fitem/activities/activity%2Fitem/", query, plane.AuthModeAPIKey)
		if len(body) != 0 {
			t.Errorf("detail body = %q, want empty", body)
		}
		testsupport.JSON(writer, http.StatusOK, activityPageFixture())
	})
	shared, err := plane.NewClient(plane.Options{BaseURL: server.URL, AuthMode: plane.AuthModeAPIKey, APIKey: "secret-key"})
	if err != nil {
		t.Fatal(err)
	}
	client := NewClient(shared)
	result, response, err := client.Get(context.Background(), testWorkspace, testProject, testWorkItem, testActivity, DetailOptions{
		Cursor: "next", PerPage: 10, Fields: "id", Expand: "actor", OrderBy: "created_at",
	})
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != http.StatusOK || len(result.Results) != 1 || result.Results[0].Epoch == nil || *result.Results[0].Epoch != 1704067200.5 {
		t.Fatalf("detail result = %#v response=%#v", result, response)
	}
}

func TestLegacyListUsesCompatibilityRouteIndependently(t *testing.T) {
	query := url.Values{"fields": {"id,actor"}, "order_by": {"-epoch"}, "per_page": {"5"}}
	server := testsupport.NewServer(t, func(writer http.ResponseWriter, request *http.Request) {
		body := assertActivityRequest(t, request, http.MethodGet, "/api/v1/workspaces/team%2Fslash/projects/project%2Fslash/issues/issue%2Fitem/activities/", query, plane.AuthModeAPIKey)
		if len(body) != 0 {
			t.Errorf("legacy list body = %q, want empty", body)
		}
		testsupport.JSON(writer, http.StatusOK, activityPageFixture())
	})
	shared, err := plane.NewClient(plane.Options{BaseURL: server.URL, AuthMode: plane.AuthModeAPIKey, APIKey: "secret-key"})
	if err != nil {
		t.Fatal(err)
	}
	client := NewClient(shared)
	page, _, err := client.LegacyList(context.Background(), testWorkspace, testProject, testIssue, ListOptions{Fields: "id,actor", PerPage: 5, OrderBy: "-epoch"})
	if err != nil || len(page.Results) != 1 {
		t.Fatalf("legacy list result=%#v err=%v", page, err)
	}
}

func TestLegacyGetUsesCompatibilityDetailRouteIndependently(t *testing.T) {
	query := url.Values{"cursor": {"legacy"}, "expand": {"project"}, "per_page": {"1"}}
	server := testsupport.NewServer(t, func(writer http.ResponseWriter, request *http.Request) {
		body := assertActivityRequest(t, request, http.MethodGet, "/api/v1/workspaces/team%2Fslash/projects/project%2Fslash/issues/issue%2Fitem/activities/activity%2Fitem/", query, plane.AuthModeAPIKey)
		if len(body) != 0 {
			t.Errorf("legacy detail body = %q, want empty", body)
		}
		testsupport.JSON(writer, http.StatusOK, activityPageFixture())
	})
	shared, err := plane.NewClient(plane.Options{BaseURL: server.URL, AuthMode: plane.AuthModeAPIKey, APIKey: "secret-key"})
	if err != nil {
		t.Fatal(err)
	}
	client := NewClient(shared)
	result, _, err := client.LegacyGet(context.Background(), testWorkspace, testProject, testIssue, testActivity, DetailOptions{Cursor: "legacy", Expand: "project", PerPage: 1})
	if err != nil || len(result.Results) != 1 || string(result.Results[0].IssueComment) != "null" {
		t.Fatalf("legacy detail result=%#v err=%v", result, err)
	}
}

func TestActivitySupportsBearerAuthentication(t *testing.T) {
	query := url.Values{}
	server := testsupport.NewServer(t, func(writer http.ResponseWriter, request *http.Request) {
		body := assertActivityRequest(t, request, http.MethodGet, "/api/v1/workspaces/team%2Fslash/projects/project%2Fslash/work-items/work%2Fitem/activities/", query, plane.AuthModeBearer)
		if len(body) != 0 {
			t.Errorf("bearer body = %q, want empty", body)
		}
		testsupport.JSON(writer, http.StatusOK, activityPageFixture())
	})
	shared, err := plane.NewClient(plane.Options{BaseURL: server.URL, AuthMode: plane.AuthModeBearer, BearerToken: "bearer-token"})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := NewClient(shared).List(context.Background(), testWorkspace, testProject, testWorkItem, ListOptions{}); err != nil {
		t.Fatal(err)
	}
}

func TestActivityHandlesSharedResponseFailuresAndValidation(t *testing.T) {
	if _, _, err := NewClient(nil).List(context.Background(), "team", "project", "work-item", ListOptions{}); err == nil || !strings.Contains(err.Error(), "not initialized") {
		t.Fatalf("nil client error = %v", err)
	}
	if _, err := (ListOptions{PerPage: 101}).Query(); err == nil {
		t.Fatal("invalid page size was accepted")
	}

	cases := []struct {
		name     string
		handler  http.HandlerFunc
		wantText string
	}{
		{name: "malformed", handler: testsupport.MalformedJSONHandler(http.StatusOK), wantText: "malformed JSON"},
		{name: "non-json", handler: testsupport.NonJSONHandler(http.StatusOK, "plain response"), wantText: "malformed JSON"},
		{name: "api-error", handler: testsupport.JSONHandler(http.StatusForbidden, map[string]any{"error": "forbidden", "message": "denied"}), wantText: "plane API error"},
		{name: "empty-success", handler: testsupport.EmptyHandler(http.StatusOK), wantText: ""},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			server := testsupport.NewServer(t, testCase.handler)
			shared, err := plane.NewClient(plane.Options{BaseURL: server.URL, AuthMode: plane.AuthModeAPIKey, APIKey: "secret-key"})
			if err != nil {
				t.Fatal(err)
			}
			_, _, err = NewClient(shared).List(context.Background(), "team", "project", "work-item", ListOptions{})
			if testCase.wantText == "" {
				if err != nil {
					t.Fatal(err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), testCase.wantText) {
				t.Fatalf("error = %v, want text %q", err, testCase.wantText)
			}
		})
	}

	server := testsupport.NewServer(t, testsupport.JSONHandler(http.StatusOK, activityPageFixture()))
	shared, err := plane.NewClient(plane.Options{BaseURL: server.URL, AuthMode: plane.AuthModeAPIKey, APIKey: "secret-key"})
	if err != nil {
		t.Fatal(err)
	}
	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	if _, _, err := NewClient(shared).List(canceled, "team", "project", "work-item", ListOptions{}); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled request error = %v", err)
	}
}

func TestActivityRejectsUnexpectedStatus(t *testing.T) {
	client := NewClient(statusClient{status: http.StatusCreated})
	if _, _, err := client.List(context.Background(), "team", "project", "work-item", ListOptions{}); err == nil || !strings.Contains(err.Error(), "want 200") {
		t.Fatalf("status mismatch error = %v", err)
	}
}

type statusClient struct{ status int }

func (c statusClient) Do(context.Context, string, string, url.Values, any, http.Header, any) (plane.Response, error) {
	return plane.Response{StatusCode: c.status}, nil
}
