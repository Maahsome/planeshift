package objects

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/maahsome/gron"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

// RawJSON is a dynamic output object for resource responses. It keeps the
// original JSON document so unknown fields, nullable values, and numeric
// representations survive until a selected output format is rendered.
type RawJSON struct {
	data json.RawMessage
}

// NewRawJSON validates and copies a JSON document into a RawJSON output.
func NewRawJSON(data []byte) (RawJSON, error) {
	if !json.Valid(data) {
		return RawJSON{}, fmt.Errorf("invalid JSON document")
	}
	return RawJSON{data: append(json.RawMessage(nil), data...)}, nil
}

// NewRawJSONFromValue marshals a Go value into a dynamic output object.
func NewRawJSONFromValue(value any) (RawJSON, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return RawJSON{}, fmt.Errorf("encode JSON document")
	}
	return NewRawJSON(data)
}

// Raw returns a defensive copy of the original document.
func (r RawJSON) Raw() json.RawMessage {
	return append(json.RawMessage(nil), r.data...)
}

// UnmarshalJSON validates and stores the document without discarding unknown
// fields.
func (r *RawJSON) UnmarshalJSON(data []byte) error {
	if !json.Valid(data) {
		return fmt.Errorf("invalid JSON document")
	}
	r.data = append(r.data[:0], data...)
	return nil
}

// MarshalJSON returns the preserved JSON document.
func (r RawJSON) MarshalJSON() ([]byte, error) {
	if !json.Valid(r.data) {
		return nil, fmt.Errorf("invalid JSON document")
	}
	return r.Raw(), nil
}

func (r RawJSON) prettyJSON() string {
	if !json.Valid(r.data) {
		logrus.Error("Error extracting JSON")
		return ""
	}
	var output bytes.Buffer
	if err := json.Indent(&output, r.data, "", "  "); err != nil {
		logrus.WithError(err).Error("Error formatting JSON")
		return ""
	}
	return output.String()
}

func (r RawJSON) value() (any, error) {
	if !json.Valid(r.data) {
		return nil, fmt.Errorf("invalid JSON document")
	}
	decoder := json.NewDecoder(bytes.NewReader(r.data))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		return nil, err
	}
	return value, nil
}

// ToJSON implements the existing output object shape.
func (r RawJSON) ToJSON() string {
	return r.prettyJSON()
}

// ToYAML converts the dynamic JSON value while retaining nulls and unknown
// object keys.
func (r RawJSON) ToYAML() string {
	value, err := r.value()
	if err != nil {
		logrus.WithError(err).Error("Error extracting YAML")
		return ""
	}
	data, err := yaml.Marshal(value)
	if err != nil {
		logrus.WithError(err).Error("Error extracting YAML")
		return ""
	}
	return string(data)
}

// ToGRON converts the preserved JSON through the repository's existing GRON
// dependency.
func (r RawJSON) ToGRON() string {
	pretty := r.prettyJSON()
	if pretty == "" {
		return ""
	}
	reader := strings.NewReader(pretty)
	output := &bytes.Buffer{}
	converter := gron.NewGron(reader, output)
	converter.SetMonochrome(false)
	if err := converter.ToGron(); err != nil {
		logrus.WithError(err).Error("Problem generating GRON syntax")
		return ""
	}
	return output.String()
}

// ToRAW supports config's optional raw-output hook.
func (r RawJSON) ToRAW() string {
	return string(r.data)
}

// ToTEXT is a deterministic JSON fallback for dynamic resources. It ignores
// noHeaders because a raw document has no table header to suppress.
func (r RawJSON) ToTEXT(noHeaders bool) string {
	return r.prettyJSON()
}
