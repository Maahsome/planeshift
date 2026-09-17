package states

import (
	"context"
	"errors"
	"fmt"
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
	createRequest := CreateStateRequest{
		Name: "Example", Color: "#ff0000", Description: stringPointer(""),
		Sequence: jsonRaw("2"), Group: stringPointer("triage"), IsTriage: &falseValue,
		Default: &falseValue, ExternalSource: stringPointer("github"), ExternalID: stringPointer("external-1"),
	}
	updateRequest := UpdateStateRequest{
		Description: stringPointer(""), Sequence: jsonRaw("null"), IsTriage: &falseValue,
		ExternalID: stringPointer(""),
	}

	tests := []struct {
		name      string
		method    string
		path      string
		query     url.Values
		body      any
		status    int
		wantPage  bool
		wantEmpty bool
		call      func(context.Context, *Client) (plane.Response, error)
	}{
		{
			name: "create", method: http.MethodPost,
			path: "/api/v1/workspaces/team%20space/projects/project%2Fid/states/", body: createRequest,
			status: http.StatusOK,
			call: func(ctx context.Context, client *Client) (plane.Response, error) {
				_, response, err := client.Create(ctx, "team space", "project/id", createRequest)
				return response, err
			},
		},
		{
			name: "list", method: http.MethodGet,
			path:   "/api/v1/workspaces/team%20space/projects/project%2Fid/states/",
			query:  url.Values{"cursor": {"20:1:0"}, "expand": {"project,workspace"}, "fields": {"id,name"}, "per_page": {"20"}},
			status: http.StatusOK, wantPage: true,
			call: func(ctx context.Context, client *Client) (plane.Response, error) {
				page, response, err := client.List(ctx, "team space", "project/id", ListOptions{
					Cursor: "20:1:0", PerPage: 20, Fields: "id,name", Expand: "project,workspace",
				})
				if err == nil {
					if page.GroupedBy == nil || *page.GroupedBy != "state" || len(page.Results) != 1 || page.Unknown["future_page"] == nil {
						t.Fatalf("decoded state page = %#v", page)
					}
					if page.Results[0].Sequence == nil || string(page.Results[0].Sequence) != `"2"` {
						t.Fatalf("decoded sequence = %s", page.Results[0].Sequence)
					}
					if response.Metadata.RateLimit.Remaining != 42 || !response.Metadata.RateLimit.RemainingParsed {
						t.Fatalf("rate limit metadata = %#v", response.Metadata.RateLimit)
					}
				}
				return response, err
			},
		},
		{
			name: "get", method: http.MethodGet,
			path: "/api/v1/workspaces/team%20space/projects/project%2Fid/states/state%2Fid/", status: http.StatusOK,
			call: func(ctx context.Context, client *Client) (plane.Response, error) {
				_, response, err := client.Get(ctx, "team space", "project/id", "state/id")
				return response, err
			},
		},
		{
			name: "update", method: http.MethodPatch,
			path: "/api/v1/workspaces/team%20space/projects/project%2Fid/states/state%2Fid/", body: updateRequest, status: http.StatusOK,
			call: func(ctx context.Context, client *Client) (plane.Response, error) {
				_, response, err := client.Update(ctx, "team space", "project/id", "state/id", updateRequest)
				return response, err
			},
		},
		{
			name: "delete", method: http.MethodDelete,
			path: "/api/v1/workspaces/team%20space/projects/project%2Fid/states/state%2Fid/", status: http.StatusNoContent, wantEmpty: true,
			call: func(ctx context.Context, client *Client) (plane.Response, error) {
				return client.Delete(ctx, "team space", "project/id", "state/id")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := testsupport.NewServer(t, func(writer http.ResponseWriter, request *http.Request) {
				body := testsupport.AssertRequest(t, request, testsupport.RequestExpectation{
					Method:          tt.method,
					Query:           tt.query,
					ExpectedHeaders: http.Header{"X-API-Key": {"api-key"}, "Accept": {"application/json"}},
					AbsentHeaders:   []string{"Authorization"},
				})
				if request.URL.EscapedPath() != tt.path {
					t.Errorf("escaped path = %q, want %q", request.URL.EscapedPath(), tt.path)
				}
				if tt.query == nil && request.URL.RawQuery != "" {
					t.Errorf("detail query = %q, want empty", request.URL.RawQuery)
				}
				if tt.body == nil {
					if len(body) != 0 || request.Header.Get("Content-Type") != "" {
						t.Errorf("body/content-type = %q/%q, want empty/absent", body, request.Header.Get("Content-Type"))
					}
				} else {
					if request.Header.Get("Content-Type") != "application/json" {
						t.Errorf("Content-Type = %q, want application/json", request.Header.Get("Content-Type"))
					}
					testsupport.JSONBodyEqual(t, body, tt.body)
				}
				if tt.wantEmpty {
					testsupport.NoContent(writer)
					return
				}
				if tt.wantPage {
					writer.Header().Set("X-RateLimit-Remaining", "42")
					testsupport.JSON(writer, tt.status, map[string]any{
						"grouped_by": "state", "sub_grouped_by": "priority", "total_count": 1,
						"next_cursor": "20:2:0", "prev_cursor": nil, "next_page_results": false,
						"prev_page_results": false, "count": 1, "total_pages": 1, "total_results": 1,
						"extra_stats": nil, "results": []any{map[string]any{"id": "state-1", "name": "Started", "sequence": "2"}},
						"future_page": map[string]any{"kept": true},
					})
					return
				}
				testsupport.JSON(writer, tt.status, map[string]any{
					"id": "state-1", "name": "Started", "description": nil, "color": "#ff0000",
					"sequence": 2, "group": "started", "default": false, "future": map[string]any{"keep": true},
				})
			})
			client, err := plane.NewClient(plane.Options{BaseURL: server.URL, AuthMode: plane.AuthModeAPIKey, APIKey: "api-key"})
			if err != nil {
				t.Fatal(err)
			}
			response, err := tt.call(context.Background(), NewClient(client))
			if err != nil {
				t.Fatal(err)
			}
			if response.StatusCode != tt.status {
				t.Fatalf("status = %d, want %d", response.StatusCode, tt.status)
			}
		})
	}
}

