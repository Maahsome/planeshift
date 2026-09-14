// Package projectfeatures contains the typed contracts and client for the
// documented Plane Project Features operations.
package projectfeatures

import (
	"bytes"
	"encoding/json"
	"fmt"
)

var featureFields = []string{
	"epics",
	"modules",
	"cycles",
	"views",
	"pages",
	"intakes",
	"work_item_types",
}

// ProjectFeatures is the feature object returned by Plane. The documented
// fields are strict booleans; Unknown retains future or undocumented fields
// without changing their JSON types or null values.
type ProjectFeatures struct {
	Epics         bool                       `json:"epics"`
	Modules       bool                       `json:"modules"`
	Cycles        bool                       `json:"cycles"`
	Views         bool                       `json:"views"`
	Pages         bool                       `json:"pages"`
	Intakes       bool                       `json:"intakes"`
	WorkItemTypes bool                       `json:"work_item_types"`
	Unknown       map[string]json.RawMessage `json:"-"`
	present       map[string]bool
}

// UnmarshalJSON decodes the documented boolean fields strictly and retains
// every unknown field as its original JSON value. In particular, null,
// numeric, and string values are rejected for documented feature fields.
func (p *ProjectFeatures) UnmarshalJSON(data []byte) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	for _, name := range featureFields {
		value, ok := fields[name]
		if !ok {
			continue
		}
		trimmed := bytes.TrimSpace(value)
		if !bytes.Equal(trimmed, []byte("true")) && !bytes.Equal(trimmed, []byte("false")) {
			return fmt.Errorf("project feature %q must be a boolean", name)
		}
	}

	type featureAlias ProjectFeatures
	var decoded featureAlias
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	unknown := make(map[string]json.RawMessage)
	present := make(map[string]bool)
	for name, value := range fields {
		if isFeatureField(name) {
			present[name] = true
			continue
		}
		unknown[name] = append(json.RawMessage(nil), value...)
	}
	decoded.Unknown = unknown
	decoded.present = present
	*p = ProjectFeatures(decoded)
	return nil
}

// MarshalJSON emits all seven typed fields for values constructed by callers,
// while decoded values retain the original known-field presence. Unknown
// fields are merged without allowing them to overwrite typed fields.
func (p ProjectFeatures) MarshalJSON() ([]byte, error) {
	type featureAlias ProjectFeatures
	data, err := json.Marshal(featureAlias(p))
	if err != nil {
		return nil, err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return nil, err
	}
	if p.present != nil {
		presentFields := make(map[string]json.RawMessage, len(fields))
		for name := range p.present {
			presentFields[name] = fields[name]
		}
		fields = presentFields
	}
	for name, value := range p.Unknown {
		if !isFeatureField(name) {
			fields[name] = append(json.RawMessage(nil), value...)
		}
	}
	return json.Marshal(fields)
}

func isFeatureField(name string) bool {
	for _, field := range featureFields {
		if field == name {
			return true
		}
	}
	return false
}

// UpdateProjectFeaturesRequest contains the optional PATCH fields. A nil
// pointer omits a field; a pointer to false remains present on the wire.
type UpdateProjectFeaturesRequest struct {
	Epics         *bool `json:"epics,omitempty"`
	Modules       *bool `json:"modules,omitempty"`
	Cycles        *bool `json:"cycles,omitempty"`
	Views         *bool `json:"views,omitempty"`
	Pages         *bool `json:"pages,omitempty"`
	Intakes       *bool `json:"intakes,omitempty"`
	WorkItemTypes *bool `json:"work_item_types,omitempty"`
}
