package links

import (
	"context"
	"encoding/json"
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
	testLink      = "link/item"
)

func newHTTPClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()
	server := testsupport.NewServer(t, handler)
	shared, err := plane.NewClient(plane.Options{
		BaseURL: server.URL, AuthMode: plane.AuthModeAPIKey, APIKey: "secret-key",
	})
	if err != nil {
		t.Fatal(err)
	}
	return NewClient(shared)
}

func assertCommonRequest(t *testing.T, request *http.Request, method, escapedPath string, query url.Values) []byte {
	t.Helper()
	if request.Method != method {
		t.Errorf("method = %q, want %q", request.Method, method)
	}
	if request.URL.EscapedPath() != escapedPath {
		t.Errorf("escaped path = %q, want %q", request.URL.EscapedPath(), escapedPath)
	}
	if !equalValues(request.URL.Query(), query) {
		t.Errorf("query = %v, want %v", request.URL.Query(), query)
	}
	if got := request.Header.Get("X-API-Key"); got != "secret-key" {
		t.Errorf("X-API-Key = %q, want secret-key", got)
	}
	if got := request.Header.Get("Authorization"); got != "" {
		t.Errorf("Authorization = %q, want absent", got)
	}
	body, err := io.ReadAll(request.Body)
	if err != nil {
		t.Fatal(err)
	}
	return body
}

func equalValues(left, right url.Values) bool {
	return left.Encode() == right.Encode()
}

func TestListUsesPrimaryRouteQueryAuthAndPageMetadata(t *testing.T) {
	query := url.Values{"cursor": {"20:1:0"}, "expand": {"project"}, "fields": {"id,url"}, "order_by": {"-created_at"}, "per_page": {"20"}}
	client := newHTTPClient(t, func(writer http.ResponseWriter, request *http.Request) {
		body := assertCommonRequest(t, request, http.MethodGet, "/api/v1/workspaces/team%2Fslash/projects/project%2Fslash/work-items/work%2Fitem/links/", query)
		if len(body) != 0 {
			t.Errorf("list body = %q, want empty", body)
		}
		writer.Header().Set("X-RateLimit-Remaining", "7")
		writer.Header().Set("X-RateLimit-Reset", "1234")
		testsupport.JSON(writer, http.StatusOK, map[string]any{
			"next_cursor": "next", "count": 1, "total_results": 1,
			"extra_stats": nil, "grouped_by": "state", "future": true,
			"results": []map[string]any{{"id": "link-1", "url": "https://example.test"}},
		})
	})
	page, response, err := client.List(context.Background(), testWorkspace, testProject, testWorkItem, ListOptions{
		Cursor: "20:1:0", PerPage: 20, Fields: "id,url", Expand: "project", OrderBy: "-created_at",
	})
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != http.StatusOK || response.RateLimit.Remaining != 7 || response.RateLimit.Reset != 1234 {
		t.Fatalf("response metadata = %#v", response)
	}
	if len(page.Results) != 1 || page.Results[0].URL == nil || *page.Results[0].URL != "https://example.test" {
		t.Fatalf("page = %#v", page)
	}
	if string(page.Unknown["future"]) != "true" {
		t.Fatalf("unknown page field = %s", page.Unknown["future"])
	}
}

func TestCreateUsesPrimaryRouteBodyAnd201(t *testing.T) {
	title := "Example"
	client := newHTTPClient(t, func(writer http.ResponseWriter, request *http.Request) {
		body := assertCommonRequest(t, request, http.MethodPost, "/api/v1/workspaces/team%2Fslash/projects/project%2Fslash/work-items/work%2Fitem/links/", nil)
		if request.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Content-Type = %q", request.Header.Get("Content-Type"))
		}
		if string(body) != `{"url":"https://example.test","title":"Example"}` {
			t.Errorf("body = %s", body)
		}
		testsupport.JSON(writer, http.StatusCreated, map[string]any{"id": "link-1", "url": "https://example.test", "title": "Example"})
	})
	link, response, err := client.Create(context.Background(), testWorkspace, testProject, testWorkItem, CreateLinkRequest{URL: "https://example.test", Title: &title})
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != http.StatusCreated || link.ID == nil || *link.ID != "link-1" {
		t.Fatalf("create result = %#v / %#v", link, response)
	}
}

