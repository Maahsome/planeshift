package projectfeatures

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestProjectFeaturesRoundTripPreservesTypedAndUnknownFields(t *testing.T) {
	sample := []byte(`{"epics":true,"modules":false,"cycles":true,"views":false,"pages":true,"intakes":false,"work_item_types":true,"future":null,"dynamic":{"enabled":true}}`)
	var features ProjectFeatures
	if err := json.Unmarshal(sample, &features); err != nil {
		t.Fatal(err)
	}
	if !features.Epics || features.Modules || !features.Cycles || features.Views || !features.Pages || features.Intakes || !features.WorkItemTypes {
		t.Fatalf("decoded features = %#v", features)
	}
	if string(features.Unknown["future"]) != "null" || string(features.Unknown["dynamic"]) != `{"enabled":true}` {
		t.Fatalf("unknown fields = %#v", features.Unknown)
	}

	encoded, err := json.Marshal(features)
	if err != nil {
		t.Fatal(err)
	}
	var roundTrip map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &roundTrip); err != nil {
		t.Fatal(err)
	}
	for _, name := range featureFields {
		if _, ok := roundTrip[name]; !ok {
			t.Fatalf("round trip omitted %q: %s", name, encoded)
		}
	}
	if string(roundTrip["future"]) != "null" || string(roundTrip["dynamic"]) != `{"enabled":true}` {
		t.Fatalf("round trip changed unknown fields: %s", encoded)
	}
}

func TestProjectFeaturesRejectsNonBooleanDocumentedFields(t *testing.T) {
	for _, value := range []string{`null`, `1`, `"true"`, `{}`} {
		var features ProjectFeatures
		if err := json.Unmarshal([]byte(`{"epics":`+value+`}`), &features); err == nil {
			t.Fatalf("accepted epics=%s", value)
		}
	}
}

func TestProjectFeaturesRetainsUnknownJSONTypes(t *testing.T) {
	var features ProjectFeatures
	if err := json.Unmarshal([]byte(`{"future_number":12,"future_string":"value","future_null":null}`), &features); err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(features)
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{`"future_number":12`, `"future_string":"value"`, `"future_null":null`} {
		if !strings.Contains(string(encoded), expected) {
			t.Fatalf("encoded output omitted %s: %s", expected, encoded)
		}
	}
}

func TestUpdateProjectFeaturesRequestPresenceAndWireNames(t *testing.T) {
	falseValue := false
	trueValue := true
	request := UpdateProjectFeaturesRequest{Epics: &falseValue, Modules: &trueValue}
	data, err := json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}
	encoded := string(data)
	if !strings.Contains(encoded, `"epics":false`) || !strings.Contains(encoded, `"modules":true`) {
		t.Fatalf("explicit values omitted: %s", data)
	}
	for _, name := range []string{"cycles", "views", "pages", "intakes", "work_item_types"} {
		if strings.Contains(encoded, `"`+name+`"`) {
			t.Fatalf("unset field %q was serialized: %s", name, data)
		}
	}
	minimal, err := json.Marshal(UpdateProjectFeaturesRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if string(minimal) != "{}" {
		t.Fatalf("minimal request = %s", minimal)
	}
}
