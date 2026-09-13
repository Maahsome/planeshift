package plane

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// Client is the deliberately small seam consumed by future resource
// packages. A resource supplies its documented method/path/query/body; the
// foundation owns URL joining, auth, JSON, metadata, and error handling.
type Client interface {
	Do(ctx context.Context, method, route string, query url.Values, body any, headers http.Header, destination any) (Response, error)
}

// UploadClient is kept separate from Client so resource code that does not
// upload files can use the smaller request seam.
type UploadClient interface {
	Upload(ctx context.Context, request UploadRequest) (UploadResponse, error)
}

// HTTPClient is the standard-library implementation of Client and
// UploadClient.
type HTTPClient struct {
	baseURL    *url.URL
	httpClient *http.Client
	authMode   string
	apiKey     string
	bearer     string
}

var _ Client = (*HTTPClient)(nil)
var _ UploadClient = (*HTTPClient)(nil)

// NewClient validates options and constructs a synchronous client. No network
// request is made until Do or Upload is called.
func NewClient(options Options) (*HTTPClient, error) {
	options, err := options.normalized()
	if err != nil {
		return nil, err
	}
	baseURL, err := normalizeBaseURL(options.BaseURL)
	if err != nil {
		return nil, err
	}

	httpClient := &http.Client{}
	if options.HTTPClient != nil {
		copy := *options.HTTPClient
		httpClient = &copy
	}
	transport := options.RoundTripper
	if transport == nil {
		transport = options.Transport
	}
	if transport != nil {
		httpClient.Transport = transport
	}
	if httpClient.Timeout == 0 || httpClient.Timeout > options.Timeout {
		httpClient.Timeout = options.Timeout
	}

	return &HTTPClient{
		baseURL: baseURL, httpClient: httpClient,
		authMode: options.AuthMode, apiKey: options.APIKey, bearer: options.BearerToken,
	}, nil
}

// Do sends exactly one authenticated JSON request.
func (c *HTTPClient) Do(ctx context.Context, method, route string, query url.Values, body any, headers http.Header, destination any) (Response, error) {
	return c.do(ctx, Request{
		Method: method, Path: route, Query: query, Body: body,
		Headers: headers, Destination: destination,
	})
}

// DoRequest is the struct-shaped equivalent of Do for callers that prefer a
// named request object.
func (c *HTTPClient) DoRequest(ctx context.Context, request Request) (Response, error) {
	return c.do(ctx, request)
}

// Get is a convenience method for the common no-body GET operation.
func (c *HTTPClient) Get(ctx context.Context, route string, query url.Values, headers http.Header, destination any) (Response, error) {
	return c.Do(ctx, http.MethodGet, route, query, nil, headers, destination)
}

func (c *HTTPClient) do(ctx context.Context, request Request) (Response, error) {
	if c == nil || c.httpClient == nil || c.baseURL == nil {
		return Response{}, fmt.Errorf("plane client is not initialized")
	}
	if ctx == nil {
		return Response{}, fmt.Errorf("plane request context is nil")
	}
	if strings.TrimSpace(request.Method) == "" {
		return Response{}, fmt.Errorf("plane request method is required")
	}

	requestURL, err := BuildURL(c.baseURL.String(), request.Path, request.Query)
	if err != nil {
		return Response{}, err
	}

	var bodyReader io.Reader
	var bodyPresent bool
	if request.Body != nil {
		body, marshalErr := json.Marshal(request.Body)
		if marshalErr != nil {
			return Response{}, fmt.Errorf("encode Plane JSON request body")
		}
		bodyReader = bytes.NewReader(body)
		bodyPresent = true
	}

	httpRequest, err := http.NewRequestWithContext(ctx, request.Method, requestURL, bodyReader)
	if err != nil {
		return Response{}, fmt.Errorf("create Plane request")
	}
	if err := copySafeHeaders(httpRequest.Header, request.Headers); err != nil {
		return Response{}, err
	}
	httpRequest.Header.Set("Accept", "application/json")
	if bodyPresent {
		httpRequest.Header.Set("Content-Type", "application/json")
	}
	if err := c.setAuthentication(httpRequest.Header); err != nil {
		return Response{}, err
	}

	httpResponse, err := c.httpClient.Do(httpRequest)
	if err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return Response{}, ctxErr
		}
		if errors.Is(err, context.Canceled) {
			return Response{}, context.Canceled
		}
		if errors.Is(err, context.DeadlineExceeded) {
			return Response{}, context.DeadlineExceeded
		}
		return Response{}, fmt.Errorf("Plane request failed: %s", boundedMessage([]byte(err.Error()), c.apiKey, c.bearer))
	}
	if httpResponse.Body == nil {
		httpResponse.Body = io.NopCloser(strings.NewReader(""))
	}
	defer httpResponse.Body.Close()

	metadata := ResponseMetadata{
		StatusCode: httpResponse.StatusCode,
		RateLimit:  rateLimitMetadata(httpResponse.Header),
	}
	response := Response{
		StatusCode: httpResponse.StatusCode,
		Metadata:   metadata,
		RateLimit:  metadata.RateLimit,
	}

	if httpResponse.StatusCode == http.StatusNoContent {
		return response, nil
	}

	readLimit := int64(maxResponseBodyBytes)
	if httpResponse.StatusCode < 200 || httpResponse.StatusCode >= 300 {
		readLimit = maxErrorMessageBytes + 1
	}
	data, err := io.ReadAll(io.LimitReader(httpResponse.Body, readLimit))
	if err != nil {
		return response, fmt.Errorf("read Plane response body")
	}
	if int64(len(data)) == readLimit && readLimit != maxResponseBodyBytes {
		data = data[:maxErrorMessageBytes]
	}

	if httpResponse.StatusCode < 200 || httpResponse.StatusCode >= 300 {
		return response, newAPIError(httpResponse.StatusCode, data, c.apiKey, c.bearer)
	}
	if len(strings.TrimSpace(string(data))) == 0 {
		return response, nil
	}
	if !json.Valid(data) {
		return response, fmt.Errorf("decode Plane JSON response: malformed JSON")
	}
	response.Metadata.Pagination = paginationMetadata(data)
	response.Pagination = response.Metadata.Pagination
	if request.destination() != nil {
		if err := json.Unmarshal(data, request.destination()); err != nil {
			return response, fmt.Errorf("decode Plane JSON response")
		}
	}
	return response, nil
}

func (r Request) destination() any {
	if r.Destination != nil {
		return r.Destination
	}
	return r.Into
}

func copySafeHeaders(destination http.Header, source http.Header) error {
	for key, values := range source {
		if isForbiddenRequestHeader(key) {
			return fmt.Errorf("Plane request header %q is managed by the client", key)
		}
		for _, value := range values {
			destination.Add(key, value)
		}
	}
	return nil
}

func isForbiddenRequestHeader(key string) bool {
	switch {
	case strings.EqualFold(key, "Authorization"):
		return true
	case strings.EqualFold(key, "X-API-Key"):
		return true
	case strings.EqualFold(key, "Proxy-Authorization"):
		return true
	case strings.EqualFold(key, "Cookie"):
		return true
	default:
		return false
	}
}

func (c *HTTPClient) setAuthentication(headers http.Header) error {
	switch c.authMode {
	case AuthModeAPIKey:
		headers.Set("X-API-Key", c.apiKey)
	case AuthModeBearer:
		headers.Set("Authorization", "Bearer "+c.bearer)
	default:
		return fmt.Errorf("plane auth mode is invalid")
	}
	return nil
}
