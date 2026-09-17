package functional

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

const (
	functionalOptIn       = "PLANE_FUNCTIONAL_RUN"
	functionalWorkspace   = "PLANE_FUNCTIONAL_WORKSPACE_SLUG"
	functionalLegacy      = "PLANE_FUNCTIONAL_INCLUDE_LEGACY"
	functionalTemplate    = "PLANE_FUNCTIONAL_PROJECT_TEMPLATE_ID"
	functionalAllowProd   = "PLANE_FUNCTIONAL_ALLOW_PRODUCTION"
	functionalBinary      = "PLANESHIFT_BINARY"
	functionalDefaultWait = 90 * time.Second
)

// FunctionalConfig contains only the settings that the functional subprocess
// is allowed to inherit. Secrets are retained solely for redaction and are
// never included in command arguments or test diagnostics.
type FunctionalConfig struct {
	Binary          string
	APIURL          string
	AuthMode        string
	APIKey          string
	BearerToken     string
	Timeout         time.Duration
	WorkspaceSlug   string
	IncludeLegacy   bool
	TemplateID      string
	AllowProduction bool
}

type commandResult struct {
	Args     []string
	Stdout   string
	Stderr   string
	ExitCode int
	Err      error
}

// Runner invokes the built CLI directly. It deliberately does not use a
// shell, pipelines, or resource clients so the real executable path remains
// under test.
type Runner struct {
	binary  string
	env     []string
	secrets []string
	timeout time.Duration
}

func discoveryRunner(t *testing.T) (*Runner, bool) {
	t.Helper()
	binary := strings.TrimSpace(os.Getenv(functionalBinary))
	if binary == "" {
		for _, candidate := range []string{"../planeshift", "./planeshift"} {
			if stat, err := os.Stat(candidate); err == nil && !stat.IsDir() {
				binary = candidate
				break
			}
		}
	}
	if binary == "" {
		t.Skip("functional command discovery requires PLANESHIFT_BINARY or a built ../planeshift binary")
		return nil, false
	}
	runner, err := newRunner(t, FunctionalConfig{Binary: binary})
	if err != nil {
		t.Fatalf("configure discovery runner: %v", err)
	}
	return runner, true
}

func loadFunctionalConfig() (FunctionalConfig, error) {
	config := FunctionalConfig{
		Binary:          strings.TrimSpace(os.Getenv(functionalBinary)),
		APIURL:          strings.TrimSpace(os.Getenv("PLANE_API_URL")),
		AuthMode:        strings.ToLower(strings.TrimSpace(os.Getenv("PLANE_AUTH_MODE"))),
		APIKey:          strings.TrimSpace(os.Getenv("PLANE_API_KEY")),
		BearerToken:     strings.TrimSpace(os.Getenv("PLANE_BEARER_TOKEN")),
		WorkspaceSlug:   strings.TrimSpace(os.Getenv(functionalWorkspace)),
		IncludeLegacy:   isTrue(os.Getenv(functionalLegacy)),
		TemplateID:      strings.TrimSpace(os.Getenv(functionalTemplate)),
		AllowProduction: isTrue(os.Getenv(functionalAllowProd)),
	}

	if config.Binary == "" {
		return config, errors.New("PLANESHIFT_BINARY is required")
	}
	if config.APIURL == "" {
		return config, errors.New("PLANE_API_URL is required; mutating tests never use the CLI default")
	}
	parsed, err := url.Parse(config.APIURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return config, errors.New("PLANE_API_URL must be an absolute URL without credentials, query, or fragment")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return config, errors.New("PLANE_API_URL must use http or https")
	}
	if isProductionHost(parsed.Hostname()) && !config.AllowProduction {
		return config, fmt.Errorf("PLANE_API_URL targets Plane Cloud production; use Prism/non-production or set %s=true after team review", functionalAllowProd)
	}
	if config.AuthMode != "api-key" && config.AuthMode != "bearer" {
		return config, errors.New("PLANE_AUTH_MODE must be api-key or bearer")
	}
	if config.APIKey != "" && config.BearerToken != "" {
		return config, errors.New("PLANE_API_KEY and PLANE_BEARER_TOKEN cannot both be configured")
	}
	if config.AuthMode == "api-key" && config.APIKey == "" {
		return config, errors.New("PLANE_API_KEY is required for api-key auth")
	}
	if config.AuthMode == "bearer" && config.BearerToken == "" {
		return config, errors.New("PLANE_BEARER_TOKEN is required for bearer auth")
	}
	config.Timeout, err = parseFunctionalTimeout(os.Getenv("PLANE_TIMEOUT"))
	if err != nil {
		return config, err
	}
	if config.WorkspaceSlug == "" {
		return config, fmt.Errorf("%s is required", functionalWorkspace)
	}
	if err := validateOptionalBool(functionalLegacy); err != nil {
		return config, err
	}
	return config, nil
}

func parseFunctionalTimeout(value string) (time.Duration, error) {
	if strings.TrimSpace(value) == "" {
		return 30 * time.Second, nil
	}
	duration, err := time.ParseDuration(strings.TrimSpace(value))
	if err != nil || duration <= 0 || duration > 10*time.Minute {
		return 0, errors.New("PLANE_TIMEOUT must be a duration greater than zero and no greater than 10m")
	}
	return duration, nil
}

