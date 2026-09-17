// Package labels contains the route-specific contracts for project-scoped
// Plane labels. Transport, authentication, pagination validation, and errors
// remain owned by the shared plane package.
package labels

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"planeshift/plane"
)

// Label is a project-scoped label response. Stable scalar fields use typed
// pointers so nullable and omitted values remain distinguishable. Association
// fields remain raw JSON because Plane may return an ID, null, or an expanded
// object depending on the requested response shape. Unknown response fields
// are copied and retained for forward-compatible output.
type Label struct {
	ID             *string                    `json:"id,omitempty"`
	CreatedAt      *string                    `json:"created_at,omitempty"`
	UpdatedAt      *string                    `json:"updated_at,omitempty"`
	Name           *string                    `json:"name,omitempty"`
	Description    *string                    `json:"description,omitempty"`
	Color          *string                    `json:"color,omitempty"`
	SortOrder      *float64                   `json:"sort_order,omitempty"`
	CreatedBy      json.RawMessage            `json:"created_by,omitempty"`
	UpdatedBy      json.RawMessage            `json:"updated_by,omitempty"`
	Project        json.RawMessage            `json:"project,omitempty"`
	Workspace      json.RawMessage            `json:"workspace,omitempty"`
	Parent         json.RawMessage            `json:"parent,omitempty"`
	ExternalSource *string                    `json:"external_source,omitempty"`
	ExternalID     *string                    `json:"external_id,omitempty"`
	Unknown        map[string]json.RawMessage `json:"-"`
	present        map[string]bool            `json:"-"`
}

var labelKnownFields = map[string]struct{}{
	"id": {}, "created_at": {}, "updated_at": {}, "name": {},
	"description": {}, "color": {}, "sort_order": {}, "created_by": {},
	"updated_by": {}, "project": {}, "workspace": {}, "parent": {},
	"external_source": {}, "external_id": {},
}

// UnmarshalJSON validates documented scalar fields and retains all unknown
// values without converting numbers or explicit nulls through interface{}.
func (l *Label) UnmarshalJSON(data []byte) error {
	type labelAlias Label
	var decoded labelAlias
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	var allFields map[string]json.RawMessage
	if err := json.Unmarshal(data, &allFields); err != nil {
		return err
	}
	unknown := make(map[string]json.RawMessage)
	present := make(map[string]bool, len(allFields))
	for key, value := range allFields {
		if _, known := labelKnownFields[key]; known {
			present[key] = true
			continue
		}
		unknown[key] = append(json.RawMessage(nil), value...)
	}
	decoded.Unknown = unknown
	decoded.present = present
	*l = Label(decoded)
	return nil
}

// MarshalJSON preserves explicit nulls, omitted response fields, and unknown
// values while preventing unknown data from overriding typed fields.
func (l Label) MarshalJSON() ([]byte, error) {
	type labelAlias Label
	data, err := json.Marshal(labelAlias(l))
	if err != nil {
		return nil, err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return nil, err
	}
	if l.present != nil {
		presentFields := make(map[string]json.RawMessage, len(l.present))
		for key := range l.present {
			if value, ok := fields[key]; ok {
				presentFields[key] = value
			} else {
				presentFields[key] = json.RawMessage("null")
			}
		}
		fields = presentFields
	}
	for key, value := range l.Unknown {
		if _, exists := fields[key]; !exists {
			fields[key] = append(json.RawMessage(nil), value...)
		}
	}
	return json.Marshal(fields)
}

// LabelPage is the cursor page returned by the project label collection.
// It retains all shared cursor metadata, label results, grouping metadata,
// explicit extra_stats values, and unknown page keys.
type LabelPage struct {
	plane.CursorPage[Label]
	GroupedBy    *string         `json:"grouped_by,omitempty"`
	SubGroupedBy *string         `json:"sub_grouped_by,omitempty"`
	TotalCount   *int            `json:"total_count,omitempty"`
	present      map[string]bool `json:"-"`
}

var labelPageKnownFields = map[string]struct{}{
	"next_cursor": {}, "prev_cursor": {}, "next_page_results": {},
	"prev_page_results": {}, "count": {}, "total_pages": {},
	"total_results": {}, "extra_stats": {}, "results": {},
	"grouped_by": {}, "sub_grouped_by": {}, "total_count": {},
}

// UnmarshalJSON retains grouping metadata and unknown page keys.
func (p *LabelPage) UnmarshalJSON(data []byte) error {
	var cursor plane.CursorPage[Label]
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
	var allFields map[string]json.RawMessage
	if err := json.Unmarshal(data, &allFields); err != nil {
		return err
	}
	unknown := make(map[string]json.RawMessage)
	present := make(map[string]bool, len(allFields))
	for key, value := range allFields {
		if _, known := labelPageKnownFields[key]; known {
			present[key] = true
			continue
		}
		unknown[key] = append(json.RawMessage(nil), value...)
	}
	cursor.Unknown = unknown
	*p = LabelPage{
		CursorPage: cursor,
		GroupedBy:  grouped.GroupedBy, SubGroupedBy: grouped.SubGroupedBy,
		TotalCount: grouped.TotalCount, present: present,
	}
	return nil
}

