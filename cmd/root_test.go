package cmd

import (
	"os"
	"path/filepath"
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

func TestPlaneFactoryIsLazyAndVersionHierarchyRemainsAvailable(t *testing.T) {
	conf := &config.Config{}
	factory := newPlaneClientFactory(conf)
	if factory == nil {
		t.Fatal("newPlaneClientFactory returned nil")
	}
	// Calling the factory would correctly reject missing resource credentials;
	// the version commands never call it.
	if _, _, err := RootCmd.Find([]string{"get", "version"}); err != nil {
		t.Fatalf("get version is not registered below RootCmd: %v", err)
	}
}
