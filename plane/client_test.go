package plane

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"planeshift/internal/testsupport"
)

func TestClientAuthenticationModesAndJSON(t *testing.T) {
	tests := []struct {
		name       string
		mode       string
		apiKey     string
		bearer     string
		selected   string
		unselected string
	}{
		{name: "api key", mode: AuthModeAPIKey, apiKey: "api-key-secret", selected: "X-API-Key", unselected: "Authorization"},
		{name: "bearer", mode: AuthModeBearer, bearer: "bearer-secret", selected: "Authorization", unselected: "X-API-Key"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := testsupport.NewServer(t, func(writer http.ResponseWriter, request *http.Request) {
				testsupport.AssertRequest(t, request, testsupport.RequestExpectation{
					Method: http.MethodGet,
					Path:   "/api/v1/workspaces/team/projects/",
					Query:  url.Values{"filter": {"hello world"}},
					ExpectedHeaders: http.Header{
						"Accept":    {"application/json"},
						tt.selected: {map[string]string{"X-API-Key": tt.apiKey, "Authorization": "Bearer " + tt.bearer}[tt.selected]},
					},
					AbsentHeaders: []string{tt.unselected, "Content-Type"},
				})
				testsupport.JSON(writer, http.StatusOK, map[string]any{"id": "project-1", "unknown": true})
			})
			client, err := NewClient(Options{BaseURL: server.URL, AuthMode: tt.mode, APIKey: tt.apiKey, BearerToken: tt.bearer})
			if err != nil {
				t.Fatal(err)
			}
			var destination map[string]any
			response, err := client.Do(context.Background(), http.MethodGet, "/workspaces/team/projects/", url.Values{"filter": {"hello world"}}, nil, nil, &destination)
			if err != nil {
				t.Fatal(err)
			}
			if response.StatusCode != http.StatusOK || destination["unknown"] != true {
				t.Fatalf("response = %#v, destination = %#v", response, destination)
			}
		})
	}
}

func TestClientJSONBodyAndSafeAdditionalHeaders(t *testing.T) {
	server := testsupport.NewServer(t, func(writer http.ResponseWriter, request *http.Request) {
		body := testsupport.AssertRequest(t, request, testsupport.RequestExpectation{
			Method: http.MethodPost,
			Path:   "/api/v1/widgets/",
			ExpectedHeaders: http.Header{
				"Accept":       {"application/json"},
				"Content-Type": {"application/json"},
				"X-Request-ID": {"request-1"},
				"X-API-Key":    {"request-key"},
			},
		})
		testsupport.JSONBodyEqual(t, body, map[string]any{"name": "widget", "nullable": nil})
		testsupport.JSON(writer, http.StatusCreated, map[string]any{"created": true})
	})
	client, err := NewClient(Options{BaseURL: server.URL, AuthMode: AuthModeAPIKey, APIKey: "request-key"})
	if err != nil {
		t.Fatal(err)
	}
	var destination map[string]any
	response, err := client.Do(context.Background(), http.MethodPost, "widgets/", nil,
		map[string]any{"name": "widget", "nullable": nil}, http.Header{"X-Request-ID": {"request-1"}}, &destination)
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != http.StatusCreated || destination["created"] != true {
		t.Fatalf("response = %#v, destination = %#v", response, destination)
	}
}

