package cmd

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"planeshift/config"

	"github.com/spf13/viper"
)

func TestResolvePlaneSettingsEnvironmentOverridesConfig(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(configPath, []byte("plane:\n  api_url: https://config.example\n  auth_mode: api-key\n  api_key: config-key\n  timeout: 5s\n"), 0600); err != nil {
		t.Fatal(err)
	}
	v := viper.New()
	v.SetConfigFile(configPath)
	if err := v.ReadInConfig(); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PLANE_API_URL", "https://env.example/api/v1")
	t.Setenv("PLANE_AUTH_MODE", "bearer")
	t.Setenv("PLANE_BEARER_TOKEN", "env-token")
	t.Setenv("PLANE_TIMEOUT", "2s")
	if err := bindPlaneEnvironment(v); err != nil {
		t.Fatal(err)
	}
	settings, err := resolvePlaneSettings(v)
	if err != nil {
		t.Fatal(err)
	}
	if settings.APIURL != "https://env.example/api/v1" || settings.AuthMode != config.AuthModeBearer ||
		settings.BearerToken != "env-token" || settings.Timeout != 2*time.Second {
		t.Fatalf("resolved settings = %#v", settings)
	}
	if settings.APIKey != "config-key" {
		t.Fatalf("config API key was not retained in the model: %#v", settings)
	}
	if err := settings.Validate(); err == nil {
		t.Fatal("ambiguous config/env credentials were not rejected")
	}
}

func TestResolvePlaneSettingsUsesConfigWhenEnvironmentAbsent(t *testing.T) {
	v := viper.New()
	v.Set("plane.api_url", "https://config.example")
	v.Set("plane.auth_mode", "api-key")
	v.Set("plane.api_key", "config-key")
	v.Set("plane.timeout", "5s")
	for _, key := range []string{"PLANE_API_URL", "PLANE_AUTH_MODE", "PLANE_API_KEY", "PLANE_BEARER_TOKEN", "PLANE_TIMEOUT"} {
		t.Setenv(key, "")
	}
	if err := bindPlaneEnvironment(v); err != nil {
		t.Fatal(err)
	}
	settings, err := resolvePlaneSettings(v)
	if err != nil {
		t.Fatal(err)
	}
	if settings.APIURL != "https://config.example" || settings.AuthMode != config.AuthModeAPIKey ||
		settings.APIKey != "config-key" || settings.Timeout != 5*time.Second {
		t.Fatalf("resolved settings = %#v", settings)
	}
}

func TestCreateRestrictedConfigFile(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.yaml")
	if !createRestrictedConfigFile(configPath) {
		t.Fatal("createRestrictedConfigFile reported no creation")
	}
	info, err := os.Stat(configPath)
	if err != nil {
		t.Fatal(err)
	}
	if mode := info.Mode().Perm(); mode != 0600 {
		t.Fatalf("config mode = %#o, want 0600", mode)
	}
	if createRestrictedConfigFile(configPath) {
		t.Fatal("createRestrictedConfigFile reported creation for existing file")
	}
}

func TestRootCommandHierarchyAndLazyFactory(t *testing.T) {
	conf := &config.Config{}
	factory := newPlaneClientFactory(conf)
	if factory == nil {
		t.Fatal("newPlaneClientFactory returned nil")
	}
	project, _, err := RootCmd.Find([]string{config.ProjectCommandName})
	if err != nil || project == nil {
		t.Fatalf("project is not registered below RootCmd: command=%v err=%v", project, err)
	}
	if project.Parent() != RootCmd {
		t.Fatalf("project parent = %v, want RootCmd", project.Parent())
	}
	if project.Name() != config.ProjectCommandName || !project.HasAlias(config.ProjectCommandAlias) {
		t.Fatalf("project command = name %q aliases %v", project.Name(), project.Aliases)
	}
	workItem, _, err := RootCmd.Find([]string{config.WorkItemCommandName})
	if err != nil || workItem == nil || workItem.Parent() != RootCmd {
		t.Fatalf("work-item is not registered below RootCmd: command=%v err=%v", workItem, err)
	}
	if workItem.Name() != config.WorkItemCommandName || !workItem.HasAlias(config.WorkItemCommandAlias) {
		t.Fatalf("work-item command = name %q aliases %v", workItem.Name(), workItem.Aliases)
	}
	state, _, err := RootCmd.Find([]string{config.StateCommandName})
	if err != nil || state == nil || state.Parent() != RootCmd {
		t.Fatalf("state is not registered below RootCmd: command=%v err=%v", state, err)
	}
	if state.Name() != config.StateCommandName || !state.HasAlias(config.StateCommandAlias) {
		t.Fatalf("state command = name %q aliases %v", state.Name(), state.Aliases)
	}
	pluralState, _, err := RootCmd.Find([]string{config.StateCommandAlias})
	if err != nil || pluralState != state {
		t.Fatalf("plural state alias resolved to command=%v err=%v", pluralState, err)
	}
	pluralWorkItem, _, err := RootCmd.Find([]string{config.WorkItemCommandAlias})
	if err != nil || pluralWorkItem != workItem {
		t.Fatalf("plural work-item alias resolved to command=%v err=%v", pluralWorkItem, err)
	}
	plural, _, err := RootCmd.Find([]string{config.ProjectCommandAlias})
	if err != nil || plural != project {
		t.Fatalf("plural project alias resolved to command=%v err=%v, want %v", plural, err, project)
	}
	for _, name := range []string{"project-label", "project-labels"} {
		removed, _, err := RootCmd.Find([]string{name})
		if err == nil && removed != nil && removed.Parent() == RootCmd {
			t.Fatalf("removed command %q is still registered below RootCmd", name)
		}
	}
	version, _, err := RootCmd.Find([]string{"version"})
	if err != nil || version == nil {
		t.Fatalf("root version is unavailable: command=%v err=%v", version, err)
	}
	if obsolete, _, err := RootCmd.Find([]string{"get"}); err == nil {
		t.Fatalf("obsolete get hierarchy is still registered: command=%v", obsolete)
	}
	// The factory is only invoked by an operation; constructing the command
	// hierarchy above must not require Plane credentials.
}

func TestRootVersionDefaultsToJSON(t *testing.T) {
	previous := *c
	t.Cleanup(func() { *c = previous })
	c.VersionJSON = `{"SemVer":"v0.0.999","BuildDate":"","GitCommit":"","GitRef":""}`
	c.OutputFormat = ""
	c.FormatOverridden = false
	version, _, err := RootCmd.Find([]string{"version"})
	if err != nil {
		t.Fatal(err)
	}
	output := captureRootStdout(t, func() { version.Run(version, nil) })
	if c.OutputFormat != "json" {
		t.Fatalf("version output format = %q, want json", c.OutputFormat)
	}
	if !strings.Contains(output, `"SemVer"`) {
		t.Fatalf("version output was not JSON: %q", output)
	}
}

func captureRootStdout(t *testing.T, function func()) string {
	t.Helper()
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	previous := os.Stdout
	os.Stdout = writer
	function()
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	os.Stdout = previous
	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	if err := reader.Close(); err != nil {
		t.Fatal(err)
	}
	return string(data)
}
