package projects

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"

	"planeshift/internal/testsupport"
	"planeshift/plane"
)

func TestClientOperationMatrix(t *testing.T) {
	falseValue := false
	zeroValue := 0
	icon := json.RawMessage(`{"emoji":"rocket","color":"blue"}`)
	createRequest := CreateProjectRequest{
		Name: "Project", Identifier: "PRJ", ModuleView: &falseValue, ArchiveIn: &zeroValue,
		IconProp: &icon,
	}
	templateRequest := CreateProjectFromTemplateRequest{
		TemplateID: "template/id", Name: stringPointer("From template"), Network: intPointer(0),
	}
	updateRequest := UpdateProjectRequest{
		Description: stringPointer(""), CycleView: &falseValue, CloseIn: &zeroValue,
		DefaultState: stringPointer("state/id"), Estimate: stringPointer("estimate/id"),
	}

	tests := []struct {
		name       string
		method     string
		wantPath   string
		wantQuery  url.Values
		body       any
		call       func(context.Context, *Client) (plane.Response, error)
		wantStatus int
		wantPage   bool
	}{
		{
			name: "create", method: http.MethodPost,
			wantPath: "/api/v1/workspaces/team%20space/projects/", body: createRequest,
			call: func(ctx context.Context, client *Client) (plane.Response, error) {
				_, response, err := client.Create(ctx, "team space", createRequest)
				return response, err
			}, wantStatus: http.StatusCreated,
		},
		{
			name: "create from template", method: http.MethodPost,
			wantPath: "/api/v1/workspaces/team%20space/projects/templates/use/", body: templateRequest,
			call: func(ctx context.Context, client *Client) (plane.Response, error) {
				_, response, err := client.CreateFromTemplate(ctx, "team space", templateRequest)
				return response, err
			}, wantStatus: http.StatusCreated,
		},
		{
			name: "list", method: http.MethodGet,
			wantPath: "/api/v1/workspaces/team%20space/projects/",
			wantQuery: url.Values{
				"cursor": {"next:1"}, "expand": {"members"}, "fields": {"id,name"},
				"order_by": {"-created_at"}, "per_page": {"20"},
			}, wantPage: true,
			call: func(ctx context.Context, client *Client) (plane.Response, error) {
				page, response, err := client.List(ctx, "team space", ListOptions{
					Cursor: "next:1", Expand: "members", Fields: "id,name", OrderBy: "-created_at", PerPage: 20,
				})
				if err == nil && (page.GroupedBy == nil || *page.GroupedBy != "state" || len(page.Results) != 1 || page.Unknown["future_page"] == nil) {
					t.Fatalf("decoded project page = %#v", page)
				}
				if err == nil && (response.Metadata.RateLimit.Remaining != 42 || !response.Metadata.RateLimit.RemainingParsed) {
					t.Fatalf("response metadata = %#v", response.Metadata)
				}
				return response, err
			}, wantStatus: http.StatusOK,
		},
		{
			name: "get", method: http.MethodGet,
			wantPath: "/api/v1/workspaces/team%20space/projects/project%2Fid/",
			call: func(ctx context.Context, client *Client) (plane.Response, error) {
				_, response, err := client.Get(ctx, "team space", "project/id")
				return response, err
			}, wantStatus: http.StatusOK,
		},
		{
			name: "update", method: http.MethodPatch,
			wantPath: "/api/v1/workspaces/team%20space/projects/project%2Fid/", body: updateRequest,
			call: func(ctx context.Context, client *Client) (plane.Response, error) {
				_, response, err := client.Update(ctx, "team space", "project/id", updateRequest)
				return response, err
			}, wantStatus: http.StatusOK,
		},
		{
			name: "archive", method: http.MethodPost,
			wantPath: "/api/v1/workspaces/team%20space/projects/project%2Fid/archive/",
			call: func(ctx context.Context, client *Client) (plane.Response, error) {
				return client.Archive(ctx, "team space", "project/id")
			}, wantStatus: http.StatusNoContent,
		},
		{
			name: "unarchive", method: http.MethodDelete,
			wantPath: "/api/v1/workspaces/team%20space/projects/project%2Fid/archive/",
			call: func(ctx context.Context, client *Client) (plane.Response, error) {
				return client.Unarchive(ctx, "team space", "project/id")
			}, wantStatus: http.StatusNoContent,
		},
		{
			name: "delete", method: http.MethodDelete,
			wantPath: "/api/v1/workspaces/team%20space/projects/project%2Fid/",
			call: func(ctx context.Context, client *Client) (plane.Response, error) {
				return client.Delete(ctx, "team space", "project/id")
			}, wantStatus: http.StatusNoContent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := testsupport.NewServer(t, func(writer http.ResponseWriter, request *http.Request) {
				if request.Method != tt.method {
					t.Errorf("method = %q, want %q", request.Method, tt.method)
				}
				if request.URL.EscapedPath() != tt.wantPath {
					t.Errorf("escaped path = %q, want %q", request.URL.EscapedPath(), tt.wantPath)
				}
				if tt.wantQuery != nil && !sameValues(request.URL.Query(), tt.wantQuery) {
					t.Errorf("query = %v, want %v", request.URL.Query(), tt.wantQuery)
				}
				if request.Header.Get("X-API-Key") != "api-key" {
					t.Errorf("X-API-Key = %q, want api-key", request.Header.Get("X-API-Key"))
				}
				if request.Header.Get("Authorization") != "" {
					t.Errorf("Authorization = %q, want absent", request.Header.Get("Authorization"))
				}
				if tt.body == nil {
					body, err := io.ReadAll(request.Body)
					if err != nil {
						t.Errorf("read lifecycle body: %v", err)
					}
					if len(body) != 0 || request.Header.Get("Content-Type") != "" {
						t.Errorf("lifecycle body/content-type = %q/%q, want empty/absent", body, request.Header.Get("Content-Type"))
					}
				} else {
					body := testsupport.AssertRequest(t, request, testsupport.RequestExpectation{
						ExpectedHeaders: http.Header{"Content-Type": {"application/json"}},
					})
					testsupport.JSONBodyEqual(t, body, tt.body)
				}
				if tt.wantStatus == http.StatusNoContent {
					testsupport.NoContent(writer)
					return
				}
				if tt.wantPage {
					writer.Header().Set("X-RateLimit-Remaining", "42")
					testsupport.JSON(writer, tt.wantStatus, map[string]any{
						"grouped_by": "state", "sub_grouped_by": "priority", "total_count": 1,
						"next_cursor": "next:2", "prev_cursor": nil, "next_page_results": true,
						"prev_page_results": false, "count": 1, "total_pages": 1, "total_results": 1,
						"extra_stats": map[string]any{"total": 1}, "results": []any{map[string]any{"id": "project/id", "name": "Project", "future": true}},
						"future_page": nil,
					})
					return
				}
				testsupport.JSON(writer, tt.wantStatus, map[string]any{
					"id": "project/id", "name": "Project", "identifier": "PRJ", "description": nil,
					"is_deployed": false, "icon_prop": map[string]any{"color": "blue"}, "future": true,
				})
			})
			client, err := plane.NewClient(plane.Options{
				BaseURL: server.URL, AuthMode: plane.AuthModeAPIKey, APIKey: "api-key",
			})
			if err != nil {
				t.Fatal(err)
			}
			response, err := tt.call(context.Background(), NewClient(client))
			if err != nil {
				t.Fatal(err)
			}
			if response.StatusCode != tt.wantStatus {
				t.Fatalf("status = %d, want %d", response.StatusCode, tt.wantStatus)
			}
		})
	}
}