func TestURLNormalizationAndPaginationQuery(t *testing.T) {
	tests := []struct {
		name string
		base string
		want string
	}{
		{name: "default", want: "https://api.plane.so/api/v1/"},
		{name: "host", base: "https://plane.example/", want: "https://plane.example/api/v1/"},
		{name: "self hosted path", base: "https://plane.example/plane/", want: "https://plane.example/plane/api/v1/"},
		{name: "already prefixed", base: "https://plane.example/api/v1", want: "https://plane.example/api/v1/"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NormalizeBaseURL(tt.base)
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Fatalf("NormalizeBaseURL() = %q, want %q", got, tt.want)
			}
		})
	}
	for _, invalid := range []string{
		"not a URL", "ftp://plane.example", "https://user:password@plane.example", "https://plane.example/#fragment",
		"https://plane.example/?unexpected=query", "https://plane.example?", "https://",
	} {
		t.Run("invalid "+invalid, func(t *testing.T) {
			if _, err := NormalizeBaseURL(invalid); err == nil {
				t.Fatalf("NormalizeBaseURL(%q) accepted invalid URL", invalid)
			}
		})
	}

	query := url.Values{"unknown": {"a&b"}}
	withPagination, err := WithPagination(query, PaginationOptions{PerPage: 20, Cursor: "20:1:0", Fields: "id,name", Expand: "assignees"})
	if err != nil {
		t.Fatal(err)
	}
	if len(query) != 1 || query.Get("per_page") != "" {
		t.Fatalf("WithPagination mutated input: %v", query)
	}
	wantQuery := url.Values{
		"unknown": {"a&b"}, "per_page": {"20"}, "cursor": {"20:1:0"},
		"fields": {"id,name"}, "expand": {"assignees"},
	}
	if !reflectDeepEqualValues(withPagination, wantQuery) {
		t.Fatalf("pagination query = %v, want %v", withPagination, wantQuery)
	}
	requestURL, err := BuildURL("https://plane.example/api/v1/", "/widgets/a b/", withPagination)
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := url.Parse(requestURL)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.EscapedPath() != "/api/v1/widgets/a%20b/" || !reflectDeepEqualValues(parsed.Query(), wantQuery) {
		t.Fatalf("built URL = %q, path/query mismatch", requestURL)
	}
	encodedSegmentURL, err := BuildURL("https://plane.example", "/widgets/"+EscapePathSegment("a/b")+"/", nil)
	if err != nil {
		t.Fatal(err)
	}
	encodedSegment, err := url.Parse(encodedSegmentURL)
	if err != nil || encodedSegment.EscapedPath() != "/api/v1/widgets/a%2Fb/" {
		t.Fatalf("escaped path = %q, want encoded segment", encodedSegmentURL)
	}
	if !strings.Contains(parsed.RawQuery, "unknown=a%26b") || !strings.Contains(parsed.RawQuery, "cursor=20%3A1%3A0") {
		t.Fatalf("query was not encoded safely: %q", parsed.RawQuery)
	}
	if _, err := BuildURL("https://plane.example", "/widgets?bad=true", nil); err == nil {
		t.Fatal("BuildURL accepted route query state")
	}
	if _, err := BuildURL("https://plane.example", "/widgets/#bad", nil); err == nil {
		t.Fatal("BuildURL accepted route fragment")
	}

	for _, perPage := range []int{-1, 101} {
		if _, err := WithPagination(nil, PaginationOptions{PerPage: perPage}); err == nil {
			t.Fatalf("WithPagination accepted per_page=%d", perPage)
		}
	}
	if _, err := BuildURL("https://plane.example", "/widgets", url.Values{"per_page": {"0"}}); err == nil {
		t.Fatal("BuildURL accepted query per_page=0")
	}
}

func TestClientResponseMetadataAndDynamicPage(t *testing.T) {
	server := testsupport.NewServer(t, func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("X-RateLimit-Remaining", "45")
		writer.Header().Set("X-RateLimit-Reset", "1700327957")
		writer.Header().Set("X-Internal-Secret", "must-not-be-retained")
		testsupport.JSON(writer, http.StatusOK, map[string]any{
			"next_cursor":       nil,
			"prev_cursor":       "20:0:1",
			"next_page_results": true,
			"prev_page_results": false,
			"count":             1,
			"total_pages":       3,
			"total_results":     3,
			"extra_stats":       map[string]any{"sum": 1},
			"results":           []any{map[string]any{"id": "1", "nullable": nil, "new_field": "kept"}},
			"future_field":      nil,
		})
	})
	client, err := NewClient(Options{BaseURL: server.URL, AuthMode: AuthModeAPIKey, APIKey: "key"})
	if err != nil {
		t.Fatal(err)
	}
	var page RawCursorPage
	response, err := client.Do(context.Background(), http.MethodGet, "/widgets/", nil, nil, nil, &page)
	if err != nil {
		t.Fatal(err)
	}
	if response.Metadata.RateLimit.Remaining != 45 || !response.Metadata.RateLimit.RemainingParsed ||
		response.Metadata.RateLimit.Reset != 1700327957 || response.RateLimit.RemainingValue != "45" {
		t.Fatalf("rate metadata = %#v", response.Metadata.RateLimit)
	}
	if response.Metadata.Pagination == nil || response.Metadata.Pagination.NextCursor != nil {
		t.Fatalf("pagination metadata = %#v", response.Metadata.Pagination)
	}
	if len(page.Results) != 1 || !strings.Contains(string(page.Results[0]), `"id":"1"`) ||
		!strings.Contains(string(page.Results[0]), `"nullable":null`) ||
		!strings.Contains(string(page.Results[0]), `"new_field":"kept"`) || string(page.Unknown["future_field"]) != "null" {
		t.Fatalf("dynamic page = %#v", page)
	}
}

