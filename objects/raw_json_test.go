package objects

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestRawJSONPreservesUnknownAndNullableValues(t *testing.T) {
	raw, err := NewRawJSON([]byte(`{"id":"1","nullable":null,"unknown":{"enabled":true}}`))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(raw.ToJSON(), `"nullable": null`) || !strings.Contains(raw.ToJSON(), `"unknown"`) {
		t.Fatalf("JSON output lost dynamic fields: %s", raw.ToJSON())
	}
	encoded, err := json.Marshal(raw)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(encoded), `"unknown"`) {
		t.Fatalf("MarshalJSON lost unknown field: %s", encoded)
	}
	if !strings.Contains(raw.ToYAML(), "nullable") || !strings.Contains(raw.ToTEXT(false), "unknown") {
		t.Fatalf("dynamic output lost fields")
	}
	if raw.ToRAW() != `{"id":"1","nullable":null,"unknown":{"enabled":true}}` {
		t.Fatalf("raw output did not preserve source document: %s", raw.ToRAW())
	}
}

func TestRawJSONRejectsMalformedDataWithoutPanicking(t *testing.T) {
	if _, err := NewRawJSON([]byte(`{"broken":`)); err == nil {
		t.Fatal("NewRawJSON accepted malformed data")
	}
	var zero RawJSON
	if zero.ToJSON() != "" || zero.ToYAML() != "" || zero.ToGRON() != "" || zero.ToTEXT(false) != "" {
		t.Fatal("zero RawJSON did not return safe empty output")
	}
}