func TestClientBearerModeDoesNotAddAPIKey(t *testing.T) {
	server := testsupport.NewServer(t, func(writer http.ResponseWriter, request *http.Request) {
		testsupport.AssertRequest(t, request, testsupport.RequestExpectation{
			Method: http.MethodGet, Path: "/api/v1/workspaces/team/projects/",
			ExpectedHeaders: http.Header{
				"Accept":        {"application/json"},
				"Authorization": {"Bearer bearer-token"},
			}, AbsentHeaders: []string{"X-API-Key", "Content-Type"},
		})
		testsupport.JSON(writer, http.StatusOK, map[string]any{"results": []any{}})
	})
	client, err := plane.NewClient(plane.Options{BaseURL: server.URL, AuthMode: plane.AuthModeBearer, BearerToken: "bearer-token"})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := NewClient(client).List(context.Background(), "team", ListOptions{}); err != nil {
		t.Fatal(err)
	}
}

func TestClientListValidationAndRequiredBodies(t *testing.T) {
	var requests atomic.Int32
	fake := countingClient{requests: &requests}
	client := NewClient(fake)
	if _, _, err := client.Create(context.Background(), "team", CreateProjectRequest{Identifier: "PRJ"}); err == nil {
		t.Fatal("Create accepted missing name")
	}
	if _, _, err := client.CreateFromTemplate(context.Background(), "team", CreateProjectFromTemplateRequest{}); err == nil {
		t.Fatal("CreateFromTemplate accepted missing template_id")
	}
	if _, _, err := client.List(context.Background(), "team", ListOptions{PerPage: 101}); err == nil {
		t.Fatal("List accepted per_page=101")
	}
	if requests.Load() != 0 {
		t.Fatalf("validation sent %d request(s)", requests.Load())
	}
}