func TestRateLimitMetadataPreservesUnparseableValues(t *testing.T) {
	server := testsupport.NewServer(t, func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("X-RateLimit-Remaining", "later")
		writer.Header().Set("X-RateLimit-Reset", "unknown")
		testsupport.Empty(writer, http.StatusOK)
	})
	client, err := NewClient(Options{BaseURL: server.URL, AuthMode: AuthModeAPIKey, APIKey: "key"})
	if err != nil {
		t.Fatal(err)
	}
	response, err := client.Do(context.Background(), http.MethodGet, "/widgets/", nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	rateLimit := response.Metadata.RateLimit
	if rateLimit.RemainingValue != "later" || rateLimit.ResetValue != "unknown" || rateLimit.RemainingParsed || rateLimit.ResetParsed {
		t.Fatalf("rate metadata = %#v", rateLimit)
	}
}

func TestClientEmptyMalformedAndNoContentResponses(t *testing.T) {
	tests := []struct {
		name    string
		handler http.HandlerFunc
		wantErr string
	}{
		{name: "empty", handler: testsupport.EmptyHandler(http.StatusOK)},
		{name: "malformed", handler: testsupport.MalformedJSONHandler(http.StatusOK), wantErr: "malformed JSON"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := testsupport.NewServer(t, tt.handler)
			client, err := NewClient(Options{BaseURL: server.URL, AuthMode: AuthModeAPIKey, APIKey: "key"})
			if err != nil {
				t.Fatal(err)
			}
			var destination map[string]any
			_, err = client.Do(context.Background(), http.MethodGet, "/widgets/", nil, nil, nil, &destination)
			if tt.wantErr == "" && err != nil {
				t.Fatal(err)
			}
			if tt.wantErr != "" && (err == nil || !strings.Contains(err.Error(), tt.wantErr)) {
				t.Fatalf("error = %v, want %q", err, tt.wantErr)
			}
		})
	}

	closed := &trackingBody{reader: strings.NewReader(`{"should":"not decode"}`)}
	transport := roundTripFunc(func(request *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusNoContent, Header: http.Header{}, Body: closed, Request: request}, nil
	})
	client, err := NewClient(Options{BaseURL: "https://plane.example", AuthMode: AuthModeAPIKey, APIKey: "key", RoundTripper: transport})
	if err != nil {
		t.Fatal(err)
	}
	var destination map[string]any
	if _, err := client.Do(context.Background(), http.MethodDelete, "/widgets/1/", nil, nil, nil, &destination); err != nil {
		t.Fatal(err)
	}
	if closed.read || !closed.closed {
		t.Fatalf("204 body read=%v closed=%v", closed.read, closed.closed)
	}
}

