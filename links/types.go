// Package links contains the route-specific contracts and client for Plane
// Work Item Links. Transport, authentication, pagination mechanics, output
// dispatch, and safe errors remain owned by planeshift/plane and the existing
// command/config packages.
package links

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"planeshift/plane"
)

// Link is the documented external-link response. Association and metadata
// values remain raw JSON because Plane may return an ID, object, array, or
// explicit null depending on expansion and server version. Unknown fields are
// retained so future response data is not lost by a typed round trip.
type Link struct {
	ID        *string                    `json:"id,omitempty"`
	Title     *string                    `json:"title,omitempty"`
	URL       *string                    `json:"url,omitempty"`
	Metadata  json.RawMessage            `json:"metadata,omitempty"`
	CreatedAt *string                    `json:"created_at,omitempty"`
	UpdatedAt *string                    `json:"updated_at,omitempty"`
	CreatedBy json.RawMessage            `json:"created_by,omitempty"`
	UpdatedBy json.RawMessage            `json:"updated_by,omitempty"`
	Project   json.RawMessage            `json:"project,omitempty"`
	Workspace json.RawMessage            `json:"workspace,omitempty"`
	Issue     json.RawMessage            `json:"issue,omitempty"`
	Unknown   map[string]json.RawMessage `json:"-"`
	present   map[string]bool            `json:"-"`
}

var linkKnownFields = map[string]struct{}{
	"id": {}, "title": {}, "url": {}, "metadata": {}, "created_at": {},
	"updated_at": {}, "created_by": {}, "updated_by": {}, "project": {},
	"workspace": {}, "issue": {},
}

// UnmarshalJSON preserves presence, nulls, and unknown fields.
func (l *Link) UnmarshalJSON(data []byte) error {
	type linkAlias Link
	var decoded linkAlias
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
		if _, known := linkKnownFields[key]; known {
			decoded.present[key] = true
			continue
		}
		decoded.Unknown[key] = append(json.RawMessage(nil), value...)
	}
	*l = Link(decoded)
	return nil
}

// MarshalJSON preserves explicit nulls and unknown fields without allowing an
// unknown field to override a typed response member.
func (l Link) MarshalJSON() ([]byte, error) {
	type linkAlias Link
	data, err := json.Marshal(linkAlias(l))
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

// LinkPage is the cursor envelope returned by collection operations. It
// retains grouping/count fields, explicit extra_stats values, and unknown
// envelope members in addition to the shared cursor metadata.
type LinkPage struct {
	plane.CursorPage[Link]
	GroupedBy    *string                    `json:"grouped_by,omitempty"`
	SubGroupedBy *string                    `json:"sub_grouped_by,omitempty"`
	TotalCount   *int                       `json:"total_count,omitempty"`
	Unknown      map[string]json.RawMessage `json:"-"`
	present      map[string]bool            `json:"-"`
}

var linkPageKnownFields = map[string]struct{}{
	"next_cursor": {}, "prev_cursor": {}, "next_page_results": {},
	"prev_page_results": {}, "count": {}, "total_pages": {},
	"total_results": {}, "extra_stats": {}, "results": {},
	"grouped_by": {}, "sub_grouped_by": {}, "total_count": {},
}

// UnmarshalJSON retains all documented and forward-compatible page members.
func (p *LinkPage) UnmarshalJSON(data []byte) error {
	var cursor plane.CursorPage[Link]
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
		if _, known := linkPageKnownFields[key]; known {
			present[key] = true
			continue
		}
		unknown[key] = append(json.RawMessage(nil), value...)
	}
	cursor.Unknown = nil
	*p = LinkPage{
		CursorPage: cursor, GroupedBy: grouped.GroupedBy,
		SubGroupedBy: grouped.SubGroupedBy, TotalCount: grouped.TotalCount,
		Unknown: unknown, present: present,
	}
	return nil
}

// MarshalJSON emits decoded fields according to their original presence and
// emits non-zero manually built fields, retaining unknown envelope members.
func (p LinkPage) MarshalJSON() ([]byte, error) {
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

// LinkDetailResponse is the page-shaped response documented by the detail
// operation. It is distinct from LinkPage so callers can preserve the
// operation-level contract even though the envelopes currently share fields.
type LinkDetailResponse struct {
	LinkPage
}

func (r *LinkDetailResponse) UnmarshalJSON(data []byte) error {
	var page LinkPage
	if err := json.Unmarshal(data, &page); err != nil {
		return err
	}
	r.LinkPage = page
	return nil
}

func (r LinkDetailResponse) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.LinkPage)
}

// CreateLinkRequest is the documented create body. URL is required; title is
// optional and remains presence-aware for explicit empty values.
type CreateLinkRequest struct {
	URL   string  `json:"url"`
	Title *string `json:"title,omitempty"`
}

func (r CreateLinkRequest) validate() error {
	if strings.TrimSpace(r.URL) == "" {
		return fmt.Errorf("link url is required")
	}
	return nil
}

// UpdateLinkRequest is the documented partial update body. Nil fields are
// omitted while non-nil pointers preserve explicit empty strings.
type UpdateLinkRequest struct {
	URL   *string `json:"url,omitempty"`
	Title *string `json:"title,omitempty"`
}

// ListOptions contains only the documented collection query controls.
type ListOptions struct {
	Cursor  string
	PerPage int
	Fields  string
	Expand  string
	OrderBy string
}

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

// DetailOptions contains the documented detail GET controls. order_by is
// intentionally absent because the detail operation does not permit it.
type DetailOptions struct {
	Cursor  string
	PerPage int
	Fields  string
	Expand  string
}

func (o DetailOptions) Query() (url.Values, error) {
	return plane.WithPagination(nil, plane.PaginationOptions{
		Cursor: o.Cursor, PerPage: o.PerPage, Fields: o.Fields, Expand: o.Expand,
	})
}
