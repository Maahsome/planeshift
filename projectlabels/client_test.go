package projectlabels

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"planeshift/internal/testsupport"
	"planeshift/plane"
)

func TestClientOperationMatrix(t *testing.T) {
	empty := ""
	zero := float64(0)
	createRequest := CreateProjectLabelRequest{Name: "Label", Description: &empty, Color: &empty, SortOrder: &zero}
	updateRequest := UpdateProjectLabelRequest{Name: &empty, Description: &empty, Color: &empty, SortOrder: &zero}

	tests := []struct {
		name            string
		method          string
		wantPath        string
		wantEscapedPath string
		wantQuery       url.Values
		body            any
		wantStatus      int
		call            func(context.Context, *Client) (plane.Response, error)
	}{
		{
			name: "create with api key", method: http.MethodPost,
			wantPath: "/api/v1/workspaces/team space/project-labels/", wantEscapedPath: "/api/v1/workspaces/team%20space/project-labels/", body: createRequest,
			call: func(ctx context.Context, client *Client) (plane.Response, error) {
				_, response, err := client.Create(ctx, "team space", createRequest)
				return response, err
			},
			wantStatus: http.StatusCreated,
		},
		{
			name: "list with bearer", method: http.MethodGet,
			wantPath: "/api/v1/workspaces/team space/project-labels/", wantEscapedPath: "/api/v1/workspaces/team%20space/project-labels/",
			wantQuery: url.Values{
				"cursor": {"next:1"}, "expand": {"projects"}, "fields": {"id,name"},
				"order_by": {"-sort_order"}, "per_page": {"20"},
			},
			call: func(ctx context.Context, client *Client) (plane.Response, error) {
				_, response, err := client.List(ctx, "team space", ListOptions{
					Cursor: "next:1", PerPage: 20, Fields: "id,name", Expand: "projects", OrderBy: "-sort_order",
				})
				return response, err
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "get escaped label", method: http.MethodGet,
			wantPath: "/api/v1/workspaces/team space/project-labels/label/id/", wantEscapedPath: "/api/v1/workspaces/team%20space/project-labels/label%2Fid/",
			call: func(ctx context.Context, client *Client) (plane.Response, error) {
				_, response, err := client.Get(ctx, "team space", "label/id")
				return response, err
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "update", method: http.MethodPatch,
			wantPath: "/api/v1/workspaces/team space/project-labels/label/id/", wantEscapedPath: "/api/v1/workspaces/team%20space/project-labels/label%2Fid/", body: updateRequest,
			call: func(ctx context.Context, client *Client) (plane.Response, error) {
				_, response, err := client.Update(ctx, "team space", "label/id", updateRequest)
				return response, err
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "delete", method: http.MethodDelete,
			wantPath: "/api/v1/workspaces/team space/project-labels/label/id/", wantEscapedPath: "/api/v1/workspaces/team%20space/project-labels/label%2Fid/",
			call: func(ctx context.Context, client *Client) (plane.Response, error) {
				return client.Delete(ctx, "team space", "label/id")
			},
			wantStatus: http.StatusNoContent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := testsupport.NewServer(t, func(writer http.ResponseWriter, request *http.Request) {
				wantAuth := http.Header{"X-API-Key": {"api-key"}}
				absentAuth := []string{"Authorization"}
				if tt.name == "list with bearer" {
					wantAuth = http.Header{"Authorization": {"Bearer bearer-token"}}
					absentAuth = []string{"X-API-Key"}
				}
				expectedHeaders := http.Header{"Accept": {"application/json"}}
				if tt.body != nil {
					expectedHeaders.Set("Content-Type", "application/json")
				}
				for key, values := range wantAuth {
					expectedHeaders[key] = values
				}
				body := testsupport.AssertRequest(t, request, testsupport.RequestExpectation{
					Method: tt.method, Path: tt.wantPath, Query: queryOrEmpty(tt.wantQuery),
					ExpectedHeaders: expectedHeaders, AbsentHeaders: absentAuth,
				})
				if request.URL.EscapedPath() != tt.wantEscapedPath {
					t.Errorf("escaped path = %q, want %q", request.URL.EscapedPath(), tt.wantEscapedPath)
				}
				if tt.body == nil {
					if len(body) != 0 || request.Header.Get("Content-Type") != "" {
						t.Errorf("no-body request = %q/%q", body, request.Header.Get("Content-Type"))
					}
				} else {
					testsupport.JSONBodyEqual(t, body, tt.body)
				}
				writer.Header().Set("X-RateLimit-Remaining", "42")
				writer.Header().Set("X-RateLimit-Reset", "1700327957")
				switch tt.method {
				case http.MethodDelete:
					testsupport.NoContent(writer)
				case http.MethodGet:
					if tt.wantQuery != nil {
						testsupport.JSON(writer, http.StatusOK, map[string]any{
							"grouped_by": "state", "sub_grouped_by": "priority", "total_count": 1,
							"next_cursor": "next", "extra_stats": nil, "results": []any{map[string]any{
								"id": "label-1", "name": "Label", "sort_order": 2, "future": nil,
							}}, "future_page": true,
						})
					} else {
						testsupport.JSON(writer, http.StatusOK, map[string]any{
							"id": "label-1", "name": "Label", "sort_order": 2, "workspace": "workspace-1", "future": nil,
						})
					}
				case http.MethodPost:
					testsupport.JSON(writer, http.StatusCreated, map[string]any{"id": "label-1", "name": "Label"})
				case http.MethodPatch:
					testsupport.JSON(writer, http.StatusOK, map[string]any{"id": "label-1", "name": "Label"})
				}
			})

			options := plane.Options{BaseURL: server.URL}
			if tt.name == "list with bearer" {
				options.AuthMode = plane.AuthModeBearer
				options.BearerToken = "bearer-token"
			} else {
				options.AuthMode = plane.AuthModeAPIKey
				options.APIKey = "api-key"
			}
			shared, err := plane.NewClient(options)
			if err != nil {
				t.Fatal(err)
			}
			response, err := tt.call(context.Background(), NewClient(shared))
			if err != nil {
				t.Fatal(err)
			}
			if response.StatusCode != tt.wantStatus || response.RateLimit.Remaining != 42 || response.RateLimit.Reset != 1700327957 {
				t.Fatalf("response metadata = %#v", response)
			}
		})
	}
}

func TestClientDecodesTypedCreateListAndGetResponses(t *testing.T) {
	server := testsupport.NewServer(t, func(writer http.ResponseWriter, request *http.Request) {
		switch request.Method {
		case http.MethodPost, http.MethodGet:
			if request.Method == http.MethodGet && strings.HasSuffix(request.URL.Path, "/label/") {
				testsupport.JSON(writer, http.StatusOK, map[string]any{"id": "label-1", "name": "Label", "sort_order": 3})
				return
			}
			if request.Method == http.MethodGet {
				testsupport.JSON(writer, http.StatusOK, map[string]any{
					"total_count": 1, "extra_stats": nil, "results": []any{map[string]any{"id": "label-1", "name": "Label"}}, "future_page": "kept",
				})
				return
			}
			testsupport.JSON(writer, http.StatusCreated, map[string]any{"id": "label-1", "name": "Label", "workspace": "team"})
		}
	})
	shared, err := plane.NewClient(plane.Options{BaseURL: server.URL, AuthMode: plane.AuthModeAPIKey, APIKey: "key"})
	if err != nil {
		t.Fatal(err)
	}
	client := NewClient(shared)
	created, response, err := client.Create(context.Background(), "team", CreateProjectLabelRequest{Name: "Label"})
	if err != nil || response.StatusCode != http.StatusCreated || created.ID != "label-1" || created.Workspace == nil || *created.Workspace != "team" {
		t.Fatalf("created label/response/error = %#v/%#v/%v", created, response, err)
	}
	page, response, err := client.List(context.Background(), "team", ListOptions{})
	if err != nil || response.StatusCode != http.StatusOK || len(page.Results) != 1 || page.TotalCount == nil || *page.TotalCount != 1 || string(page.Unknown["future_page"]) != `"kept"` {
		t.Fatalf("list page/response/error = %#v/%#v/%v", page, response, err)
	}
	got, response, err := client.Get(context.Background(), "team", "label")
	if err != nil || response.StatusCode != http.StatusOK || got.ID != "label-1" || got.SortOrder == nil || *got.SortOrder != 3 {
		t.Fatalf("get label/response/error = %#v/%#v/%v", got, response, err)
	}
}

func queryOrEmpty(query url.Values) url.Values {
	if query == nil {
		return url.Values{}
	}
	return query
}

func TestClientUpdateEmptyRequestSendsEmptyObjectAndDeleteReturns204(t *testing.T) {
	server := testsupport.NewServer(t, func(writer http.ResponseWriter, request *http.Request) {
		switch request.Method {
		case http.MethodPatch:
			body := testsupport.AssertRequest(t, request, testsupport.RequestExpectation{
				Method: http.MethodPatch, Path: "/api/v1/workspaces/team/project-labels/label/",
				ExpectedHeaders: http.Header{"Accept": {"application/json"}, "Content-Type": {"application/json"}, "X-API-Key": {"key"}},
				AbsentHeaders:   []string{"Authorization"},
			})
			if string(body) != `{}` {
				t.Errorf("empty update body = %s, want {}", body)
			}
			testsupport.JSON(writer, http.StatusOK, map[string]string{"id": "label"})
		case http.MethodDelete:
			testsupport.AssertRequest(t, request, testsupport.RequestExpectation{
				Method: http.MethodDelete, Path: "/api/v1/workspaces/team/project-labels/label/",
				ExpectedHeaders: http.Header{"Accept": {"application/json"}, "X-API-Key": {"key"}},
				AbsentHeaders:   []string{"Authorization"},
			})
			testsupport.NoContent(writer)
		}
	})
	shared, err := plane.NewClient(plane.Options{BaseURL: server.URL, AuthMode: plane.AuthModeAPIKey, APIKey: "key"})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := NewClient(shared).Update(context.Background(), "team", "label", UpdateProjectLabelRequest{}); err != nil {
		t.Fatal(err)
	}
	response, err := NewClient(shared).Delete(context.Background(), "team", "label")
	if err != nil || response.StatusCode != http.StatusNoContent {
		t.Fatalf("delete response = %#v, err=%v", response, err)
	}
}

func TestClientDelegatesSharedErrorsAndMalformedSuccess(t *testing.T) {
	tests := []struct {
		name       string
		handler    http.HandlerFunc
		wantStatus int
	}{
		{name: "typed api error", wantStatus: http.StatusUnprocessableEntity, handler: func(writer http.ResponseWriter, _ *http.Request) {
			testsupport.JSON(writer, http.StatusUnprocessableEntity, map[string]string{"error": "invalid", "detail": "bad request"})
		}},
		{name: "bounded non-json error", wantStatus: http.StatusBadGateway, handler: testsupport.NonJSONHandler(http.StatusBadGateway, strings.Repeat("x", 5000))},
		{name: "malformed success", wantStatus: http.StatusOK, handler: testsupport.MalformedJSONHandler(http.StatusOK)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := testsupport.NewServer(t, tt.handler)
			shared, err := plane.NewClient(plane.Options{BaseURL: server.URL, AuthMode: plane.AuthModeAPIKey, APIKey: "key"})
			if err != nil {
				t.Fatal(err)
			}
			_, response, err := NewClient(shared).Get(context.Background(), "team", "label")
			if err == nil || response.StatusCode != tt.wantStatus {
				t.Fatalf("error/status = %v/%d", err, response.StatusCode)
			}
			var apiError *plane.APIError
			if tt.name != "malformed success" && !errors.As(err, &apiError) {
				t.Fatalf("error type = %T, want *plane.APIError", err)
			}
			if tt.name == "bounded non-json error" && len(apiError.Message) > 4096 {
				t.Fatalf("bounded error length = %d", len(apiError.Message))
			}
		})
	}
}

func TestClientValidatesCreateBeforeSharedRequest(t *testing.T) {
	var calls int
	client := NewClient(countingClient{calls: &calls})
	if _, _, err := client.Create(context.Background(), "team", CreateProjectLabelRequest{}); err == nil {
		t.Fatal("empty create name was accepted")
	}
	if calls != 0 {
		t.Fatalf("shared client calls = %d, want 0", calls)
	}
}

type countingClient struct {
	calls *int
}

func (c countingClient) Do(_ context.Context, _ string, _ string, _ url.Values, _ any, _ http.Header, destination any) (plane.Response, error) {
	*c.calls++
	if destination != nil {
		_ = json.Unmarshal([]byte(`{"id":"label"}`), destination)
	}
	return plane.Response{StatusCode: http.StatusCreated}, nil
}
