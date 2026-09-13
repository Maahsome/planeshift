package plane

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"unicode/utf8"
)

const (
	maxErrorMessageBytes = 4096
	maxResponseBodyBytes = 16 << 20
)

// RateLimitMetadata contains only the two documented safe rate-limit headers.
// The raw values are retained even when a server sends a non-numeric value;
// Parsed flags distinguish an absent/unparseable value from numeric zero.
type RateLimitMetadata struct {
	Remaining       int    `json:"remaining,omitempty"`
	Reset           int64  `json:"reset,omitempty"`
	RemainingValue  string `json:"remaining_value,omitempty"`
	ResetValue      string `json:"reset_value,omitempty"`
	RemainingParsed bool   `json:"remaining_parsed,omitempty"`
	ResetParsed     bool   `json:"reset_parsed,omitempty"`
}

// PaginationMetadata contains the documented cursor envelope fields. Pointer
// fields preserve an explicit JSON null instead of turning it into a value.
type PaginationMetadata struct {
	NextCursor      *string         `json:"next_cursor"`
	PrevCursor      *string         `json:"prev_cursor"`
	NextPageResults *bool           `json:"next_page_results"`
	PrevPageResults *bool           `json:"prev_page_results"`
	Count           *int            `json:"count"`
	TotalPages      *int            `json:"total_pages"`
	TotalResults    *int            `json:"total_results"`
	ExtraStats      json.RawMessage `json:"extra_stats"`
}

// ResponseMetadata is the safe metadata returned with every successful or
// typed-error response. No arbitrary response headers are retained.
type ResponseMetadata struct {
	StatusCode int                 `json:"status_code"`
	RateLimit  RateLimitMetadata   `json:"rate_limit"`
	Pagination *PaginationMetadata `json:"pagination,omitempty"`
}

// Metadata is a short alias useful to resource packages.
type Metadata = ResponseMetadata

// Response is the non-sensitive response envelope. Decoded JSON is written to
// the destination supplied to Client.Do; status and safe metadata remain here.
type Response struct {
	StatusCode int                 `json:"status_code"`
	Metadata   ResponseMetadata    `json:"metadata"`
	RateLimit  RateLimitMetadata   `json:"rate_limit"`
	Pagination *PaginationMetadata `json:"pagination,omitempty"`
}

// Request describes the route-agnostic JSON request contract.
type Request struct {
	Method      string
	Path        string
	Query       url.Values
	Body        any
	Headers     http.Header
	Destination any
	Into        any
}

// PaginationOptions adds the shared pagination controls while preserving
// caller-provided resource-specific query values.
type PaginationOptions struct {
	PerPage int
	Cursor  string
	Fields  string
	Expand  string
}

// Pagination is a compatibility alias for the public option name used by
// resource packages.
type Pagination = PaginationOptions

// CursorPage is a generic cursor page. Use json.RawMessage as T when a future
// resource wants to preserve unknown result fields without a resource model.
type CursorPage[T any] struct {
	NextCursor      *string                    `json:"next_cursor"`
	PrevCursor      *string                    `json:"prev_cursor"`
	NextPageResults *bool                      `json:"next_page_results"`
	PrevPageResults *bool                      `json:"prev_page_results"`
	Count           *int                       `json:"count"`
	TotalPages      *int                       `json:"total_pages"`
	TotalResults    *int                       `json:"total_results"`
	ExtraStats      json.RawMessage            `json:"extra_stats"`
	Results         []T                        `json:"results"`
	Unknown         map[string]json.RawMessage `json:"-"`
}

// RawCursorPage preserves dynamic result documents and unknown top-level
// fields. It is the default page representation for resource-agnostic code.
type RawCursorPage = CursorPage[json.RawMessage]

// UnmarshalJSON keeps unknown page keys in Unknown while allowing the normal
// encoding/json decoder to retain nullable and dynamic values.
func (p *CursorPage[T]) UnmarshalJSON(data []byte) error {
	type cursorPage CursorPage[T]
	var decoded cursorPage
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	for _, key := range []string{
		"next_cursor", "prev_cursor", "next_page_results", "prev_page_results",
		"count", "total_pages", "total_results", "extra_stats", "results",
	} {
		delete(raw, key)
	}
	*p = CursorPage[T](decoded)
	if len(raw) > 0 {
		p.Unknown = raw
	}
	return nil
}

// APIError is a safe typed non-2xx response error. ErrorCode maps the
// documented JSON "error" field because Error is reserved for the Go error
// method; its JSON representation remains "error".
type APIError struct {
	StatusCode int    `json:"status_code"`
	ErrorCode  string `json:"error,omitempty"`
	Detail     string `json:"detail,omitempty"`
	Message    string `json:"message,omitempty"`
}

