package plane

import (
	"context"
	"errors"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"planeshift/internal/testsupport"
)

func TestUploadStreamsRawBodyWithoutPlaneAuthentication(t *testing.T) {
	server := testsupport.NewServer(t, func(writer http.ResponseWriter, request *http.Request) {
		testsupport.AssertRequest(t, request, testsupport.RequestExpectation{
			Method: http.MethodPut,
			Path:   "/storage/object",
			Query:  url.Values{"signature": {"signed-value"}},
			ExpectedHeaders: http.Header{
				"Content-Type": {"application/octet-stream"},
				"X-Storage":    {"storage-header"},
			},
			AbsentHeaders: []string{"X-API-Key", "Authorization"},
			Body:          []byte("streamed content"),
		})
		testsupport.Empty(writer, http.StatusCreated)
	})
	client, err := NewClient(Options{BaseURL: "https://plane.example", AuthMode: AuthModeAPIKey, APIKey: "plane-secret"})
	if err != nil {
		t.Fatal(err)
	}
	response, err := client.Upload(context.Background(), UploadRequest{
		Method: http.MethodPut,
		URL:    server.URL + "/storage/object?signature=signed-value",
		Headers: http.Header{
			"Content-Type": {"application/octet-stream"},
			"X-Storage":    {"storage-header"},
		},
		Body: strings.NewReader("streamed content"),
	})
	if err != nil || response.StatusCode != http.StatusCreated {
		t.Fatalf("Upload() = %#v, %v", response, err)
	}
}

func TestUploadBuildsStreamingMultipartWithReturnedFields(t *testing.T) {
	server := testsupport.NewServer(t, func(writer http.ResponseWriter, request *http.Request) {
		contentType := request.Header.Get("Content-Type")
		mediaType, parameters, err := mime.ParseMediaType(contentType)
		if err != nil || mediaType != "multipart/form-data" || parameters["boundary"] == "" {
			t.Errorf("Content-Type = %q, parse error = %v", contentType, err)
			return
		}
		if request.Header.Get("X-API-Key") != "" || request.Header.Get("Authorization") != "" {
			t.Errorf("upload carried Plane auth headers")
		}
		reader := multipart.NewReader(request.Body, parameters["boundary"])
		values := map[string]string{}
		fileData := ""
		fileName := ""
		for {
			part, readErr := reader.NextPart()
			if errors.Is(readErr, io.EOF) {
				break
			}
			if readErr != nil {
				t.Errorf("read multipart part: %v", readErr)
				return
			}
			data, readErr := io.ReadAll(part)
			if readErr != nil {
				t.Errorf("read multipart data: %v", readErr)
				return
			}
			if part.FileName() != "" {
				fileData = string(data)
				fileName = part.FileName()
			} else {
				values[part.FormName()] = string(data)
			}
		}
		if values["policy"] != "signed-policy" || values["key"] != "object-key" ||
			fileName != "payload.txt" || fileData != "uploaded bytes" {
			t.Errorf("multipart values=%v file=%q filename=%q", values, fileData, fileName)
		}
		testsupport.Empty(writer, http.StatusNoContent)
	})
	client, err := NewClient(Options{BaseURL: "https://plane.example", AuthMode: AuthModeBearer, BearerToken: "plane-token"})
	if err != nil {
		t.Fatal(err)
	}
	response, err := client.Upload(context.Background(), UploadRequest{
		Method: http.MethodPost,
		URL:    server.URL + "/presigned/upload?X-Amz-Signature=signature",
		FormFields: url.Values{
			"key": {"object-key"}, "policy": {"signed-policy"},
		},
		File:      strings.NewReader("uploaded bytes"),
		FileField: "file",
		FileName:  "payload.txt",
	})
	if err != nil || response.StatusCode != http.StatusNoContent {
		t.Fatalf("Upload() = %#v, %v", response, err)
	}
}

func TestUploadErrorsAreSafeAndContextAware(t *testing.T) {
	secretURL := "https://storage.example/presigned?signature=secret"
	server := testsupport.NewServer(t, func(writer http.ResponseWriter, request *http.Request) {
		testsupport.NonJSON(writer, http.StatusBadRequest, "form=secret-field url="+secretURL+" content=secret-content")
	})
	client, err := NewClient(Options{BaseURL: "https://plane.example", AuthMode: AuthModeAPIKey, APIKey: "plane-key"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.Upload(context.Background(), UploadRequest{
		Method:     http.MethodPost,
		URL:        server.URL + "/upload",
		FormFields: url.Values{"form": {"secret-field"}},
		Body:       strings.NewReader("secret-content"),
	})
	if err == nil || !strings.Contains(err.Error(), "400") || strings.Contains(err.Error(), secretURL) ||
		strings.Contains(err.Error(), "secret-field") || strings.Contains(err.Error(), "secret-content") {
		t.Fatalf("unsafe upload error = %v", err)
	}

	blockingTransport := roundTripFunc(func(request *http.Request) (*http.Response, error) {
		<-request.Context().Done()
		return nil, request.Context().Err()
	})
	blockingClient, err := NewClient(Options{
		BaseURL: "https://plane.example", AuthMode: AuthModeAPIKey, APIKey: "key", RoundTripper: blockingTransport,
	})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = blockingClient.Upload(ctx, UploadRequest{Method: http.MethodPut, URL: "https://storage.example/upload", Body: strings.NewReader("data")})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("upload cancellation error = %v", err)
	}

	if _, err := client.Upload(context.Background(), UploadRequest{
		Method: http.MethodPut, URL: "https://storage.example/upload", Headers: http.Header{"Authorization": {"Bearer leaked"}},
	}); err == nil || strings.Contains(err.Error(), "leaked") {
		t.Fatalf("upload auth header error = %v", err)
	}

	timeoutClient, err := NewClient(Options{
		BaseURL: "https://plane.example", AuthMode: AuthModeAPIKey, APIKey: "key",
		HTTPClient: &http.Client{Timeout: 10 * time.Millisecond}, RoundTripper: blockingTransport,
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = timeoutClient.Upload(context.Background(), UploadRequest{Method: http.MethodPut, URL: "https://storage.example/upload", Body: strings.NewReader("data")})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("upload timeout error = %v", err)
	}
}
