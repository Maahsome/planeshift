// Package projectlabels contains the typed contracts and client for Plane's
// workspace-scoped Project Labels resource.
package projectlabels

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"planeshift/plane"
)

var projectLabelKnownFields = []string{
	"id", "name", "description", "color", "sort_order", "workspace",
	"created_at", "updated_at", "created_by", "updated_by",
}

// ProjectLabel is the documented workspace-scoped Project Label response.
// Nullable fields use pointers so null and non-null values remain distinct.
// Unknown retains future response fields as their original JSON values.
type ProjectLabel struct {
	ID          string                     `json:"id"`
	Name        string                     `json:"name"`
	Description *string                    `json:"description"`
	Color       *string                    `json:"color"`
	SortOrder   *float64                   `json:"sort_order"`
	Workspace   *string                    `json:"workspace"`
	CreatedAt   *string                    `json:"created_at"`
	UpdatedAt   *string                    `json:"updated_at"`
	CreatedBy   *string                    `json:"created_by"`
	UpdatedBy   *string                    `json:"updated_by"`
	Unknown     map[string]json.RawMessage `json:"-"`
	present     map[string]bool            `json:"-"`
}

// UnmarshalJSON strictly decodes documented scalar fields and retains
// unmodeled response fields without changing their JSON representation.
func (p *ProjectLabel) UnmarshalJSON(data []byte) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	for _, name := range []string{"id", "name"} {
		if value, ok := fields[name]; ok {
			if bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
				return fmt.Errorf("project label %q must be a string", name)
			}
		}
	}

	type projectLabelAlias ProjectLabel
	var decoded projectLabelAlias
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}

	unknown := make(map[string]json.RawMessage)
	present := make(map[string]bool, len(fields))
	for name, value := range fields {
		if containsProjectLabelField(name) {
			present[name] = true
			continue
		}
		unknown[name] = append(json.RawMessage(nil), value...)
	}
	decoded.Unknown = unknown
	decoded.present = present
	*p = ProjectLabel(decoded)
	return nil
}

// MarshalJSON preserves the presence of documented fields, explicit nulls,
// and unknown fields retained by UnmarshalJSON.
func (p ProjectLabel) MarshalJSON() ([]byte, error) {
	type projectLabelAlias ProjectLabel
	data, err := json.Marshal(projectLabelAlias(p))
	if err != nil {
		return nil, err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return nil, err
	}
	if p.present != nil {
		presentFields := make(map[string]json.RawMessage, len(p.present))
		for name := range p.present {
			if value, ok := fields[name]; ok {
				presentFields[name] = value
			}
		}
		fields = presentFields
	}
	for name, value := range p.Unknown {
		if _, exists := fields[name]; !exists {
			fields[name] = append(json.RawMessage(nil), value...)
		}
	}
	return json.Marshal(fields)
}

func containsProjectLabelField(target string) bool {
	for _, field := range projectLabelKnownFields {
		if field == target {
			return true
		}
	}
	return false
}

// ProjectLabelPage is the cursor page returned by the list operation. It
// retains the shared cursor metadata plus Project Labels grouping/count keys.
// Unknown top-level page fields are stored in the embedded Unknown map.
type ProjectLabelPage struct {
	plane.CursorPage[ProjectLabel]
	GroupedBy    *string `json:"grouped_by"`
	SubGroupedBy *string `json:"sub_grouped_by"`
	TotalCount   *int    `json:"total_count"`
	present      map[string]bool
}

var projectLabelPageKnownFields = []string{
	"next_cursor", "prev_cursor", "next_page_results", "prev_page_results",
	"count", "total_pages", "total_results", "extra_stats", "results",
	"grouped_by", "sub_grouped_by", "total_count",
}

// UnmarshalJSON decodes cursor/grouping metadata while remembering which
// fields were present so omitted values are not manufactured on round trips.
func (p *ProjectLabelPage) UnmarshalJSON(data []byte) error {
	var cursor plane.CursorPage[ProjectLabel]
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
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	unknown := make(map[string]json.RawMessage)
	present := make(map[string]bool, len(fields))
	for name, value := range fields {
		present[name] = true
		if !containsProjectLabelPageField(name) {
			unknown[name] = append(json.RawMessage(nil), value...)
		}
	}
	cursor.Unknown = unknown
	*p = ProjectLabelPage{
		CursorPage:   cursor,
		GroupedBy:    grouped.GroupedBy,
		SubGroupedBy: grouped.SubGroupedBy,
		TotalCount:   grouped.TotalCount,
		present:      present,
	}
	return nil
}

// MarshalJSON emits the shared cursor envelope, grouping metadata, and
// retained unknown page fields without losing explicit nulls or omissions.
func (p ProjectLabelPage) MarshalJSON() ([]byte, error) {
	type cursorAlias plane.CursorPage[ProjectLabel]
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
	for name, value := range groupedFields {
		fields[name] = value
	}
	if p.present != nil {
		presentFields := make(map[string]json.RawMessage, len(p.present))
		for name := range p.present {
			if value, ok := fields[name]; ok {
				presentFields[name] = value
			}
		}
		fields = presentFields
	}
	for name, value := range p.CursorPage.Unknown {
		if _, exists := fields[name]; !exists {
			fields[name] = append(json.RawMessage(nil), value...)
		}
	}
	return json.Marshal(fields)
}

func containsProjectLabelPageField(target string) bool {
	for _, field := range projectLabelPageKnownFields {
		if field == target {
			return true
		}
	}
	return false
}

// CreateProjectLabelRequest contains the documented create body. Name is the
// only required property; pointer optionals preserve empty strings and zero.
type CreateProjectLabelRequest struct {
	Name        string   `json:"name"`
	Description *string  `json:"description,omitempty"`
	Color       *string  `json:"color,omitempty"`
	SortOrder   *float64 `json:"sort_order,omitempty"`
}

func (r CreateProjectLabelRequest) validate() error {
	if strings.TrimSpace(r.Name) == "" {
		return fmt.Errorf("project label name is required")
	}
	return nil
}

// UpdateProjectLabelRequest contains all documented optional PATCH fields.
// Pointer fields distinguish omission from explicit empty/zero values.
type UpdateProjectLabelRequest struct {
	Name        *string  `json:"name,omitempty"`
	Description *string  `json:"description,omitempty"`
	Color       *string  `json:"color,omitempty"`
	SortOrder   *float64 `json:"sort_order,omitempty"`
}
