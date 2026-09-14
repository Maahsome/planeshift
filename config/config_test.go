package config

import (
	"strings"
	"testing"
	"time"

	"planeshift/objects"
)

func TestPlaneSettingsValidation(t *testing.T) {
	tests := []struct {
		name     string
		settings PlaneSettings
		wantErr  bool
	}{
		{name: "api key", settings: PlaneSettings{AuthMode: AuthModeAPIKey, APIKey: "key"}},
		{name: "bearer", settings: PlaneSettings{AuthMode: AuthModeBearer, BearerToken: "token"}},
		{name: "missing mode", settings: PlaneSettings{APIKey: "key"}, wantErr: true},
		{name: "missing api key", settings: PlaneSettings{AuthMode: AuthModeAPIKey}, wantErr: true},
		{name: "missing bearer", settings: PlaneSettings{AuthMode: AuthModeBearer}, wantErr: true},
		{name: "ambiguous credentials", settings: PlaneSettings{AuthMode: AuthModeAPIKey, APIKey: "key", BearerToken: "token"}, wantErr: true},
		{name: "negative timeout", settings: PlaneSettings{AuthMode: AuthModeAPIKey, APIKey: "key", Timeout: -time.Second}, wantErr: true},
		{name: "long timeout", settings: PlaneSettings{AuthMode: AuthModeAPIKey, APIKey: "key", Timeout: MaxPlaneTimeout + time.Nanosecond}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.settings.Validate()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPlaneSettingsDefaultsAndSecretRedaction(t *testing.T) {
	settings := (PlaneSettings{AuthMode: AuthModeAPIKey, APIKey: "super-secret"}).Normalize()
	if settings.APIURL != DefaultPlaneAPIURL {
		t.Fatalf("APIURL = %q, want %q", settings.APIURL, DefaultPlaneAPIURL)
	}
	if settings.Timeout != DefaultPlaneTimeout {
		t.Fatalf("Timeout = %s, want %s", settings.Timeout, DefaultPlaneTimeout)
	}
	text := settings.String()
	if strings.Contains(text, "super-secret") {
		t.Fatalf("settings String disclosed credential: %q", text)
	}
	if _, err := ParsePlaneTimeout("not-a-duration"); err == nil {
		t.Fatal("ParsePlaneTimeout accepted invalid duration")
	}
	if got, err := ParsePlaneTimeout("2s"); err != nil || got != 2*time.Second {
		t.Fatalf("ParsePlaneTimeout(2s) = %s, %v", got, err)
	}
}

func TestOutputDataDispatchesLowercaseFormats(t *testing.T) {
	data, err := objects.NewRawJSON([]byte(`{"nullable":null,"unknown":{"value":true}}`))
	if err != nil {
		t.Fatal(err)
	}

	for _, format := range []string{"JSON", "YAML", "GRON", "TEXT", "TABLE", "RAW"} {
		t.Run(format, func(t *testing.T) {
			output := (&Config{OutputFormat: format}).outputData(data)
			if output == "" {
				t.Fatalf("outputData(%q) returned empty output", format)
			}
			if format == "RAW" && output != `{"nullable":null,"unknown":{"value":true}}` {
				t.Fatalf("raw output = %q", output)
			}
		})
	}
}

func TestProjectOutputImplementsDispatcherContract(t *testing.T) {
	var output Outputtable = objects.Project{}
	if output == nil {
		t.Fatal("objects.Project did not satisfy Outputtable")
	}
}