func TestClientDelegatesTypedAndNonJSONErrors(t *testing.T) {
	secret := "not-in-error"
	server := testsupport.NewServer(t, func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path == "/api/v1/workspaces/team/projects/project/" {
			writer.Header().Set("Content-Type", "application/json")
			writer.WriteHeader(http.StatusUnprocessableEntity)
			_, _ = io.WriteString(writer, fmt.Sprintf(`{"error":"invalid","detail":"bad %s"}`, secret))
			return
		}
		testsupport.NonJSON(writer, http.StatusBadGateway, strings.Repeat("x", 5000))
	})
	client, err := plane.NewClient(plane.Options{BaseURL: server.URL, AuthMode: plane.AuthModeBearer, BearerToken: secret})
	if err != nil {
		t.Fatal(err)
	}
	_, response, err := NewClient(client).Get(context.Background(), "team", "project")
	if err == nil {
		t.Fatal("Get returned nil error for JSON failure")
	}
	var apiError *plane.APIError
	if !errors.As(err, &apiError) || apiError.StatusCode != http.StatusUnprocessableEntity || response.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("JSON error = %T %#v, response = %#v", err, apiError, response)
	}
	if strings.Contains(err.Error(), secret) {
		t.Fatalf("JSON error disclosed credential: %q", err)
	}

	_, err = NewClient(client).Archive(context.Background(), "team", "other")
	var nonJSONError *plane.APIError
	if !errors.As(err, &nonJSONError) || nonJSONError.StatusCode != http.StatusBadGateway || len(nonJSONError.Message) > 4096 {
		t.Fatalf("non-JSON error = %#v", err)
	}
}

type countingClient struct {
	requests *atomic.Int32
}

func (c countingClient) Do(context.Context, string, string, url.Values, any, http.Header, any) (plane.Response, error) {
	c.requests.Add(1)
	return plane.Response{}, nil
}

func sameValues(left, right url.Values) bool {
	if len(left) != len(right) {
		return false
	}
	for key, want := range right {
		got, ok := left[key]
		if !ok || len(got) != len(want) {
			return false
		}
		for index := range got {
			if got[index] != want[index] {
				return false
			}
		}
	}
	return true
}

func stringPointer(value string) *string { return &value }

func intPointer(value int) *int { return &value }