func (e *APIError) Error() string {
	if e == nil {
		return "plane API error"
	}
	message := e.Message
	if message == "" {
		message = e.Detail
	}
	if message == "" {
		message = e.ErrorCode
	}
	if message == "" {
		message = http.StatusText(e.StatusCode)
	}
	if message == "" {
		message = "request failed"
	}
	return fmt.Sprintf("plane API error (status %d): %s", e.StatusCode, message)
}

func (e *APIError) status() int {
	if e == nil {
		return 0
	}
	return e.StatusCode
}

func rateLimitMetadata(headers http.Header) RateLimitMetadata {
	metadata := RateLimitMetadata{
		RemainingValue: headers.Get("X-RateLimit-Remaining"),
		ResetValue:     headers.Get("X-RateLimit-Reset"),
	}
	if metadata.RemainingValue != "" {
		if value, err := strconv.Atoi(strings.TrimSpace(metadata.RemainingValue)); err == nil {
			metadata.Remaining = value
			metadata.RemainingParsed = true
		}
	}
	if metadata.ResetValue != "" {
		if value, err := strconv.ParseInt(strings.TrimSpace(metadata.ResetValue), 10, 64); err == nil {
			metadata.Reset = value
			metadata.ResetParsed = true
		}
	}
	return metadata
}

func paginationMetadata(data []byte) *PaginationMetadata {
	var envelope struct {
		NextCursor      *string         `json:"next_cursor"`
		PrevCursor      *string         `json:"prev_cursor"`
		NextPageResults *bool           `json:"next_page_results"`
		PrevPageResults *bool           `json:"prev_page_results"`
		Count           *int            `json:"count"`
		TotalPages      *int            `json:"total_pages"`
		TotalResults    *int            `json:"total_results"`
		ExtraStats      json.RawMessage `json:"extra_stats"`
	}
	if err := json.Unmarshal(data, &envelope); err != nil {
		return nil
	}
	if envelope.NextCursor == nil && envelope.PrevCursor == nil && envelope.NextPageResults == nil &&
		envelope.PrevPageResults == nil && envelope.Count == nil && envelope.TotalPages == nil &&
		envelope.TotalResults == nil && envelope.ExtraStats == nil {
		return nil
	}
	return &PaginationMetadata{
		NextCursor: envelope.NextCursor, PrevCursor: envelope.PrevCursor,
		NextPageResults: envelope.NextPageResults, PrevPageResults: envelope.PrevPageResults,
		Count: envelope.Count, TotalPages: envelope.TotalPages, TotalResults: envelope.TotalResults,
		ExtraStats: envelope.ExtraStats,
	}
}

func boundedMessage(data []byte, secrets ...string) string {
	message := strings.TrimSpace(string(data))
	for _, secret := range secrets {
		if secret != "" {
			message = strings.ReplaceAll(message, secret, "[REDACTED]")
		}
	}
	words := strings.Fields(message)
	for index, word := range words {
		for _, scheme := range []string{"https://", "http://"} {
			if urlIndex := strings.Index(word, scheme); urlIndex >= 0 {
				words[index] = word[:urlIndex] + "[URL REDACTED]"
				break
			}
		}
	}
	message = strings.Join(words, " ")
	if len(message) <= maxErrorMessageBytes {
		return message
	}
	truncated := message[:maxErrorMessageBytes-3]
	for len(truncated) > 0 && !utf8.ValidString(truncated) {
		truncated = truncated[:len(truncated)-1]
	}
	return truncated + "..."
}

func jsonErrorField(value json.RawMessage, secrets ...string) string {
	var text string
	if err := json.Unmarshal(value, &text); err != nil {
		return ""
	}
	return boundedMessage([]byte(text), secrets...)
}

func newAPIError(statusCode int, data []byte, secrets ...string) *APIError {
	apiError := &APIError{StatusCode: statusCode}
	var fields map[string]json.RawMessage
	if json.Unmarshal(data, &fields) == nil && fields != nil {
		apiError.ErrorCode = jsonErrorField(fields["error"], secrets...)
		apiError.Detail = jsonErrorField(fields["detail"], secrets...)
		apiError.Message = jsonErrorField(fields["message"], secrets...)
	}
	if apiError.Message == "" && apiError.Detail == "" && apiError.ErrorCode == "" {
		apiError.Message = boundedMessage(data, secrets...)
	}
	if apiError.Message == "" && apiError.Detail == "" && apiError.ErrorCode == "" {
		apiError.Message = http.StatusText(statusCode)
	}
	return apiError
}
