// Package states contains the route-specific contracts for Plane Work Item
// States. Transport, authentication, pagination validation, and errors remain
// owned by the shared plane package.
package states

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strings"

	"planeshift/plane"
)

// State is a Work Item State response. Pointer fields preserve nullable
// scalar values, while Sequence preserves the documented number/string/null
// discrepancy without coercion. Unknown response fields are retained.
type State struct {
	ID             *string                    `json:"id,omitempty"`
	CreatedAt      *string                    `json:"created_at,omitempty"`
	UpdatedAt      *string                    `json:"updated_at,omitempty"`
	Name           *string                    `json:"name,omitempty"`
	Description    *string                    `json:"description,omitempty"`
	Color          *string                    `json:"color,omitempty"`
	WorkspaceSlug  *string                    `json:"workspace_slug,omitempty"`
	Sequence       json.RawMessage            `json:"sequence,omitempty"`
	Group          *string                    `json:"group,omitempty"`
	Default        *bool                      `json:"default,omitempty"`
	IsTriage       *bool                      `json:"is_triage,omitempty"`
	ExternalSource *string                    `json:"external_source,omitempty"`
	ExternalID     *string                    `json:"external_id,omitempty"`
	CreatedBy      *string                    `json:"created_by,omitempty"`
	UpdatedBy      *string                    `json:"updated_by,omitempty"`
	Project        *string                    `json:"project,omitempty"`
	Workspace      *string                    `json:"workspace,omitempty"`
	Unknown        map[string]json.RawMessage `json:"-"`
	present        map[string]bool            `json:"-"`
}

var stateKnownFields = []string{
	"id", "created_at", "updated_at", "name", "description", "color",
	"workspace_slug", "sequence", "group", "default", "is_triage",
	"external_source", "external_id", "created_by", "updated_by", "project",
	"workspace",
}

// UnmarshalJSON validates documented scalar fields and keeps unknown keys.
func (s *State) UnmarshalJSON(data []byte) error {
	type stateAlias State
	var decoded stateAlias
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	var allFields map[string]json.RawMessage
	if err := json.Unmarshal(data, &allFields); err != nil {
		return err
	}
	if sequence, ok := allFields["sequence"]; ok {
		if err := validateSequence(sequence); err != nil {
			return err
		}
	}
	unknown := make(map[string]json.RawMessage)
	present := make(map[string]bool, len(allFields))
	for key, value := range allFields {
		if contains(stateKnownFields, key) {
			present[key] = true
			continue
		}
		unknown[key] = append(json.RawMessage(nil), value...)
	}
	decoded.Unknown = unknown
	decoded.present = present
	*s = State(decoded)
	return nil
}

// MarshalJSON emits only response fields that were present when decoded and
// merges unknown fields without allowing them to override typed fields.
func (s State) MarshalJSON() ([]byte, error) {
	type stateAlias State
	data, err := json.Marshal(stateAlias(s))
	if err != nil {
		return nil, err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return nil, err
	}
	if s.present != nil {
		presentFields := make(map[string]json.RawMessage, len(s.present))
		for key := range s.present {
			if value, ok := fields[key]; ok {
				presentFields[key] = value
			} else {
				presentFields[key] = json.RawMessage("null")
			}
		}
		fields = presentFields
	}
	for key, value := range s.Unknown {
		if _, exists := fields[key]; !exists {
			fields[key] = append(json.RawMessage(nil), value...)
		}
	}
	return json.Marshal(fields)
}

// StatePage is the cursor page returned by the state collection route.
type StatePage struct {
	plane.CursorPage[State]
	GroupedBy    *string `json:"grouped_by,omitempty"`
	SubGroupedBy *string `json:"sub_grouped_by,omitempty"`
	TotalCount   *int    `json:"total_count,omitempty"`
}

var statePageKnownFields = []string{
	"next_cursor", "prev_cursor", "next_page_results", "prev_page_results",
	"count", "total_pages", "total_results", "extra_stats", "results",
	"grouped_by", "sub_grouped_by", "total_count",
}

// UnmarshalJSON retains page grouping metadata and unknown page keys.
func (p *StatePage) UnmarshalJSON(data []byte) error {
	var cursor plane.CursorPage[State]
	if err := json.Unmarshal(data, &cursor); err != nil {
		return err
	}
	var grouped struct {
		GroupedBy    *string `json:"grouped_by"`
		SubGroupedBy *string `json:"sub_grouped_by"`
		TotalCount   *int    `json:"total_count"`
	}
	if err := json.Unmarshal(data, &grouped); err != nil {
		return err
	}
	var unknown map[string]json.RawMessage
	if err := json.Unmarshal(data, &unknown); err != nil {
		return err
	}
	for _, key := range statePageKnownFields {
		delete(unknown, key)
	}
	cursor.Unknown = unknown
	*p = StatePage{CursorPage: cursor, GroupedBy: grouped.GroupedBy, SubGroupedBy: grouped.SubGroupedBy, TotalCount: grouped.TotalCount}
	return nil
}