func TestClientTypedErrorsAreBoundedAndRedacted(t *testing.T) {
	secret := "client-secret"
	server := testsupport.NewServer(t, func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusUnauthorized)
		_, _ = io.WriteString(writer, fmt.Sprintf(`{"error":"invalid","detail":"%s","message":"token %s at https://storage.example/presigned?sig=secret"}`, secret, secret))
	})
	client, err := NewClient(Options{BaseURL: server.URL, AuthMode: AuthModeBearer, BearerToken: secret})
	if err != nil {
		t.Fatal(err)
	}
	response, err := client.Do(context.Background(), http.MethodGet, "/widgets/", nil, nil, nil, nil)
	if err == nil {
		t.Fatal("Do returned nil error for 401")
	}
	var apiError *APIError
	if !errors.As(err, &apiError) {
		t.Fatalf("error type = %T, want *APIError", err)
	}
	if apiError.StatusCode != http.StatusUnauthorized || response.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status response=%#v error=%#v", response, apiError)
	}
	if strings.Contains(err.Error(), secret) || strings.Contains(err.Error(), "https://storage.example") {
		t.Fatalf("error disclosed sensitive data: %q", err)
	}
	if len(apiError.Error()) > maxErrorMessageBytes+64 {
		t.Fatalf("error was not bounded: %d bytes", len(apiError.Error()))
	}

	nonJSONServer := testsupport.NewServer(t, func(writer http.ResponseWriter, request *http.Request) {
		testsupport.NonJSON(writer, http.StatusBadGateway, strings.Repeat("x", maxErrorMessageBytes+100))
	})
	nonJSONClient, err := NewClient(Options{BaseURL: nonJSONServer.URL, AuthMode: AuthModeAPIKey, APIKey: "key"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = nonJSONClient.Do(context.Background(), http.MethodGet, "/widgets/", nil, nil, nil, nil)
	var nonJSONError *APIError
	if !errors.As(err, &nonJSONError) || len(nonJSONError.Message) > maxErrorMessageBytes {
		t.Fatalf("non-JSON error = %#v", err)
	}
}

func TestClientInvalidConfigurationAndNoRetry(t *testing.T) {
	var requests atomic.Int32
	transport := roundTripFunc(func(request *http.Request) (*http.Response, error) {
		requests.Add(1)
		return &http.Response{StatusCode: http.StatusTooManyRequests, Header: http.Header{}, Body: io.NopCloser(strings.NewReader("rate limited")), Request: request}, nil
	})
	invalid := []Options{
		{BaseURL: "https://plane.example", AuthMode: "", APIKey: "key", RoundTripper: transport},
		{BaseURL: "https://plane.example", AuthMode: AuthModeAPIKey, APIKey: "key", BearerToken: "token", RoundTripper: transport},
		{BaseURL: "https://plane.example", AuthMode: AuthModeAPIKey, Timeout: -time.Second, APIKey: "key", RoundTripper: transport},
	}
	for _, options := range invalid {
		if _, err := NewClient(options); err == nil {
			t.Fatalf("NewClient accepted invalid options: %#v", options)
		}
	}
	if requests.Load() != 0 {
		t.Fatalf("invalid configuration sent %d request(s)", requests.Load())
	}

	client, err := NewClient(Options{BaseURL: "https://plane.example", AuthMode: AuthModeAPIKey, APIKey: "key", RoundTripper: transport})
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.Do(context.Background(), http.MethodGet, "/widgets/", nil, nil, nil, nil)
	if err == nil || requests.Load() != 1 {
		t.Fatalf("request count = %d, error = %v", requests.Load(), err)
	}
}

func TestClientContextCancellationAndTimeout(t *testing.T) {
	server := testsupport.NewServer(t, func(writer http.ResponseWriter, request *http.Request) {
		<-request.Context().Done()
	})
	client, err := NewClient(Options{BaseURL: server.URL, AuthMode: AuthModeAPIKey, APIKey: "key", Timeout: time.Second})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = client.Do(ctx, http.MethodGet, "/widgets/", nil, nil, nil, nil)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation error = %v", err)
	}

	timeoutClient, err := NewClient(Options{
		BaseURL: server.URL, AuthMode: AuthModeAPIKey, APIKey: "key", Timeout: 50 * time.Millisecond,
		HTTPClient: &http.Client{Timeout: 10 * time.Millisecond},
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = timeoutClient.Do(context.Background(), http.MethodGet, "/widgets/", nil, nil, nil, nil)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("timeout error = %v", err)
	}
}

func TestClientRejectsCallerAuthenticationHeadersBeforeRequest(t *testing.T) {
	var requests atomic.Int32
	transport := roundTripFunc(func(request *http.Request) (*http.Response, error) {
		requests.Add(1)
		return nil, errors.New("should not be called")
	})
	client, err := NewClient(Options{BaseURL: "https://plane.example", AuthMode: AuthModeAPIKey, APIKey: "key", RoundTripper: transport})
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.Do(context.Background(), http.MethodGet, "/widgets/", nil, nil, http.Header{"Authorization": {"Bearer leaked"}}, nil)
	if err == nil || requests.Load() != 0 || strings.Contains(err.Error(), "leaked") {
		t.Fatalf("auth header handling error=%v requests=%d", err, requests.Load())
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

type trackingBody struct {
	reader io.Reader
	read   bool
	closed bool
}

func (b *trackingBody) Read(p []byte) (int, error) {
	b.read = true
	return b.reader.Read(p)
}

func (b *trackingBody) Close() error {
	b.closed = true
	return nil
}

func reflectDeepEqualValues(left, right url.Values) bool {
	if len(left) != len(right) {
		return false
	}
	for key, values := range right {
		leftValues, ok := left[key]
		if !ok || len(leftValues) != len(values) {
			return false
		}
		for index := range values {
			if leftValues[index] != values[index] {
				return false
			}
		}
	}
	return true
}
