package config

import (
	"fmt"
	"net/url"
	"strings"
	"time"
)

const (
	// DefaultPlaneAPIURL is the public Plane Cloud host. The plane package
	// appends /api/v1/ when it normalizes this value for requests.
	DefaultPlaneAPIURL = "https://api.plane.so"

	// DefaultPlaneTimeout bounds a request when no timeout is configured.
	DefaultPlaneTimeout = 30 * time.Second

	// MaxPlaneTimeout prevents a typo from creating an effectively unbounded
	// resource request while still allowing slow self-hosted installations.
	MaxPlaneTimeout = 10 * time.Minute

	AuthModeAPIKey = "api-key"
	AuthModeBearer = "bearer"
)

// PlaneSettings contains the settings required to create a Plane API client.
// Credentials are intentionally excluded from JSON output and String output.
// A zero APIURL or Timeout receives the documented default during normalization;
// credentials and AuthMode remain required for resource requests.
type PlaneSettings struct {
	APIURL      string        `mapstructure:"api_url" json:"-"`
	AuthMode    string        `mapstructure:"auth_mode" json:"-"`
	APIKey      string        `mapstructure:"api_key" json:"-"`
	BearerToken string        `mapstructure:"bearer_token" json:"-"`
	Timeout     time.Duration `mapstructure:"timeout" json:"-"`
}

// Normalize returns a copy with whitespace removed from scalar settings and
// defaults applied to the optional host and timeout.
func (s PlaneSettings) Normalize() PlaneSettings {
	s.APIURL = strings.TrimSpace(s.APIURL)
	if s.APIURL == "" {
		s.APIURL = DefaultPlaneAPIURL
	}
	s.AuthMode = strings.ToLower(strings.TrimSpace(s.AuthMode))
	s.APIKey = strings.TrimSpace(s.APIKey)
	s.BearerToken = strings.TrimSpace(s.BearerToken)
	if s.Timeout == 0 {
		s.Timeout = DefaultPlaneTimeout
	}
	return s
}

// Validate checks settings that must be correct before a resource request is
// sent. The caller must select one credential mode explicitly; the client does
// not infer a mode from whichever secret happens to be present.
func (s PlaneSettings) Validate() error {
	s = s.Normalize()

	if s.AuthMode != AuthModeAPIKey && s.AuthMode != AuthModeBearer {
		return fmt.Errorf("plane auth mode must be %q or %q", AuthModeAPIKey, AuthModeBearer)
	}
	if s.APIKey != "" && s.BearerToken != "" {
		return fmt.Errorf("plane API key and bearer token cannot both be configured")
	}
	if s.AuthMode == AuthModeAPIKey && s.APIKey == "" {
		return fmt.Errorf("plane API key is required for %q auth", AuthModeAPIKey)
	}
	if s.AuthMode == AuthModeBearer && s.BearerToken == "" {
		return fmt.Errorf("plane bearer token is required for %q auth", AuthModeBearer)
	}
	if s.Timeout <= 0 {
		return fmt.Errorf("plane timeout must be greater than zero")
	}
	if s.Timeout > MaxPlaneTimeout {
		return fmt.Errorf("plane timeout must not exceed %s", MaxPlaneTimeout)
	}
	return nil
}

// String intentionally reports only non-sensitive configuration details.
func (s PlaneSettings) String() string {
	s = s.Normalize()
	apiURL := "[invalid]"
	if parsed, err := url.Parse(s.APIURL); err == nil && parsed.User == nil && parsed.RawQuery == "" &&
		parsed.Fragment == "" && parsed.Scheme != "" && parsed.Host != "" {
		apiURL = parsed.Scheme + "://" + parsed.Host + parsed.EscapedPath()
	}
	return fmt.Sprintf("PlaneSettings{api_url:%q auth_mode:%q timeout:%s}", apiURL, s.AuthMode, s.Timeout)
}

// ParsePlaneTimeout parses the duration syntax used by PLANE_TIMEOUT and the
// plane.timeout configuration key.
func ParsePlaneTimeout(value string) (time.Duration, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return DefaultPlaneTimeout, nil
	}
	duration, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("invalid plane timeout %q: %w", value, err)
	}
	if duration <= 0 {
		return 0, fmt.Errorf("plane timeout must be greater than zero")
	}
	if duration > MaxPlaneTimeout {
		return 0, fmt.Errorf("plane timeout must not exceed %s", MaxPlaneTimeout)
	}
	return duration, nil
}
