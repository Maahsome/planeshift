package projectfeatures

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
	falseValue := false
	trueValue := true
	request := UpdateProjectFeaturesRequest{
		Epics:   &falseValue,
		Modules: &trueValue,
		Views:   &falseValue,
	}

	tests := []struct {
		name       string
		method     string
		wantAuth   http.Header
		absentAuth string
		body       any
		call       func(context.Context, *Client) (ProjectFeatures, plane.Response, error)
	}{
		{
			name: "get with api key", method: http.MethodGet,
			wantAuth: http.Header{"X-API-Key": {"api-key"}}, absentAuth: "Authorization",
			call: func(ctx context.Context, client *Client) (ProjectFeatures, plane.Response, error) {
				return client.Get(ctx, "team space", "project/id")
			},
		},
		{
			name: "update with bearer", method: http.MethodPatch,
			wantAuth: http.Header{"Authorization": {"Bearer bearer-token"}}, absentAuth: "X-API-Key",
			body: request,
			call: func(ctx context.Context, client *Client) (ProjectFeatures, plane.Response, error) {
				return client.Update(ctx, "team space", "project/id", request)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := testsupport.NewServer(t, func(writer http.ResponseWriter, req *http.Request) {
				body := testsupport.AssertRequest(t, req, testsupport.RequestExpectation{
					Method:          tt.method,
					Path:            "/api/v1/workspaces/team space/projects/project/id/features/",
					Query:           urlValuesEmpty(),
					ExpectedHeaders: mergeHeaders(tt.wantAuth, http.Header{"Accept": {"application/json"}}),
					AbsentHeaders:   []string{tt.absentAuth},
				})
				if req.URL.EscapedPath() != "/api/v1/workspaces/team%20space/projects/project%2Fid/features/" {
					t.Errorf("escaped path = %q", req.URL.EscapedPath())
				}
				if req.Method != tt.method {
					t.Errorf("method = %q, want %q", req.Method, tt.method)
				}
				if tt.body == nil {
					if len(body) != 0 || req.Header.Get("Content-Type") != "" {
						t.Errorf("GET body/content-type = %q/%q, want empty/absent", body, req.Header.Get("Content-Type"))
					}
				} else {
					if req.Header.Get("Content-Type") != "application/json" {
						t.Errorf("Content-Type = %q, want application/json", req.Header.Get("Content-Type"))
					}
					testsupport.JSONBodyEqual(t, body, tt.body)
					for _, omitted := range []string{"cycles", "pages", "intakes", "work_item_types"} {
						if strings.Contains(string(body), `"`+omitted+`"`) {
							t.Errorf("body serialized omitted field %q: %s", omitted, body)
						}
					}
				}
				writer.Header().Set("X-RateLimit-Remaining", "42")
				writer.Header().Set("X-RateLimit-Reset", "1700327957")
				testsupport.JSON(writer, http.StatusOK, map[string]any{
					"epics": true, "modules": false, "cycles": true, "views": false,
					"pages": true, "intakes": false, "work_item_types": true,
					"future": nil,
				})
			})

			options := plane.Options{BaseURL: server.URL}
			if tt.name == "get with api key" {
				options.AuthMode = plane.AuthModeAPIKey
				options.APIKey = "api-key"
			} else {
				options.AuthMode = plane.AuthModeBearer
				options.BearerToken = "bearer-token"
			}
			shared, err := plane.NewClient(options)
			if err != nil {
				t.Fatal(err)
			}
			features, response, err := tt.call(context.Background(), NewClient(shared))
			if err != nil {
				t.Fatal(err)
			}
			if response.StatusCode != http.StatusOK || response.RateLimit.Remaining != 42 || response.RateLimit.Reset != 1700327957 {
				t.Fatalf("response metadata = %#v", response)
			}
			if !features.Epics || features.Modules || !features.Cycles || features.Views || !features.Pages || features.Intakes || !features.WorkItemTypes {
				t.Fatalf("decoded features = %#v", features)
			}
			if string(features.Unknown["future"]) != "null" {
				t.Fatalf("unknown response field = %#v", features.Unknown)
			}
		})
	}
}

func TestClientUpdateEmptyRequestSendsEmptyTypedObject(t *testing.T) {
	server := testsupport.NewServer(t, func(writer http.ResponseWriter, request *http.Request) {
		body := testsupport.AssertRequest(t, request, testsupport.RequestExpectation{
			Method: http.MethodPatch,
			Path:   "/api/v1/workspaces/team/projects/project/features/",
			ExpectedHeaders: http.Header{
				"Accept": {"application/json"}, "Content-Type": {"application/json"},
				"X-API-Key": {"api-key"},
			},
			AbsentHeaders: []string{"Authorization"},
		})
		if string(body) != `{}` {
			t.Errorf("empty update body = %s, want {}", body)
		}
		testsupport.JSON(writer, http.StatusOK, map[string]any{})
	})
	shared, err := plane.NewClient(plane.Options{BaseURL: server.URL, AuthMode: plane.AuthModeAPIKey, APIKey: "api-key"})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := NewClient(shared).Update(context.Background(), "team", "project", UpdateProjectFeaturesRequest{}); err != nil {
		t.Fatal(err)
	}
}

