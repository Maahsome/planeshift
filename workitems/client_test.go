package workitems

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"planeshift/internal/testsupport"
	"planeshift/plane"
)

const (
	testWorkspace = "team space"
	testProject   = "project/id"
	testWorkItem  = "work/item"
)

func testClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()
	server := testsupport.NewServer(t, handler)
	shared, err := plane.NewClient(plane.Options{BaseURL: server.URL, AuthMode: plane.AuthModeAPIKey, APIKey: "test-key"})
	if err != nil {
		t.Fatal(err)
	}
	return NewClient(shared)
}

func TestClientOperationMatrixUsesExactCurrentAndLegacyRoutes(t *testing.T) {
	create := CreateWorkItemRequest{Name: "New item"}
	update := UpdateWorkItemRequest{}
	update.SetNull("parent")
	relation := CreateWorkItemRelationRequest{RelationType: RelationRelatesTo, Issues: []string{"related-id"}}
	pageJSON := `{"next_cursor":"next","count":1,"results":[{"id":"item-1","name":"Item","future":true}]}`
	itemJSON := `{"id":"item-1","name":"Item","parent":null,"expanded":{"id":"state-1"}}`
	searchJSON := `{"issues":[{"id":"item-1","name":"Item","sequence_id":1,"future":{"keep":true}}]}`
	relationPageJSON := `{"next_cursor":"next","results":[{"id":"relation-1","relation_type":"relates_to","future":"keep"}]}`
	relationCreateJSON := `[[{"id":"relation-1","relation_type":"relates_to","future":"keep"}]]`

	tests := []struct {
		name     string
		method   string
		path     string
		query    url.Values
		body     any
		response string
		status   int
		call     func(context.Context, *Client) error
	}{
		{name: "search", method: http.MethodGet, path: "/api/v1/workspaces/team%20space/work-items/search/", query: url.Values{"search": {"item"}, "limit": {"10"}, "project_id": {"project"}, "workspace_search": {"true"}}, response: searchJSON, status: http.StatusOK,
			call: func(ctx context.Context, c *Client) error {
				_, _, err := c.Search(ctx, testWorkspace, SearchOptions{Search: "item", Limit: 10, ProjectID: "project", WorkspaceSearch: "true"})
				return err
			}},
		{name: "identifier", method: http.MethodGet, path: "/api/v1/workspaces/team%20space/work-items/PROJ%2FKEY-1%2F2/", query: url.Values{"expand": {"state"}}, response: itemJSON, status: http.StatusOK,
			call: func(ctx context.Context, c *Client) error {
				_, _, err := c.GetByIdentifier(ctx, testWorkspace, "PROJ/KEY", "1/2", IdentifierOptions{Expand: "state"})
				return err
			}},
		{name: "list", method: http.MethodGet, path: "/api/v1/workspaces/team%20space/projects/project%2Fid/work-items/", query: url.Values{"cursor": {"next:1"}, "per_page": {"20"}, "fields": {"id,name"}, "expand": {"state"}, "external_id": {"ext"}, "external_source": {"source"}, "order_by": {"-created_at"}}, response: pageJSON, status: http.StatusOK,
			call: func(ctx context.Context, c *Client) error {
				page, response, err := c.List(ctx, testWorkspace, testProject, ListOptions{Cursor: "next:1", PerPage: 20, Fields: "id,name", Expand: "state", ExternalID: "ext", ExternalSource: "source", OrderBy: "-created_at"})
				if err == nil && (page.NextCursor == nil || *page.NextCursor != "next" || response.RateLimit.Remaining != 42 || !response.RateLimit.RemainingParsed) {
					return fmt.Errorf("page/metadata = %#v/%#v", page, response.Metadata)
				}
				return err
			}},
		{name: "create", method: http.MethodPost, path: "/api/v1/workspaces/team%20space/projects/project%2Fid/work-items/", body: create, response: itemJSON, status: http.StatusCreated,
			call: func(ctx context.Context, c *Client) error {
				_, _, err := c.Create(ctx, testWorkspace, testProject, create)
				return err
			}},
		{name: "get", method: http.MethodGet, path: "/api/v1/workspaces/team%20space/projects/project%2Fid/work-items/work%2Fitem/", query: url.Values{"expand": {"state"}}, response: itemJSON, status: http.StatusOK,
			call: func(ctx context.Context, c *Client) error {
				_, _, err := c.Get(ctx, testWorkspace, testProject, testWorkItem, DetailOptions{Expand: "state"})
				return err
			}},
		{name: "update", method: http.MethodPatch, path: "/api/v1/workspaces/team%20space/projects/project%2Fid/work-items/work%2Fitem/", body: update, response: itemJSON, status: http.StatusOK,
			call: func(ctx context.Context, c *Client) error {
				_, _, err := c.Update(ctx, testWorkspace, testProject, testWorkItem, update)
				return err
			}},
		{name: "delete", method: http.MethodDelete, path: "/api/v1/workspaces/team%20space/projects/project%2Fid/work-items/work%2Fitem/", response: "", status: http.StatusNoContent,
			call: func(ctx context.Context, c *Client) error {
				_, err := c.Delete(ctx, testWorkspace, testProject, testWorkItem)
				return err
			}},
		{name: "relations list", method: http.MethodGet, path: "/api/v1/workspaces/team%20space/projects/project%2Fid/work-items/work%2Fitem/relations/", query: url.Values{"cursor": {"next"}, "per_page": {"10"}, "fields": {"id"}, "expand": {"state"}}, response: relationPageJSON, status: http.StatusOK,
			call: func(ctx context.Context, c *Client) error {
				_, _, err := c.ListRelations(ctx, testWorkspace, testProject, testWorkItem, RelationListOptions{Cursor: "next", PerPage: 10, Fields: "id", Expand: "state"})
				return err
			}},
		{name: "relations create", method: http.MethodPost, path: "/api/v1/workspaces/team%20space/projects/project%2Fid/work-items/work%2Fitem/relations/", body: relation, response: relationCreateJSON, status: http.StatusCreated,
			call: func(ctx context.Context, c *Client) error {
				_, _, err := c.CreateRelation(ctx, testWorkspace, testProject, testWorkItem, relation)
				return err
			}},
		{name: "legacy search", method: http.MethodGet, path: "/api/v1/workspaces/team%20space/issues/search/", query: url.Values{"search": {"item"}}, response: searchJSON, status: http.StatusOK,
			call: func(ctx context.Context, c *Client) error {
				_, _, err := c.LegacySearch(ctx, testWorkspace, SearchOptions{Search: "item"})
				return err
			}},
		{name: "legacy identifier", method: http.MethodGet, path: "/api/v1/workspaces/team%20space/issues/PROJ-12/", response: itemJSON, status: http.StatusOK,
			call: func(ctx context.Context, c *Client) error {
				_, _, err := c.LegacyGetByIdentifier(ctx, testWorkspace, "PROJ", "12")
				return err
			}},
		{name: "legacy list", method: http.MethodGet, path: "/api/v1/workspaces/team%20space/projects/project%2Fid/issues/", response: pageJSON, status: http.StatusOK,
			call: func(ctx context.Context, c *Client) error {
				_, _, err := c.LegacyList(ctx, testWorkspace, testProject, ListOptions{})
				return err
			}},
		{name: "legacy create", method: http.MethodPost, path: "/api/v1/workspaces/team%20space/projects/project%2Fid/issues/", body: create, response: itemJSON, status: http.StatusCreated,
			call: func(ctx context.Context, c *Client) error {
				_, _, err := c.LegacyCreate(ctx, testWorkspace, testProject, create)
				return err
			}},
		{name: "legacy get", method: http.MethodGet, path: "/api/v1/workspaces/team%20space/projects/project%2Fid/issues/work%2Fitem/", response: itemJSON, status: http.StatusOK,
			call: func(ctx context.Context, c *Client) error {
				_, _, err := c.LegacyGet(ctx, testWorkspace, testProject, testWorkItem)
				return err
			}},
		{name: "legacy update", method: http.MethodPatch, path: "/api/v1/workspaces/team%20space/projects/project%2Fid/issues/work%2Fitem/", body: update, response: itemJSON, status: http.StatusOK,
			call: func(ctx context.Context, c *Client) error {
				_, _, err := c.LegacyUpdate(ctx, testWorkspace, testProject, testWorkItem, update)
				return err
			}},
		{name: "legacy delete", method: http.MethodDelete, path: "/api/v1/workspaces/team%20space/projects/project%2Fid/issues/work%2Fitem/", response: "", status: http.StatusNoContent,
			call: func(ctx context.Context, c *Client) error {
				_, err := c.LegacyDelete(ctx, testWorkspace, testProject, testWorkItem)
				return err
			}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := testClient(t, func(writer http.ResponseWriter, request *http.Request) {
				if request.Method != tt.method {
					t.Errorf("method = %q, want %q", request.Method, tt.method)
				}
				if request.URL.EscapedPath() != tt.path {
					t.Errorf("path = %q, want %q", request.URL.EscapedPath(), tt.path)
				}
				if tt.query != nil && request.URL.Query().Encode() != tt.query.Encode() {
					t.Errorf("query = %v, want %v", request.URL.Query(), tt.query)
				}
				if request.Header.Get("X-API-Key") != "test-key" || request.Header.Get("Authorization") != "" {
					t.Errorf("auth headers = %q/%q", request.Header.Get("X-API-Key"), request.Header.Get("Authorization"))
				}
				if request.Header.Get("Accept") != "application/json" {
					t.Errorf("Accept = %q", request.Header.Get("Accept"))
				}
				if tt.name == "list" {
					writer.Header().Set("X-RateLimit-Remaining", "42")
					writer.Header().Set("X-RateLimit-Reset", "123")
				}
				body, _ := io.ReadAll(request.Body)
				if tt.body == nil {
					if len(body) != 0 || request.Header.Get("Content-Type") != "" {
						t.Errorf("empty request body/content type = %q/%q", body, request.Header.Get("Content-Type"))
					}
				} else {
					testsupport.JSONBodyEqual(t, body, tt.body)
					if request.Header.Get("Content-Type") != "application/json" {
						t.Errorf("Content-Type = %q", request.Header.Get("Content-Type"))
					}
				}
				if tt.status == http.StatusNoContent {
					testsupport.NoContent(writer)
					return
				}
				testsupport.JSON(writer, tt.status, json.RawMessage(tt.response))
			})
			if err := tt.call(context.Background(), client); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestClientSupportsBearerWithoutAPIKey(t *testing.T) {
	server := testsupport.NewServer(t, func(writer http.ResponseWriter, request *http.Request) {
		if request.Header.Get("Authorization") != "Bearer bearer-token" {
			t.Errorf("Authorization = %q", request.Header.Get("Authorization"))
		}
		if request.Header.Get("X-API-Key") != "" {
			t.Errorf("X-API-Key = %q, want absent", request.Header.Get("X-API-Key"))
		}
		testsupport.JSON(writer, http.StatusOK, map[string]any{"id": "item"})
	})
	shared, err := plane.NewClient(plane.Options{BaseURL: server.URL, AuthMode: plane.AuthModeBearer, BearerToken: "bearer-token"})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := NewClient(shared).Get(context.Background(), "team", "project", "item"); err != nil {
		t.Fatal(err)
	}
}

func TestClientErrorsAndValidation(t *testing.T) {
	client := testClient(t, func(writer http.ResponseWriter, request *http.Request) {
		testsupport.NonJSON(writer, http.StatusBadRequest, "bad request")
	})
	if _, _, err := client.Create(context.Background(), "team", "project", CreateWorkItemRequest{}); err == nil || !strings.Contains(err.Error(), "name is required") {
		t.Fatalf("missing name error = %v", err)
	}
	if _, _, err := client.Search(context.Background(), "team", SearchOptions{}); err == nil || !strings.Contains(err.Error(), "search is required") {
		t.Fatalf("missing search error = %v", err)
	}
	if _, _, err := client.CreateRelation(context.Background(), "team", "project", "item", CreateWorkItemRelationRequest{RelationType: "invalid", Issues: []string{"id"}}); err == nil {
		t.Fatal("invalid relation type was accepted")
	}
	if _, _, err := client.Search(context.Background(), "team", SearchOptions{Search: "x"}); err == nil || !strings.Contains(err.Error(), "bad request") {
		t.Fatalf("API error = %v", err)
	}
}