func TestClientBearerModeDoesNotAddAPIKey(t *testing.T) {
	server := testsupport.NewServer(t, func(writer http.ResponseWriter, request *http.Request) {
		testsupport.AssertRequest(t, request, testsupport.RequestExpectation{
			Method: http.MethodGet, Path: "/api/v1/workspaces/team/projects/project/states/state/",
			ExpectedHeaders: http.Header{"Accept": {"application/json"}, "Authorization": {"Bearer bearer-token"}},
			AbsentHeaders:   []string{"X-API-Key", "Content-Type"},
		})
		testsupport.JSON(writer, http.StatusOK, map[string]any{"id": "state-1"})
	})
	client, err := plane.NewClient(plane.Options{BaseURL: server.URL, AuthMode: plane.AuthModeBearer, BearerToken: "bearer-token"})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := NewClient(client).Get(context.Background(), "team", "project", "state"); err != nil {
		t.Fatal(err)
	}
}

func TestClientValidationOccursBeforeRequest(t *testing.T) {
	var requests atomic.Int32
	client := NewClient(countingClient{requests: &requests})
	if _, _, err := client.Create(context.Background(), "team", "project", CreateStateRequest{Color: "#fff"}); err == nil {
		t.Fatal("Create accepted missing name")
	}
	if _, _, err := client.Create(context.Background(), "team", "project", CreateStateRequest{Name: "State"}); err == nil {
		t.Fatal("Create accepted missing color")
	}
	if _, _, err := client.List(context.Background(), "team", "project", ListOptions{PerPage: 101}); err == nil {
		t.Fatal("List accepted per_page=101")
	}
	if _, _, err := client.Update(context.Background(), "team", "project", "state", UpdateStateRequest{Sequence: jsonRaw("{}")}); err == nil {
		t.Fatal("Update accepted object sequence")
	}
	if requests.Load() != 0 {
		t.Fatalf("validation sent %d request(s)", requests.Load())
	}
}

func TestClientSharedErrorsMalformedAndEmptyResponses(t *testing.T) {
	secret := "state-secret"
	server := testsupport.NewServer(t, func(writer http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/api/v1/workspaces/team/projects/project/states/state/":
			writer.Header().Set("Content-Type", "application/json")
			writer.WriteHeader(http.StatusUnprocessableEntity)
			_, _ = fmt.Fprintf(writer, `{"error":"invalid","detail":"bad %s"}`, secret)
		case "/api/v1/workspaces/team/projects/project/states/":
			testsupport.MalformedJSON(writer, http.StatusOK)
		default:
			testsupport.Empty(writer, http.StatusOK)
		}
	})
	client, err := plane.NewClient(plane.Options{BaseURL: server.URL, AuthMode: plane.AuthModeBearer, BearerToken: secret})
	if err != nil {
		t.Fatal(err)
	}
	_, response, err := NewClient(client).Get(context.Background(), "team", "project", "state")
	if err == nil || response.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("shared API error = %v, response = %#v", err, response)
	}
	var apiError *plane.APIError
	if !errors.As(err, &apiError) || strings.Contains(err.Error(), secret) {
		t.Fatalf("error type/secret handling = %T/%v", err, err)
	}
	_, _, err = NewClient(client).List(context.Background(), "team", "project", ListOptions{})
	if err == nil || !strings.Contains(err.Error(), "malformed JSON") {
		t.Fatalf("malformed success error = %v", err)
	}
	response, err = NewClient(client).Delete(context.Background(), "team", "project", "empty")
	if err != nil || response.StatusCode != http.StatusOK {
		t.Fatalf("empty success = %#v/%v", response, err)
	}
}

type countingClient struct {
	requests *atomic.Int32
}

func (c countingClient) Do(context.Context, string, string, url.Values, any, http.Header, any) (plane.Response, error) {
	c.requests.Add(1)
	return plane.Response{}, nil
}

func stringPointer(value string) *string {
	return &value
}