func TestGetUsesPrimaryDetailPageAndNeverSendsOrderBy(t *testing.T) {
	query := url.Values{"cursor": {"next"}, "expand": {"workspace"}, "fields": {"id,url"}, "per_page": {"20"}}
	client := newHTTPClient(t, func(writer http.ResponseWriter, request *http.Request) {
		body := assertCommonRequest(t, request, http.MethodGet, "/api/v1/workspaces/team%2Fslash/projects/project%2Fslash/work-items/work%2Fitem/links/link%2Fitem/", query)
		if len(body) != 0 {
			t.Errorf("get body = %q, want empty", body)
		}
		testsupport.JSON(writer, http.StatusOK, map[string]any{
			"count": 1, "extra_stats": nil, "results": []map[string]any{{"id": "link-1", "metadata": []any{"dynamic"}}},
		})
	})
	result, response, err := client.Get(context.Background(), testWorkspace, testProject, testWorkItem, testLink, DetailOptions{Cursor: "next", PerPage: 20, Fields: "id,url", Expand: "workspace"})
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != http.StatusOK || len(result.Results) != 1 || string(result.Results[0].Metadata) != `["dynamic"]` {
		t.Fatalf("detail result = %#v / %#v", result, response)
	}
}

func TestUpdateUsesPrimaryRouteAndExplicitEmptyBodyValue(t *testing.T) {
	title := ""
	client := newHTTPClient(t, func(writer http.ResponseWriter, request *http.Request) {
		body := assertCommonRequest(t, request, http.MethodPatch, "/api/v1/workspaces/team%2Fslash/projects/project%2Fslash/work-items/work%2Fitem/links/link%2Fitem/", nil)
		if request.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Content-Type = %q", request.Header.Get("Content-Type"))
		}
		if string(body) != `{"title":""}` {
			t.Errorf("body = %s", body)
		}
		testsupport.JSON(writer, http.StatusOK, map[string]any{"id": "link-1", "title": ""})
	})
	link, response, err := client.Update(context.Background(), testWorkspace, testProject, testWorkItem, testLink, UpdateLinkRequest{Title: &title})
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != http.StatusOK || link.Title == nil || *link.Title != "" {
		t.Fatalf("update result = %#v / %#v", link, response)
	}
}

func TestDeleteUsesPrimaryRouteAnd204IsBodyless(t *testing.T) {
	client := newHTTPClient(t, func(writer http.ResponseWriter, request *http.Request) {
		body := assertCommonRequest(t, request, http.MethodDelete, "/api/v1/workspaces/team%2Fslash/projects/project%2Fslash/work-items/work%2Fitem/links/link%2Fitem/", nil)
		if len(body) != 0 {
			t.Errorf("delete body = %q, want empty", body)
		}
		testsupport.NoContent(writer)
	})
	response, err := client.Delete(context.Background(), testWorkspace, testProject, testWorkItem, testLink)
	if err != nil || response.StatusCode != http.StatusNoContent {
		t.Fatalf("delete response = %#v, err=%v", response, err)
	}
}