func TestClientUpdateRequestPresenceForEveryBooleanField(t *testing.T) {
	falseValue := false
	tests := []struct {
		name string
		make func(*bool) UpdateProjectFeaturesRequest
		key  string
	}{
		{name: "epics", make: func(value *bool) UpdateProjectFeaturesRequest { return UpdateProjectFeaturesRequest{Epics: value} }, key: "epics"},
		{name: "modules", make: func(value *bool) UpdateProjectFeaturesRequest { return UpdateProjectFeaturesRequest{Modules: value} }, key: "modules"},
		{name: "cycles", make: func(value *bool) UpdateProjectFeaturesRequest { return UpdateProjectFeaturesRequest{Cycles: value} }, key: "cycles"},
		{name: "views", make: func(value *bool) UpdateProjectFeaturesRequest { return UpdateProjectFeaturesRequest{Views: value} }, key: "views"},
		{name: "pages", make: func(value *bool) UpdateProjectFeaturesRequest { return UpdateProjectFeaturesRequest{Pages: value} }, key: "pages"},
		{name: "intakes", make: func(value *bool) UpdateProjectFeaturesRequest { return UpdateProjectFeaturesRequest{Intakes: value} }, key: "intakes"},
		{name: "work item types", make: func(value *bool) UpdateProjectFeaturesRequest {
			return UpdateProjectFeaturesRequest{WorkItemTypes: value}
		}, key: "work_item_types"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.make(&falseValue))
			if err != nil {
				t.Fatal(err)
			}
			if string(data) != `{"`+tt.key+`":false}` {
				t.Fatalf("request = %s, want only explicit false %q", data, tt.key)
			}
		})
	}
}

func TestClientDelegatesSharedErrorsAndMalformedSuccess(t *testing.T) {
	tests := []struct {
		name       string
		status     int
		handler    http.HandlerFunc
		wantAPIErr bool
	}{
		{
			name: "typed api error", status: http.StatusUnprocessableEntity, wantAPIErr: true,
			handler: func(writer http.ResponseWriter, request *http.Request) {
				testsupport.JSON(writer, http.StatusUnprocessableEntity, map[string]string{"error": "invalid", "detail": "bad request"})
			},
		},
		{
			name: "bounded non-json error", status: http.StatusBadGateway, wantAPIErr: true,
			handler: testsupport.NonJSONHandler(http.StatusBadGateway, strings.Repeat("x", 5000)),
		},
		{
			name: "malformed success", status: http.StatusOK,
			handler: testsupport.MalformedJSONHandler(http.StatusOK),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := testsupport.NewServer(t, tt.handler)
			shared, err := plane.NewClient(plane.Options{BaseURL: server.URL, AuthMode: plane.AuthModeAPIKey, APIKey: "api-key"})
			if err != nil {
				t.Fatal(err)
			}
			_, response, err := NewClient(shared).Get(context.Background(), "team", "project")
			if err == nil || response.StatusCode != tt.status {
				t.Fatalf("error/status = %v/%d, want error and %d", err, response.StatusCode, tt.status)
			}
			var apiError *plane.APIError
			if errors.As(err, &apiError) != tt.wantAPIErr {
				t.Fatalf("API error type = %T, wantAPIErr=%v", err, tt.wantAPIErr)
			}
			if tt.name == "bounded non-json error" && len(apiError.Message) > 4096 {
				t.Fatalf("error message exceeded bound: %d", len(apiError.Message))
			}
		})
	}
}

func TestClientAcceptsShared204WithoutDocumentingItAsFeatureSuccess(t *testing.T) {
	server := testsupport.NewServer(t, testsupport.NoContentHandler())
	shared, err := plane.NewClient(plane.Options{BaseURL: server.URL, AuthMode: plane.AuthModeAPIKey, APIKey: "api-key"})
	if err != nil {
		t.Fatal(err)
	}
	features, response, err := NewClient(shared).Get(context.Background(), "team", "project")
	if err != nil || response.StatusCode != http.StatusNoContent {
		t.Fatalf("204 result = %#v/%v/%v", features, response, err)
	}
	if features.Epics || features.Modules || features.Cycles || features.Views || features.Pages || features.Intakes || features.WorkItemTypes || features.Unknown != nil {
		t.Fatalf("204 decoded a feature object: %#v", features)
	}
}

func urlValuesEmpty() url.Values {
	return url.Values{}
}

func mergeHeaders(left, right http.Header) http.Header {
	merged := left.Clone()
	for key, values := range right {
		merged[key] = append([]string(nil), values...)
	}
	return merged
}
