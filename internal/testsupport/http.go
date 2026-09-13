// Package testsupport contains deterministic helpers shared by repository
// tests and future resource slices. It is intentionally test-only support and
// has no runtime dependency.
package testsupport

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
)

// RequestExpectation describes the request properties a test cares about.
// ExpectedHeaders are exact case-insensitive header values; omitted headers
// can be asserted with AbsentHeaders.
type RequestExpectation struct {
	Method          string
	Path            string
	Query           url.Values
	ExpectedHeaders http.Header
	AbsentHeaders   []string
	Body            []byte
}

// NewServer starts a local deterministic HTTP server and closes it with the
// owning test.
func NewServer(t testing.TB, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return server
}

// AssertRequest verifies method, path, exact query values, selected headers,
// absent headers, and body bytes.
func AssertRequest(t testing.TB, request *http.Request, expected RequestExpectation) []byte {
	t.Helper()
	if expected.Method != "" && request.Method != expected.Method {
		t.Errorf("method = %q, want %q", request.Method, expected.Method)
	}
	if expected.Path != "" && request.URL.Path != expected.Path {
		t.Errorf("path = %q, want %q", request.URL.Path, expected.Path)
	}
	if expected.Query != nil && !reflect.DeepEqual(request.URL.Query(), expected.Query) {
		t.Errorf("query = %v, want %v", request.URL.Query(), expected.Query)
	}
	for key, values := range expected.ExpectedHeaders {
		actual := request.Header.Values(key)
		if !reflect.DeepEqual(actual, values) {
			t.Errorf("header %q = %v, want %v", key, actual, values)
		}
	}
	for _, key := range expected.AbsentHeaders {
		if request.Header.Get(key) != "" {
			t.Errorf("header %q = %q, want absent", key, request.Header.Get(key))
		}
	}
	body, err := io.ReadAll(request.Body)
	if err != nil {
		t.Errorf("read request body: %v", err)
		return nil
	}
	if expected.Body != nil && !bytes.Equal(body, expected.Body) {
		t.Errorf("body = %q, want %q", body, expected.Body)
	}
	return body
}

// JSON writes a JSON response with the supplied status.
func JSON(writer http.ResponseWriter, status int, value any) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(status)
	if value != nil {
		_ = json.NewEncoder(writer).Encode(value)
	}
}

// JSONHandler returns a reusable JSON response handler.
func JSONHandler(status int, value any) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		JSON(writer, status, value)
	}
}

// Empty writes an empty response with the supplied status.
func Empty(writer http.ResponseWriter, status int) {
	writer.WriteHeader(status)
}

// EmptyHandler returns a reusable empty response handler.
func EmptyHandler(status int) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		Empty(writer, status)
	}
}

// MalformedJSON writes a deliberately malformed JSON body.
func MalformedJSON(writer http.ResponseWriter, status int) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(status)
	_, _ = io.WriteString(writer, `{"malformed":`)
}

// MalformedJSONHandler returns a malformed JSON response handler.
func MalformedJSONHandler(status int) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		MalformedJSON(writer, status)
	}
}

// NonJSON writes a plain-text response body.
func NonJSON(writer http.ResponseWriter, status int, message string) {
	writer.Header().Set("Content-Type", "text/plain")
	writer.WriteHeader(status)
	_, _ = io.WriteString(writer, message)
}

// NonJSONHandler returns a reusable plain-text response handler.
func NonJSONHandler(status int, message string) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		NonJSON(writer, status, message)
	}
}

// NoContent writes a successful 204 response without attempting to write a
// body, matching Plane's documented delete behavior.
func NoContent(writer http.ResponseWriter) {
	writer.WriteHeader(http.StatusNoContent)
}

// NoContentHandler returns a reusable 204 response handler.
func NoContentHandler() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		NoContent(writer)
	}
}

// JSONBodyEqual is useful when the wire body may differ only in whitespace.
func JSONBodyEqual(t testing.TB, actual []byte, expected any) {
	t.Helper()
	var actualValue any
	if err := json.Unmarshal(actual, &actualValue); err != nil {
		t.Errorf("decode actual JSON: %v", err)
		return
	}
	expectedBytes, err := json.Marshal(expected)
	if err != nil {
		t.Errorf("encode expected JSON: %v", err)
		return
	}
	var expectedValue any
	if err := json.Unmarshal(expectedBytes, &expectedValue); err != nil {
		t.Errorf("decode expected JSON: %v", err)
		return
	}
	if !reflect.DeepEqual(actualValue, expectedValue) {
		t.Errorf("JSON body = %s, want %s", strings.TrimSpace(string(actual)), expectedBytes)
	}
}