// MarshalJSON emits only page fields that were present in a decoded response,
// while manually built pages emit their non-zero/non-nil values.
func (p LabelPage) MarshalJSON() ([]byte, error) {
	fields := make(map[string]json.RawMessage)
	add := func(key string, value any, include bool) error {
		if !include {
			return nil
		}
		data, err := json.Marshal(value)
		if err != nil {
			return err
		}
		fields[key] = data
		return nil
	}
	include := func(key string, valuePresent bool) bool {
		if p.present != nil {
			_, ok := p.present[key]
			return ok
		}
		return valuePresent
	}

	if err := add("next_cursor", p.NextCursor, include("next_cursor", p.NextCursor != nil)); err != nil {
		return nil, err
	}
	if err := add("prev_cursor", p.PrevCursor, include("prev_cursor", p.PrevCursor != nil)); err != nil {
		return nil, err
	}
	if err := add("next_page_results", p.NextPageResults, include("next_page_results", p.NextPageResults != nil)); err != nil {
		return nil, err
	}
	if err := add("prev_page_results", p.PrevPageResults, include("prev_page_results", p.PrevPageResults != nil)); err != nil {
		return nil, err
	}
	if err := add("count", p.Count, include("count", p.Count != nil)); err != nil {
		return nil, err
	}
	if err := add("total_pages", p.TotalPages, include("total_pages", p.TotalPages != nil)); err != nil {
		return nil, err
	}
	if err := add("total_results", p.TotalResults, include("total_results", p.TotalResults != nil)); err != nil {
		return nil, err
	}
	extraStats := any(p.ExtraStats)
	if len(p.ExtraStats) == 0 {
		extraStats = nil
	}
	if err := add("extra_stats", extraStats, include("extra_stats", len(p.ExtraStats) > 0)); err != nil {
		return nil, err
	}
	if err := add("results", p.Results, include("results", p.Results != nil)); err != nil {
		return nil, err
	}
	if err := add("grouped_by", p.GroupedBy, include("grouped_by", p.GroupedBy != nil)); err != nil {
		return nil, err
	}
	if err := add("sub_grouped_by", p.SubGroupedBy, include("sub_grouped_by", p.SubGroupedBy != nil)); err != nil {
		return nil, err
	}
	if err := add("total_count", p.TotalCount, include("total_count", p.TotalCount != nil)); err != nil {
		return nil, err
	}
	for key, value := range p.Unknown {
		if _, exists := fields[key]; !exists {
			fields[key] = append(json.RawMessage(nil), value...)
		}
	}
	return json.Marshal(fields)
}

// CreateLabelRequest is the documented create body. Name is the only required
// field; pointer optionals retain explicit empty strings and sort_order zero.
type CreateLabelRequest struct {
	Name           string   `json:"name"`
	Color          *string  `json:"color,omitempty"`
	Description    *string  `json:"description,omitempty"`
	ExternalSource *string  `json:"external_source,omitempty"`
	ExternalID     *string  `json:"external_id,omitempty"`
	Parent         *string  `json:"parent,omitempty"`
	SortOrder      *float64 `json:"sort_order,omitempty"`
}

func (r CreateLabelRequest) validate() error {
	if strings.TrimSpace(r.Name) == "" {
		return fmt.Errorf("label name is required")
	}
	return nil
}

// UpdateLabelRequest contains all documented optional PATCH fields. Nil omits
// a key; non-nil pointers preserve explicit empty strings and sort_order zero.
type UpdateLabelRequest struct {
	Name           *string  `json:"name,omitempty"`
	Color          *string  `json:"color,omitempty"`
	Description    *string  `json:"description,omitempty"`
	ExternalSource *string  `json:"external_source,omitempty"`
	ExternalID     *string  `json:"external_id,omitempty"`
	Parent         *string  `json:"parent,omitempty"`
	SortOrder      *float64 `json:"sort_order,omitempty"`
}

// ListOptions contains only the documented project-label collection query
// controls. Name and color are intentionally not exposed because they are
// mentioned only in page prose, not in the operation's explicit parameter list.
type ListOptions struct {
	Cursor  string
	PerPage int
	Fields  string
	Expand  string
	OrderBy string
}

// Query delegates shared pagination validation and adds the label-specific
// collection ordering value without discarding shared query controls.
func (o ListOptions) Query() (url.Values, error) {
	query, err := plane.WithPagination(nil, plane.PaginationOptions{
		Cursor: o.Cursor, PerPage: o.PerPage, Fields: o.Fields, Expand: o.Expand,
	})
	if err != nil {
		return nil, err
	}
	if o.OrderBy != "" {
		query.Set("order_by", o.OrderBy)
	}
	return query, nil
}