func validateOptionalBool(name string) error {
	value := strings.TrimSpace(os.Getenv(name))
	if value != "" && !isTrue(value) && !strings.EqualFold(value, "false") {
		return fmt.Errorf("%s must be true or false when set", name)
	}
	return nil
}

func isTrue(value string) bool { return strings.EqualFold(strings.TrimSpace(value), "true") }

func isProductionHost(host string) bool {
	host = strings.ToLower(strings.TrimSpace(host))
	return host == "api.plane.so" || host == "app.plane.so" || host == "plane.so" || strings.HasSuffix(host, ".plane.so")
}

func newRunner(t *testing.T, config FunctionalConfig) (*Runner, error) {
	t.Helper()
	if config.Binary == "" {
		return nil, errors.New("binary path is empty")
	}
	configHome := t.TempDir()
	env := safeSubprocessEnvironment(config, configHome)
	secrets := make([]string, 0, 2)
	if config.APIKey != "" {
		secrets = append(secrets, config.APIKey)
	}
	if config.BearerToken != "" {
		secrets = append(secrets, config.BearerToken)
	}
	timeout := config.Timeout
	if timeout <= 0 || timeout > functionalDefaultWait {
		timeout = functionalDefaultWait
	}
	return &Runner{binary: config.Binary, env: env, secrets: secrets, timeout: timeout}, nil
}

func safeSubprocessEnvironment(config FunctionalConfig, configHome string) []string {
	allowed := map[string]bool{
		"HOME": true, "LANG": true, "LC_ALL": true, "LC_CTYPE": true,
		"PATH": true, "SYSTEMROOT": true, "TERM": true, "TMPDIR": true,
	}
	env := make([]string, 0, len(allowed)+8)
	for _, entry := range os.Environ() {
		key, _, ok := strings.Cut(entry, "=")
		if ok && allowed[key] {
			env = append(env, entry)
		}
	}
	env = append(env, "XDG_CONFIG_HOME="+configHome)
	if config.APIURL != "" {
		env = append(env, "PLANE_API_URL="+config.APIURL)
	}
	if config.AuthMode != "" {
		env = append(env, "PLANE_AUTH_MODE="+config.AuthMode)
	}
	if config.APIKey != "" {
		env = append(env, "PLANE_API_KEY="+config.APIKey)
	}
	if config.BearerToken != "" {
		env = append(env, "PLANE_BEARER_TOKEN="+config.BearerToken)
	}
	if config.Timeout > 0 {
		env = append(env, "PLANE_TIMEOUT="+config.Timeout.String())
	}
	return env
}

func (r *Runner) Run(t *testing.T, args ...string) commandResult {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()
	command := exec.CommandContext(ctx, r.binary, args...)
	command.Env = append([]string(nil), r.env...)
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	err := command.Run()
	if ctx.Err() != nil && err == nil {
		err = ctx.Err()
	}
	result := commandResult{
		Args:     redactArgs(append([]string(nil), args...), r.secrets),
		Stdout:   redactText(stdout.String(), r.secrets),
		Stderr:   redactText(stderr.String(), r.secrets),
		ExitCode: exitCode(command, err),
		Err:      err,
	}
	return result
}

func (r *Runner) RunJSON(t *testing.T, args ...string) (json.RawMessage, commandResult) {
	t.Helper()
	args = append(append([]string(nil), args...), "--output", "json")
	result := r.Run(t, args...)
	result.RequireSuccess(t)
	data := bytes.TrimSpace([]byte(result.Stdout))
	if len(data) == 0 {
		t.Fatalf("command %s returned empty JSON output", formatArgs(result.Args))
	}
	if !json.Valid(data) {
		t.Fatalf("command %s returned invalid JSON:\n%s", formatArgs(result.Args), result.Output())
	}
	assertNoCredentialMaterial(t, result.Output())
	return append(json.RawMessage(nil), data...), result
}

func (r commandResult) RequireSuccess(t *testing.T) {
	t.Helper()
	if r.Err != nil || r.ExitCode != 0 {
		t.Fatalf("command %s failed (exit %d): %v\n%s", formatArgs(r.Args), r.ExitCode, r.Err, r.Output())
	}
}

func (r commandResult) RequireQuietSuccess(t *testing.T) {
	t.Helper()
	r.RequireSuccess(t)
	if strings.TrimSpace(r.Stdout) != "" {
		t.Fatalf("command %s returned output for a quiet operation:\n%s", formatArgs(r.Args), r.Output())
	}
}

func (r commandResult) Output() string {
	parts := make([]string, 0, 2)
	if strings.TrimSpace(r.Stdout) != "" {
		parts = append(parts, strings.TrimSpace(r.Stdout))
	}
	if strings.TrimSpace(r.Stderr) != "" {
		parts = append(parts, strings.TrimSpace(r.Stderr))
	}
	return strings.Join(parts, "\n")
}

