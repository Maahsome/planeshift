// Package activities contains the lossless contracts and route adapter for
// Plane Work Item Activity history. Shared transport, authentication,
// pagination mechanics, output dispatch, and bounded errors remain owned by
// planeshift/plane and the existing command/config packages.
package activities

import (
	"encoding/json"
	"net/url"

	"planeshift/plane"
)

// Activity is one immutable work-item history record. Values that Plane can
// expand into strings, objects, arrays, or null remain raw JSON so callers do
// not lose the original activity history during typed round trips.
type Activity struct {
	ID            *string                    `json:"id,omitempty"`
	CreatedAt     *string                    `json:"created_at,omitempty"`
	UpdatedAt     *string                    `json:"updated_at,omitempty"`
	DeletedAt     *string                    `json:"deleted_at,omitempty"`
	Verb          *string                    `json:"verb,omitempty"`
	Field         *string                    `json:"field,omitempty"`
	OldValue      json.RawMessage            `json:"old_value,omitempty"`
	NewValue      json.RawMessage            `json:"new_value,omitempty"`
	Comment       *string                    `json:"comment,omitempty"`
	Attachments   json.RawMessage            `json:"attachments,omitempty"`
	OldIdentifier json.RawMessage            `json:"old_identifier,omitempty"`
	NewIdentifier json.RawMessage            `json:"new_identifier,omitempty"`
	Epoch         *float64                   `json:"epoch,omitempty"`
	Project       json.RawMessage            `json:"project,omitempty"`
	Workspace     json.RawMessage            `json:"workspace,omitempty"`
	Issue         json.RawMessage            `json:"issue,omitempty"`
	IssueComment  json.RawMessage            `json:"issue_comment,omitempty"`
	Actor         json.RawMessage            `json:"actor,omitempty"`
	Unknown       map[string]json.RawMessage `json:"-"`
	present       map[string]bool            `json:"-"`
}

var activityKnownFields = map[string]struct{}{
	"id": {}, "created_at": {}, "updated_at": {}, "deleted_at": {},
	"verb": {}, "field": {}, "old_value": {}, "new_value": {},
	"comment": {}, "attachments": {}, "old_identifier": {},
	"new_identifier": {}, "epoch": {}, "project": {}, "workspace": {},
	"issue": {}, "issue_comment": {}, "actor": {},
}

// UnmarshalJSON preserves field presence, explicit nulls, and unknown fields.
func (a *Activity) UnmarshalJSON(data []byte) error {
	type activityAlias Activity
	var decoded activityAlias
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	var allFields map[string]json.RawMessage
	if err := json.Unmarshal(data, &allFields); err != nil {
		return err
	}
	decoded.Unknown = make(map[string]json.RawMessage)
	decoded.present = make(map[string]bool, len(allFields))
	for key, value := range allFields {
		if _, known := activityKnownFields[key]; known {
			decoded.present[key] = true
			continue
		}
		decoded.Unknown[key] = append(json.RawMessage(nil), value...)
	}
	*a = Activity(decoded)
	return nil
}

// MarshalJSON preserves explicit nulls and unknown fields without allowing an
// unknown field to override a typed response member.
func (a Activity) MarshalJSON() ([]byte, error) {
	type activityAlias Activity
	data, err := json.Marshal(activityAlias(a))
	if err != nil {
		return nil, err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return nil, err
	}
	if a.present != nil {
		presentFields := make(map[string]json.RawMessage, len(a.present))
		for key := range a.present {
			if value, ok := fields[key]; ok {
				presentFields[key] = value
			} else {
				presentFields[key] = json.RawMessage("null")
			}
		}
		fields = presentFields
	}
	for key, value := range a.Unknown {
		if _, exists := fields[key]; !exists {
			fields[key] = append(json.RawMessage(nil), value...)
		}
	}
	return json.Marshal(fields)
}

// ActivityPage is the cursor envelope returned by activity collection and
// detail operations. It retains grouping/count metadata, extra_stats, and
// unknown envelope members in addition to shared cursor metadata.
type ActivityPage struct {
	plane.CursorPage[Activity]
	GroupedBy    *string                    `json:"grouped_by,omitempty"`
	SubGroupedBy *string                    `json:"sub_grouped_by,omitempty"`
	TotalCount   *int                       `json:"total_count,omitempty"`
	Unknown      map[string]json.RawMessage `json:"-"`
	present      map[string]bool            `json:"-"`
}

var activityPageKnownFields = map[string]struct{}{
	"next_cursor": {}, "prev_cursor": {}, "next_page_results": {},
	"prev_page_results": {}, "count": {}, "total_pages": {},
	"total_results": {}, "extra_stats": {}, "results": {},
	"grouped_by": {}, "sub_grouped_by": {}, "total_count": {},
}

// UnmarshalJSON retains all documented and forward-compatible page members.
func (p *ActivityPage) UnmarshalJSON(data []byte) error {
	var cursor plane.CursorPage[Activity]
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
		if _, known := activityPageKnownFields[key]; known {
			present[key] = true
			continue
		}
		unknown[key] = append(json.RawMessage(nil), value...)
	}
	cursor.Unknown = nil
	*p = ActivityPage{
		CursorPage: cursor, GroupedBy: grouped.GroupedBy,
		SubGroupedBy: grouped.SubGroupedBy, TotalCount: grouped.TotalCount,
		Unknown: unknown, present: present,
	}
	return nil
}

// MarshalJSON emits decoded fields according to their original presence and
// retains unknown envelope members.
func (p ActivityPage) MarshalJSON() ([]byte, error) {
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

// ActivityDetailResponse names the page-shaped detail response documented by
// Plane while preserving the underlying ActivityPage for callers.
type ActivityDetailResponse struct {
	ActivityPage
}

func (r *ActivityDetailResponse) UnmarshalJSON(data []byte) error {
	var page ActivityPage
	if err := json.Unmarshal(data, &page); err != nil {
		return err
	}
	r.ActivityPage = page
	return nil
}

func (r ActivityDetailResponse) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.ActivityPage)
}

// ListOptions contains exactly the documented collection query controls.
type ListOptions struct {
	Cursor  string
	PerPage int
	Fields  string
	Expand  string
	OrderBy string
}

func (o ListOptions) Query() (url.Values, error) {
	return activityQuery(o.Cursor, o.PerPage, o.Fields, o.Expand, o.OrderBy)
}

// DetailOptions contains exactly the documented detail query controls.
type DetailOptions struct {
	Cursor  string
	PerPage int
	Fields  string
	Expand  string
	OrderBy string
}

func (o DetailOptions) Query() (url.Values, error) {
	return activityQuery(o.Cursor, o.PerPage, o.Fields, o.Expand, o.OrderBy)
}

func activityQuery(cursor string, perPage int, fields, expand, orderBy string) (url.Values, error) {
	query, err := plane.WithPagination(nil, plane.PaginationOptions{
		Cursor: cursor, PerPage: perPage, Fields: fields, Expand: expand,
	})
	if err != nil {
		return nil, err
	}
	if orderBy != "" {
		query.Set("order_by", orderBy)
	}
	return query, nil
}