// MarshalJSON emits cursor, grouping, and unknown page data.
func (p StatePage) MarshalJSON() ([]byte, error) {
	type cursorAlias plane.CursorPage[State]
	data, err := json.Marshal(cursorAlias(p.CursorPage))
	if err != nil {
		return nil, err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return nil, err
	}
	grouped, err := json.Marshal(struct {
		GroupedBy    *string `json:"grouped_by"`
		SubGroupedBy *string `json:"sub_grouped_by"`
		TotalCount   *int    `json:"total_count"`
	}{p.GroupedBy, p.SubGroupedBy, p.TotalCount})
	if err != nil {
		return nil, err
	}
	var groupedFields map[string]json.RawMessage
	if err := json.Unmarshal(grouped, &groupedFields); err != nil {
		return nil, err
	}
	for key, value := range groupedFields {
		fields[key] = value
	}
	for key, value := range p.CursorPage.Unknown {
		if _, exists := fields[key]; !exists {
			fields[key] = append(json.RawMessage(nil), value...)
		}
	}
	return json.Marshal(fields)
}

// CreateStateRequest is the documented create body. Name and Color are the
// only required fields; pointer optionals retain explicit zero, false, empty,
// and null values.
type CreateStateRequest struct {
	Name           string           `json:"name"`
	Description    *string          `json:"description,omitempty"`
	Color          string           `json:"color"`
	Sequence       *json.RawMessage `json:"sequence,omitempty"`
	Group          *string          `json:"group,omitempty"`
	IsTriage       *bool            `json:"is_triage,omitempty"`
	Default        *bool            `json:"default,omitempty"`
	ExternalSource *string          `json:"external_source,omitempty"`
	ExternalID     *string          `json:"external_id,omitempty"`
}

func (r CreateStateRequest) validate() error {
	if strings.TrimSpace(r.Name) == "" {
		return fmt.Errorf("state name is required")
	}
	if strings.TrimSpace(r.Color) == "" {
		return fmt.Errorf("state color is required")
	}
	if r.Sequence != nil {
		return validateSequence(*r.Sequence)
	}
	return nil
}

// UpdateStateRequest contains all documented optional PATCH fields. A nil
// pointer omits a property; a non-nil pointer is serialized even when its
// value is empty, false, zero, or JSON null.
type UpdateStateRequest struct {
	Name           *string          `json:"name,omitempty"`
	Description    *string          `json:"description,omitempty"`
	Color          *string          `json:"color,omitempty"`
	Sequence       *json.RawMessage `json:"sequence,omitempty"`
	Group          *string          `json:"group,omitempty"`
	IsTriage       *bool            `json:"is_triage,omitempty"`
	Default        *bool            `json:"default,omitempty"`
	ExternalSource *string          `json:"external_source,omitempty"`
	ExternalID     *string          `json:"external_id,omitempty"`
}

func (r UpdateStateRequest) validate() error {
	if r.Sequence != nil {
		return validateSequence(*r.Sequence)
	}
	return nil
}

// ListOptions contains only the documented collection query controls.
type ListOptions struct {
	Cursor  string
	PerPage int
	Fields  string
	Expand  string
}

// Query delegates pagination validation and query copying to the shared
// client foundation.
func (o ListOptions) Query() (url.Values, error) {
	return plane.WithPagination(nil, plane.PaginationOptions{
		Cursor: o.Cursor, PerPage: o.PerPage, Fields: o.Fields, Expand: o.Expand,
	})
}

func validateSequence(value json.RawMessage) error {
	trimmed := bytes.TrimSpace(value)
	if bytes.Equal(trimmed, []byte("null")) {
		return nil
	}
	decoder := json.NewDecoder(bytes.NewReader(trimmed))
	decoder.UseNumber()
	var decoded any
	if err := decoder.Decode(&decoded); err != nil {
		return fmt.Errorf("state sequence must be a number, string, or null")
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return fmt.Errorf("state sequence must be a number, string, or null")
	}
	switch decoded.(type) {
	case json.Number, string:
		return nil
	default:
		return fmt.Errorf("state sequence must be a number, string, or null")
	}
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
