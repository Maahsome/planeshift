package plane

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"sort"
	"strings"
)

// UploadRequest describes the response-driven upload contract. URL is an
// absolute storage URL; it is never joined to the Plane base and Plane auth is
// never added. Body is streamed as-is unless form fields or Multipart require
// multipart framing.
type UploadRequest struct {
	Method  string
	URL     string
	Headers http.Header

	FormFields url.Values
	// Fields is a convenience for callers whose upload response contains a
	// single value per form key.
	Fields map[string]string

	Body io.Reader
	File io.Reader

	FileField string
	FileName  string
	Multipart bool
}

// UploadResponse contains only safe status information from an upload.
type UploadResponse struct {
	StatusCode int
}

// Upload sends one response-driven upload request using the configured HTTP
// transport and timeout. It does not use the Plane base URL or credentials.
func (c *HTTPClient) Upload(ctx context.Context, request UploadRequest) (UploadResponse, error) {
	if c == nil || c.httpClient == nil {
		return UploadResponse{}, fmt.Errorf("plane upload client is not initialized")
	}
	if ctx == nil {
		return UploadResponse{}, fmt.Errorf("plane upload context is nil")
	}
	if strings.TrimSpace(request.Method) == "" {
		return UploadResponse{}, fmt.Errorf("plane upload method is required")
	}
	u, err := validateUploadURL(request.URL)
	if err != nil {
		return UploadResponse{}, err
	}
	if hasUploadAuthentication(request.Headers) {
		return UploadResponse{}, fmt.Errorf("Plane authentication is not allowed on upload requests")
	}

	body, contentType, err := uploadBody(request)
	if err != nil {
		return UploadResponse{}, err
	}
	httpRequest, err := http.NewRequestWithContext(ctx, request.Method, u.String(), body)
	if err != nil {
		return UploadResponse{}, fmt.Errorf("create Plane upload request")
	}
	for key, values := range request.Headers {
		for _, value := range values {
			httpRequest.Header.Add(key, value)
		}
	}
	if contentType != "" && httpRequest.Header.Get("Content-Type") == "" {
		httpRequest.Header.Set("Content-Type", contentType)
	}

	httpResponse, err := c.httpClient.Do(httpRequest)
	if err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return UploadResponse{}, ctxErr
		}
		if errors.Is(err, context.Canceled) {
			return UploadResponse{}, context.Canceled
		}
		if errors.Is(err, context.DeadlineExceeded) {
			return UploadResponse{}, context.DeadlineExceeded
		}
		return UploadResponse{}, fmt.Errorf("Plane upload failed")
	}
	if httpResponse.Body == nil {
		httpResponse.Body = io.NopCloser(strings.NewReader(""))
	}
	defer httpResponse.Body.Close()
	response := UploadResponse{StatusCode: httpResponse.StatusCode}
	if httpResponse.StatusCode >= 200 && httpResponse.StatusCode < 300 {
		return response, nil
	}

	// Do not retain or echo an upload response body: a storage service may
	// repeat form fields or uploaded content in a failure response.
	_, _ = io.Copy(io.Discard, io.LimitReader(httpResponse.Body, maxErrorMessageBytes))
	return response, &UploadError{StatusCode: httpResponse.StatusCode}
}

// UploadError intentionally contains no URL, form fields, headers, or body.
type UploadError struct {
	StatusCode int
}

func (e *UploadError) Error() string {
	if e == nil {
		return "plane upload error"
	}
	return fmt.Sprintf("plane upload failed (status %d)", e.StatusCode)
}

func validateUploadURL(raw string) (*url.URL, error) {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || u.Scheme == "" || u.Host == "" || u.Opaque != "" || u.User != nil || u.Fragment != "" ||
		(u.Scheme != "http" && u.Scheme != "https") {
		return nil, fmt.Errorf("invalid Plane upload URL")
	}
	return u, nil
}

func hasUploadAuthentication(headers http.Header) bool {
	for key := range headers {
		if strings.EqualFold(key, "Authorization") || strings.EqualFold(key, "X-API-Key") ||
			strings.EqualFold(key, "Proxy-Authorization") {
			return true
		}
	}
	return false
}

func uploadBody(request UploadRequest) (io.Reader, string, error) {
	file := request.File
	if file == nil {
		file = request.Body
	}
	fields := copyValues(request.FormFields)
	if fields == nil {
		fields = make(url.Values)
	}
	for key, value := range request.Fields {
		fields.Set(key, value)
	}

	if len(fields) == 0 && !request.Multipart {
		return file, "", nil
	}

	prefix := &bytes.Buffer{}
	writer := multipart.NewWriter(prefix)
	keys := make([]string, 0, len(fields))
	for key := range fields {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		values := fields[key]
		for _, value := range values {
			if err := writer.WriteField(key, value); err != nil {
				return nil, "", fmt.Errorf("encode Plane upload form")
			}
		}
	}
	if file != nil {
		field := request.FileField
		if field == "" {
			field = "file"
		}
		fileName := request.FileName
		if fileName == "" {
			fileName = "upload"
		}
		if _, err := writer.CreateFormFile(field, fileName); err != nil {
			return nil, "", fmt.Errorf("encode Plane upload file part")
		}
	}
	boundary := writer.Boundary()
	// CreateFormFile wrote only the part header to prefix. The file reader is
	// placed between that header and the closing boundary, so content is never
	// buffered in memory.
	suffix := []byte("\r\n--" + boundary + "--\r\n")
	contentType := writer.FormDataContentType()
	if file == nil {
		file = bytes.NewReader(nil)
	}
	return io.MultiReader(bytes.NewReader(prefix.Bytes()), file, bytes.NewReader(suffix)), contentType, nil
}
