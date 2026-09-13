// Package plane contains the shared, route-agnostic Plane HTTP foundation.
// Resource packages provide explicit documented paths; this package owns only
// transport, authentication, JSON, pagination, upload, and safe error rules.
package plane

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

const (
	// DefaultBaseURL is the Plane Cloud host before the /api/v1/ prefix.
	DefaultBaseURL = "https://api.plane.so"

	// DefaultTimeout bounds requests when callers omit a timeout.
	DefaultTimeout = 30 * time.Second

	// MaxTimeout is the largest request timeout accepted by the foundation.
	MaxTimeout = 10 * time.Minute

	AuthModeAPIKey = "api-key"
	AuthModeBearer = "bearer"
)

// Options configures a Plane client. HTTPClient and RoundTripper/Transport are
// injectable so tests and future command packages can use deterministic local
// servers without changing request behavior.
type Options struct {
	BaseURL string
	// APIURL is accepted as an alternate spelling for callers that use the
	// configuration field name. BaseURL takes precedence when both are set.
	APIURL string

	AuthMode    string
	APIKey      string
	BearerToken string
	Timeout     time.Duration

	HTTPClient   *http.Client
	RoundTripper http.RoundTripper
	Transport    http.RoundTripper
}

func (o Options) normalized() (Options, error) {
	if strings.TrimSpace(o.BaseURL) == "" {
		o.BaseURL = o.APIURL
	}
	normalizedURL, err := normalizeBaseURL(o.BaseURL)
	if err != nil {
		return Options{}, err
	}
	o.BaseURL = normalizedURL.String()
	o.AuthMode = strings.ToLower(strings.TrimSpace(o.AuthMode))
	o.APIKey = strings.TrimSpace(o.APIKey)
	o.BearerToken = strings.TrimSpace(o.BearerToken)
	if o.Timeout == 0 {
		o.Timeout = DefaultTimeout
	}
	if o.Timeout <= 0 {
		return Options{}, fmt.Errorf("plane timeout must be greater than zero")
	}
	if o.Timeout > MaxTimeout {
		return Options{}, fmt.Errorf("plane timeout must not exceed %s", MaxTimeout)
	}
	if o.AuthMode != AuthModeAPIKey && o.AuthMode != AuthModeBearer {
		return Options{}, fmt.Errorf("plane auth mode must be %q or %q", AuthModeAPIKey, AuthModeBearer)
	}
	if o.APIKey != "" && o.BearerToken != "" {
		return Options{}, fmt.Errorf("plane API key and bearer token cannot both be configured")
	}
	if o.AuthMode == AuthModeAPIKey && o.APIKey == "" {
		return Options{}, fmt.Errorf("plane API key is required for %q auth", AuthModeAPIKey)
	}
	if o.AuthMode == AuthModeBearer && o.BearerToken == "" {
		return Options{}, fmt.Errorf("plane bearer token is required for %q auth", AuthModeBearer)
	}
	if o.RoundTripper != nil && o.Transport != nil && o.RoundTripper != o.Transport {
		return Options{}, fmt.Errorf("plane RoundTripper and Transport cannot both be set")
	}
	return o, nil
}

// String reports safe option metadata without either credential.
func (o Options) String() string {
	baseURL := "[invalid]"
	if normalized, err := NormalizeBaseURL(o.BaseURL); err == nil {
		baseURL = normalized
	}
	return fmt.Sprintf("Options{base_url:%q auth_mode:%q timeout:%s}", baseURL, o.AuthMode, o.Timeout)
}

// ClientFactory lazily constructs a Client. A factory is a function type so
// command packages can inject a deterministic constructor without depending on
// Viper or constructing transports themselves.
type ClientFactory func() (Client, error)

// NewClientFactory returns a lazy factory for options. Options are validated at
// factory invocation time, not when the command tree is initialized.
func NewClientFactory(options Options) ClientFactory {
	return func() (Client, error) {
		return NewClient(options)
	}
}

// StaticClientFactory adapts an already constructed client for tests or a
// command composition root.
func StaticClientFactory(client Client) ClientFactory {
	return func() (Client, error) {
		if client == nil {
			return nil, fmt.Errorf("plane client is nil")
		}
		return client, nil
	}
}

// New creates a client through the factory. It returns a useful error for a
// nil factory rather than panicking in a future resource command.
func (f ClientFactory) New() (Client, error) {
	if f == nil {
		return nil, fmt.Errorf("plane client factory is nil")
	}
	return f()
}