func exitCode(command *exec.Cmd, err error) int {
	if command.ProcessState != nil {
		return command.ProcessState.ExitCode()
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return -1
	}
	return -1
}

func formatArgs(args []string) string {
	quoted := make([]string, len(args))
	for index, arg := range args {
		quoted[index] = strconv.Quote(arg)
	}
	return strings.Join(quoted, " ")
}

func redactArgs(args, secrets []string) []string {
	redacted := make([]string, len(args))
	for index, arg := range args {
		redacted[index] = redactText(arg, secrets)
	}
	return redacted
}

func redactText(value string, secrets []string) string {
	ordered := append([]string(nil), secrets...)
	sort.Slice(ordered, func(i, j int) bool { return len(ordered[i]) > len(ordered[j]) })
	for _, secret := range ordered {
		if secret != "" {
			value = strings.ReplaceAll(value, secret, "[REDACTED]")
		}
	}
	return value
}

func assertNoCredentialMaterial(t *testing.T, value string) {
	t.Helper()
	lower := strings.ToLower(value)
	for _, marker := range []string{"api-key", "api_key", "bearer token", "authorization:", "presigned", "access_key_id", "secret_access_key", "security_token"} {
		if strings.Contains(lower, marker) {
			t.Fatalf("output contains credential material marker %q:\n%s", marker, value)
		}
	}
}

func jsonObject(data json.RawMessage) map[string]json.RawMessage {
	var object map[string]json.RawMessage
	if err := json.Unmarshal(data, &object); err != nil {
		return nil
	}
	return object
}

func jsonStringField(t *testing.T, data json.RawMessage, field string) string {
	t.Helper()
	object := jsonObject(data)
	if object == nil {
		t.Fatalf("expected JSON object while reading %q: %s", field, data)
	}
	value, ok := object[field]
	if !ok || string(value) == "null" {
		t.Fatalf("JSON object omitted non-null %q: %s", field, data)
	}
	var stringValue string
	if json.Unmarshal(value, &stringValue) == nil && stringValue != "" {
		return stringValue
	}
	var number json.Number
	decoder := json.NewDecoder(bytes.NewReader(value))
	decoder.UseNumber()
	if err := decoder.Decode(&number); err == nil && number.String() != "" {
		return number.String()
	}
	t.Fatalf("JSON field %q is not a string or number: %s", field, value)
	return ""
}

func projectID(t *testing.T, data json.RawMessage) string { return jsonStringField(t, data, "id") }

func projectIdentifier(t *testing.T, data json.RawMessage) string {
	return jsonStringField(t, data, "identifier")
}

func workItemID(t *testing.T, data json.RawMessage) string { return jsonStringField(t, data, "id") }

func workItemSequence(t *testing.T, data json.RawMessage) string {
	return jsonStringField(t, data, "sequence_id")
}

var uniqueNameSequence uint64

func uniqueName(prefix string) string {
	normalizedPrefix := normalizeNamePrefix(prefix)
	if normalizedPrefix == "" {
		normalizedPrefix = "generated"
	}
	sequence := atomic.AddUint64(&uniqueNameSequence, 1)
	return fmt.Sprintf("planeshiftfunctional%s%d%d%d", normalizedPrefix, time.Now().UnixNano(), os.Getpid(), sequence)
}

// Plane project names must contain only alphanumeric characters. Keep the
// helper usable for work-item names too so every generated test name follows
// the same safe convention.
func normalizeNamePrefix(prefix string) string {
	var normalized strings.Builder
	normalized.Grow(len(prefix))
	for _, character := range prefix {
		switch {
		case character >= 'a' && character <= 'z':
			normalized.WriteRune(character)
		case character >= 'A' && character <= 'Z':
			normalized.WriteRune(character)
		case character >= '0' && character <= '9':
			normalized.WriteRune(character)
		}
	}
	return normalized.String()
}

func TestUniqueNameProducesSafeDistinctValues(t *testing.T) {
	tests := []struct {
		name       string
		prefix     string
		normalized string
	}{
		{name: "project", prefix: "project", normalized: "project"},
		{name: "template project", prefix: "template-project", normalized: "templateproject"},
		{name: "work items project", prefix: "work-items-project", normalized: "workitemsproject"},
		{name: "empty after normalization", prefix: "---", normalized: "generated"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := uniqueName(test.prefix)
			if got == "" {
				t.Fatal("uniqueName returned an empty value")
			}
			if !strings.HasPrefix(got, "planeshiftfunctional"+test.normalized) {
				t.Fatalf("uniqueName(%q) = %q, want normalized prefix %q", test.prefix, got, test.normalized)
			}
			for _, character := range got {
				if !((character >= 'a' && character <= 'z') ||
					(character >= 'A' && character <= 'Z') ||
					(character >= '0' && character <= '9')) {
					t.Fatalf("uniqueName(%q) contains non-alphanumeric character %q: %q", test.prefix, character, got)
				}
			}
		})
	}

	first := uniqueName("project")
	second := uniqueName("project")
	if first == second {
		t.Fatalf("consecutive uniqueName calls returned the same value %q", first)
	}
}