func TestCompatibilityMethodsUseOnlyIssuesRoutes(t *testing.T) {
	tests := []struct {
		name   string
		method string
		path   string
		call   func(*Client) error
	}{
		{name: "list", method: http.MethodGet, path: "/api/v1/workspaces/team%2Fslash/projects/project%2Fslash/issues/issue%2Fitem/links/", call: func(client *Client) error {
			_, _, err := client.LegacyList(context.Background(), testWorkspace, testProject, testIssue, ListOptions{})
			return err
		}},
		{name: "create", method: http.MethodPost, path: "/api/v1/workspaces/team%2Fslash/projects/project%2Fslash/issues/issue%2Fitem/links/", call: func(client *Client) error {
			_, _, err := client.LegacyCreate(context.Background(), testWorkspace, testProject, testIssue, CreateLinkRequest{URL: "https://example.test"})
			return err
		}},
		{name: "get", method: http.MethodGet, path: "/api/v1/workspaces/team%2Fslash/projects/project%2Fslash/issues/issue%2Fitem/links/link%2Fitem/", call: func(client *Client) error {
			_, _, err := client.LegacyGet(context.Background(), testWorkspace, testProject, testIssue, testLink)
			return err
		}},
		{name: "update", method: http.MethodPatch, path: "/api/v1/workspaces/team%2Fslash/projects/project%2Fslash/issues/issue%2Fitem/links/link%2Fitem/", call: func(client *Client) error {
			_, _, err := client.LegacyUpdate(context.Background(), testWorkspace, testProject, testIssue, testLink, UpdateLinkRequest{URL: stringPointer("https://example.test/updated")})
			return err
		}},
		{name: "delete", method: http.MethodDelete, path: "/api/v1/workspaces/team%2Fslash/projects/project%2Fslash/issues/issue%2Fitem/links/link%2Fitem/", call: func(client *Client) error {
			_, err := client.LegacyDelete(context.Background(), testWorkspace, testProject, testIssue, testLink)
			return err
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := newHTTPClient(t, func(writer http.ResponseWriter, request *http.Request) {
				body := assertCommonRequest(t, request, test.method, test.path, nil)
				if test.method == http.MethodPost || test.method == http.MethodPatch {
					if request.Header.Get("Content-Type") != "application/json" {
						t.Errorf("Content-Type = %q", request.Header.Get("Content-Type"))
					}
				}
				if test.method == http.MethodDelete {
					testsupport.NoContent(writer)
					return
				}
				if test.method == http.MethodPost {
					if string(body) != `{"url":"https://example.test"}` {
						t.Errorf("create body = %s", body)
					}
					testsupport.JSON(writer, http.StatusCreated, map[string]any{"id": "legacy-link"})
					return
				}
				if test.method == http.MethodPatch {
					if string(body) != `{"url":"https://example.test/updated"}` {
						t.Errorf("update body = %s", body)
					}
					testsupport.JSON(writer, http.StatusOK, map[string]any{"id": "legacy-link"})
					return
				}
				if test.name == "get" {
					testsupport.JSON(writer, http.StatusOK, map[string]any{"results": []map[string]any{{"id": "legacy-link"}}})
					return
				}
				testsupport.JSON(writer, http.StatusOK, map[string]any{"results": []any{}})
			})
			if err := test.call(client); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestLinkClientValidatesBeforeRequestAndChecksStatus(t *testing.T) {
	var calls int
	client := newHTTPClient(t, func(writer http.ResponseWriter, request *http.Request) {
		calls++
		testsupport.JSON(writer, http.StatusAccepted, map[string]any{"id": "wrong-status"})
	})
	if _, _, err := client.Create(context.Background(), "team", "project", "work-item", CreateLinkRequest{}); err == nil {
		t.Fatal("blank URL was accepted")
	}
	if calls != 0 {
		t.Fatal("validation issued a request")
	}
	if _, _, err := client.Create(context.Background(), "team", "project", "work-item", CreateLinkRequest{URL: "https://example.test"}); err == nil || !strings.Contains(err.Error(), "want 201") {
		t.Fatalf("status mismatch error = %v", err)
	}
	if calls != 1 {
		t.Fatalf("request count = %d, want 1", calls)
	}
}

func TestLinkClientPropagatesMalformedAndStructuredErrors(t *testing.T) {
	malformed := newHTTPClient(t, testsupport.MalformedJSONHandler(http.StatusOK))
	if _, _, err := malformed.List(context.Background(), "team", "project", "work-item", ListOptions{}); err == nil {
		t.Fatal("malformed success body was accepted")
	}
	structured := newHTTPClient(t, func(writer http.ResponseWriter, request *http.Request) {
		testsupport.JSON(writer, http.StatusBadRequest, map[string]string{"message": "invalid link"})
	})
	_, _, err := structured.List(context.Background(), "team", "project", "work-item", ListOptions{})
	var apiError *plane.APIError
	if !errors.As(err, &apiError) || apiError.StatusCode != http.StatusBadRequest || apiError.Message != "invalid link" {
		t.Fatalf("structured error = %v (%T)", err, err)
	}
}

func TestLinkClientCancellationDoesNotReachServer(t *testing.T) {
	var calls int
	client := newHTTPClient(t, func(writer http.ResponseWriter, request *http.Request) { calls++ })
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, _, err := client.List(ctx, "team", "project", "work-item", ListOptions{})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation error = %v", err)
	}
	if calls != 0 {
		t.Fatalf("server calls = %d, want 0", calls)
	}
}

func stringPointer(value string) *string { return &value }

func TestLinkRequestsAreJSONRoundTrips(t *testing.T) {
	data, err := json.Marshal(UpdateLinkRequest{Title: stringPointer("title")})
	if err != nil || !strings.Contains(string(data), `"title":"title"`) {
		t.Fatalf("request JSON = %s, err=%v", data, err)
	}
}
