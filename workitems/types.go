// Package workitems contains the route-specific Work Item contracts and
// client. Transport, authentication, pagination mechanics, and safe errors
// remain in planeshift/plane.
package workitems

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"planeshift/plane"
)

// Work Item wire contract reconciliation:
//   - PUBLIC_API.txt is authoritative for the exact 16 routes owned here.
//   - Current detail routes use {work_item_id}; compatibility detail routes
//     use {issue_id}; the documentation's generic resource_id is not used.
//   - Search uses search, limit, project_id, and workspace_search only.
//   - Identifier lookup has two independently escaped path segments and only
//     its documented expand control.
//   - List and relation-list operations use the documented cursor controls;
//     detail reads use the documented detail filters, not cursor pagination.
//   - The official pages document 201 for create/relation and 204 for delete.
// Dynamic response values are raw JSON so UUIDs, expanded objects, arrays, and
// explicit nulls remain lossless.

// WorkItem is the documented Work Item response with lossless dynamic fields.
type WorkItem struct {
	ID                  string                     `json:"id"`
	Name                string                     `json:"name"`
	CreatedAt           *string                    `json:"created_at"`
	UpdatedAt           *string                    `json:"updated_at"`
	EstimatePoint       json.RawMessage            `json:"estimate_point"`
	DescriptionHTML     *string                    `json:"description_html"`
	DescriptionStripped *string                    `json:"description_stripped"`
	Priority            *string                    `json:"priority"`
	StartDate           *string                    `json:"start_date"`
	TargetDate          *string                    `json:"target_date"`
	SequenceID          json.RawMessage            `json:"sequence_id"`
	SortOrder           json.RawMessage            `json:"sort_order"`
	CompletedAt         *string                    `json:"completed_at"`
	ArchivedAt          *string                    `json:"archived_at"`
	LastActivityAt      *string                    `json:"last_activity_at"`
	CreatedBy           json.RawMessage            `json:"created_by"`
	UpdatedBy           json.RawMessage            `json:"updated_by"`
	Project             json.RawMessage            `json:"project"`
	Workspace           json.RawMessage            `json:"workspace"`
	Parent              json.RawMessage            `json:"parent"`
	State               json.RawMessage            `json:"state"`
	Assignees           json.RawMessage            `json:"assignees"`
	Labels              json.RawMessage            `json:"labels"`
	Type                json.RawMessage            `json:"type"`
	Module              json.RawMessage            `json:"module"`
	IsDraft             *bool                      `json:"is_draft"`
	ExternalSource      *string                    `json:"external_source"`
	ExternalID          *string                    `json:"external_id"`
	Unknown             map[string]json.RawMessage `json:"-"`
	present             map[string]bool            `json:"-"`
}

var workItemKnownFields = []string{
	"id", "name", "created_at", "updated_at", "estimate_point", "description_html",
	"description_stripped", "priority", "start_date", "target_date", "sequence_id",
	"sort_order", "completed_at", "archived_at", "last_activity_at", "created_by",
	"updated_by", "project", "workspace", "parent", "state", "assignees", "labels",
	"type", "module", "is_draft", "external_source", "external_id",
}

func (w *WorkItem) UnmarshalJSON(data []byte) error {
	type alias WorkItem
	var decoded alias
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	unknown := make(map[string]json.RawMessage)
	present := make(map[string]bool)
	for key, value := range fields {
		if contains(workItemKnownFields, key) {
			present[key] = true
		} else {
			unknown[key] = append(json.RawMessage(nil), value...)
		}
	}
	decoded.Unknown = unknown
	decoded.present = present
	*w = WorkItem(decoded)
	return nil
}

func (w WorkItem) MarshalJSON() ([]byte, error) {
	type alias WorkItem
	data, err := json.Marshal(alias(w))
	if err != nil {
		return nil, err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return nil, err
	}
	if w.present != nil {
		selected := make(map[string]json.RawMessage, len(fields))
		for key := range w.present {
			selected[key] = fields[key]
		}
		fields = selected
	}
	for key, value := range w.Unknown {
		if _, exists := fields[key]; !exists {
			fields[key] = append(json.RawMessage(nil), value...)
		}
	}
	return json.Marshal(fields)
}

// WorkItemPage is the cursor page returned by the project collection. The
// additional grouping/count values are retained because they occur in Plane's
// list examples but are not part of the generic cursor envelope.
type WorkItemPage struct {
	plane.CursorPage[WorkItem]
	GroupedBy    *string `json:"grouped_by"`
	SubGroupedBy *string `json:"sub_grouped_by"`
	TotalCount   *int    `json:"total_count"`
}

var workItemPageKnownFields = append(append([]string(nil), workItemKnownFields...),
	"next_cursor", "prev_cursor", "next_page_results", "prev_page_results", "count",
	"total_pages", "total_results", "extra_stats", "results", "grouped_by",
	"sub_grouped_by", "total_count")

func (p *WorkItemPage) UnmarshalJSON(data []byte) error {
	var cursor plane.CursorPage[WorkItem]
	if err := json.Unmarshal(data, &cursor); err != nil {
		return err
	}
	var extras struct {
		GroupedBy    *string `json:"grouped_by"`
		SubGroupedBy *string `json:"sub_grouped_by"`
		TotalCount   *int    `json:"total_count"`
	}
	if err := json.Unmarshal(data, &extras); err != nil {
		return err
	}
	var unknown map[string]json.RawMessage
	if err := json.Unmarshal(data, &unknown); err != nil {
		return err
	}
	for _, key := range workItemPageKnownFields {
		delete(unknown, key)
	}
	cursor.Unknown = unknown
	*p = WorkItemPage{CursorPage: cursor, GroupedBy: extras.GroupedBy,
		SubGroupedBy: extras.SubGroupedBy, TotalCount: extras.TotalCount}
	return nil
}

func (p WorkItemPage) MarshalJSON() ([]byte, error) {
	type cursorAlias plane.CursorPage[WorkItem]
	data, err := json.Marshal(cursorAlias(p.CursorPage))
	if err != nil {
		return nil, err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return nil, err
	}
	extras, err := json.Marshal(struct {
		GroupedBy    *string `json:"grouped_by"`
		SubGroupedBy *string `json:"sub_grouped_by"`
		TotalCount   *int    `json:"total_count"`
	}{p.GroupedBy, p.SubGroupedBy, p.TotalCount})
	if err != nil {
		return nil, err
	}
	var extraFields map[string]json.RawMessage
	if err := json.Unmarshal(extras, &extraFields); err != nil {
		return nil, err
	}
	for key, value := range extraFields {
		fields[key] = value
	}
	for key, value := range p.CursorPage.Unknown {
		if _, exists := fields[key]; !exists {
			fields[key] = append(json.RawMessage(nil), value...)
		}
	}
	return json.Marshal(fields)
}

// WorkItemSearchResult is one item in the documented search envelope.
type WorkItemSearchResult struct {
	ID                string                     `json:"id"`
	Name              *string                    `json:"name"`
	SequenceID        json.RawMessage            `json:"sequence_id"`
	ProjectIdentifier *string                    `json:"project__identifier"`
	ProjectID         json.RawMessage            `json:"project_id"`
	WorkspaceSlug     *string                    `json:"workspace__slug"`
	Unknown           map[string]json.RawMessage `json:"-"`
	present           map[string]bool            `json:"-"`
}

var searchResultKnownFields = []string{"id", "name", "sequence_id", "project__identifier", "project_id", "workspace__slug"}

func (r *WorkItemSearchResult) UnmarshalJSON(data []byte) error {
	type alias WorkItemSearchResult
	var decoded alias
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	decoded.Unknown = make(map[string]json.RawMessage)
	decoded.present = make(map[string]bool)
	for key, value := range fields {
		if contains(searchResultKnownFields, key) {
			decoded.present[key] = true
		} else {
			decoded.Unknown[key] = append(json.RawMessage(nil), value...)
		}
	}
	*r = WorkItemSearchResult(decoded)
	return nil
}

func (r WorkItemSearchResult) MarshalJSON() ([]byte, error) {
	type alias WorkItemSearchResult
	data, err := json.Marshal(alias(r))
	if err != nil {
		return nil, err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return nil, err
	}
	if r.present != nil {
		selected := make(map[string]json.RawMessage)
		for key := range r.present {
			selected[key] = fields[key]
		}
		fields = selected
	}
	for key, value := range r.Unknown {
		if _, exists := fields[key]; !exists {
			fields[key] = append(json.RawMessage(nil), value...)
		}
	}
	return json.Marshal(fields)
}

// WorkItemSearchResponse is the documented non-paginated search envelope.
type WorkItemSearchResponse struct {
	Issues  []WorkItemSearchResult     `json:"issues"`
	Results []WorkItemSearchResult     `json:"results,omitempty"`
	Unknown map[string]json.RawMessage `json:"-"`
}

func (r *WorkItemSearchResponse) UnmarshalJSON(data []byte) error {
	type alias WorkItemSearchResponse
	var decoded alias
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	delete(fields, "issues")
	delete(fields, "results")
	decoded.Unknown = fields
	*r = WorkItemSearchResponse(decoded)
	return nil
}

func (r WorkItemSearchResponse) MarshalJSON() ([]byte, error) {
	type alias WorkItemSearchResponse
	data, err := json.Marshal(alias(r))
	if err != nil {
		return nil, err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return nil, err
	}
	for key, value := range r.Unknown {
		if _, exists := fields[key]; !exists {
			fields[key] = append(json.RawMessage(nil), value...)
		}
	}
	return json.Marshal(fields)
}

// WorkItemRelation is the documented relation result. Unknown relation keys
// remain available for future or expanded server responses.
type WorkItemRelation struct {
	ID           string                     `json:"id"`
	Name         *string                    `json:"name"`
	SequenceID   json.RawMessage            `json:"sequence_id"`
	ProjectID    json.RawMessage            `json:"project_id"`
	RelationType *string                    `json:"relation_type"`
	StateID      json.RawMessage            `json:"state_id"`
	Priority     *string                    `json:"priority"`
	TypeID       json.RawMessage            `json:"type_id"`
	IsEpic       *bool                      `json:"is_epic"`
	CreatedAt    *string                    `json:"created_at"`
	UpdatedAt    *string                    `json:"updated_at"`
	CreatedBy    json.RawMessage            `json:"created_by"`
	UpdatedBy    json.RawMessage            `json:"updated_by"`
	Unknown      map[string]json.RawMessage `json:"-"`
	present      map[string]bool            `json:"-"`
}

var relationKnownFields = []string{"id", "name", "sequence_id", "project_id", "relation_type", "state_id", "priority", "type_id", "is_epic", "created_at", "updated_at", "created_by", "updated_by"}

func (r *WorkItemRelation) UnmarshalJSON(data []byte) error {
	type alias WorkItemRelation
	var decoded alias
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	decoded.Unknown = make(map[string]json.RawMessage)
	decoded.present = make(map[string]bool)
	for key, value := range fields {
		if contains(relationKnownFields, key) {
			decoded.present[key] = true
		} else {
			decoded.Unknown[key] = append(json.RawMessage(nil), value...)
		}
	}
	*r = WorkItemRelation(decoded)
	return nil
}

func (r WorkItemRelation) MarshalJSON() ([]byte, error) {
	type alias WorkItemRelation
	data, err := json.Marshal(alias(r))
	if err != nil {
		return nil, err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return nil, err
	}
	if r.present != nil {
		selected := make(map[string]json.RawMessage)
		for key := range r.present {
			selected[key] = fields[key]
		}
		fields = selected
	}
	for key, value := range r.Unknown {
		if _, exists := fields[key]; !exists {
			fields[key] = append(json.RawMessage(nil), value...)
		}
	}
	return json.Marshal(fields)
}

// WorkItemRelationPage is the cursor response for relation listing.
type WorkItemRelationPage struct {
	plane.CursorPage[WorkItemRelation]
	Unknown map[string]json.RawMessage `json:"-"`
}

func (p *WorkItemRelationPage) UnmarshalJSON(data []byte) error {
	type pageAlias plane.CursorPage[WorkItemRelation]
	var page pageAlias
	if err := json.Unmarshal(data, &page); err != nil {
		return err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	for _, key := range []string{"next_cursor", "prev_cursor", "next_page_results", "prev_page_results", "count", "total_pages", "total_results", "extra_stats", "results"} {
		delete(fields, key)
	}
	page.Unknown = fields
	*p = WorkItemRelationPage{CursorPage: plane.CursorPage[WorkItemRelation](page), Unknown: fields}
	return nil
}

func (p WorkItemRelationPage) MarshalJSON() ([]byte, error) {
	type pageAlias plane.CursorPage[WorkItemRelation]
	data, err := json.Marshal(pageAlias(p.CursorPage))
	if err != nil {
		return nil, err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return nil, err
	}
	for key, value := range p.CursorPage.Unknown {
		if _, exists := fields[key]; !exists {
			fields[key] = append(json.RawMessage(nil), value...)
		}
	}
	for key, value := range p.Unknown {
		if _, exists := fields[key]; !exists {
			fields[key] = append(json.RawMessage(nil), value...)
		}
	}
	return json.Marshal(fields)
}

// CreateWorkItemRequest contains the documented create body. Name is the
// only required field. Pointers preserve false, zero, empty arrays, and empty
// strings; NullFields can explicitly emit JSON null for any named field.
type CreateWorkItemRequest struct {
	Assignees           *[]string        `json:"assignees,omitempty"`
	Labels              *[]string        `json:"labels,omitempty"`
	TypeID              *string          `json:"type_id,omitempty"`
	Parent              *string          `json:"parent,omitempty"`
	DeletedAt           *string          `json:"deleted_at,omitempty"`
	Point               *int             `json:"point,omitempty"`
	Name                string           `json:"name"`
	DescriptionHTML     *string          `json:"description_html,omitempty"`
	DescriptionStripped *string          `json:"description_stripped,omitempty"`
	Priority            *string          `json:"priority,omitempty"`
	StartDate           *string          `json:"start_date,omitempty"`
	TargetDate          *string          `json:"target_date,omitempty"`
	SequenceID          *int             `json:"sequence_id,omitempty"`
	SortOrder           *json.RawMessage `json:"sort_order,omitempty"`
	CompletedAt         *string          `json:"completed_at,omitempty"`
	ArchivedAt          *string          `json:"archived_at,omitempty"`
	LastActivityAt      *string          `json:"last_activity_at,omitempty"`
	IsDraft             *bool            `json:"is_draft,omitempty"`
	ExternalSource      *string          `json:"external_source,omitempty"`
	ExternalID          *string          `json:"external_id,omitempty"`
	CreatedBy           *string          `json:"created_by,omitempty"`
	State               *string          `json:"state,omitempty"`
	EstimatePoint       *json.RawMessage `json:"estimate_point,omitempty"`
	Type                *json.RawMessage `json:"type,omitempty"`
	NullFields          map[string]bool  `json:"-"`
}

func (r CreateWorkItemRequest) validate() error {
	if strings.TrimSpace(r.Name) == "" {
		return fmt.Errorf("work item name is required")
	}
	return nil
}

func (r CreateWorkItemRequest) MarshalJSON() ([]byte, error) {
	type alias CreateWorkItemRequest
	return marshalRequest(alias(r), r.NullFields)
}

// UpdateWorkItemRequest contains the documented PATCH fields. Use SetNull to
// intentionally clear a field; an untouched pointer remains omitted.
type UpdateWorkItemRequest struct {
	Assignees           *[]string        `json:"assignees,omitempty"`
	Labels              *[]string        `json:"labels,omitempty"`
	TypeID              *string          `json:"type_id,omitempty"`
	Parent              *string          `json:"parent,omitempty"`
	DeletedAt           *string          `json:"deleted_at,omitempty"`
	Point               *int             `json:"point,omitempty"`
	Name                *string          `json:"name,omitempty"`
	DescriptionHTML     *string          `json:"description_html,omitempty"`
	DescriptionStripped *string          `json:"description_stripped,omitempty"`
	Priority            *string          `json:"priority,omitempty"`
	StartDate           *string          `json:"start_date,omitempty"`
	TargetDate          *string          `json:"target_date,omitempty"`
	SequenceID          *int             `json:"sequence_id,omitempty"`
	SortOrder           *json.RawMessage `json:"sort_order,omitempty"`
	CompletedAt         *string          `json:"completed_at,omitempty"`
	ArchivedAt          *string          `json:"archived_at,omitempty"`
	LastActivityAt      *string          `json:"last_activity_at,omitempty"`
	IsDraft             *bool            `json:"is_draft,omitempty"`
	ExternalSource      *string          `json:"external_source,omitempty"`
	ExternalID          *string          `json:"external_id,omitempty"`
	CreatedBy           *string          `json:"created_by,omitempty"`
	State               *string          `json:"state,omitempty"`
	EstimatePoint       *json.RawMessage `json:"estimate_point,omitempty"`
	Type                *json.RawMessage `json:"type,omitempty"`
	NullFields          map[string]bool  `json:"-"`
}

// SetNull marks one documented request field for explicit JSON null.
func (r *UpdateWorkItemRequest) SetNull(field string) {
	if r.NullFields == nil {
		r.NullFields = make(map[string]bool)
	}
	r.NullFields[field] = true
}

func (r UpdateWorkItemRequest) MarshalJSON() ([]byte, error) {
	type alias UpdateWorkItemRequest
	return marshalRequest(alias(r), r.NullFields)
}

func marshalRequest(value any, nullFields map[string]bool) ([]byte, error) {
	type rawRequest map[string]json.RawMessage
	data, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	var fields rawRequest
	if err := json.Unmarshal(data, &fields); err != nil {
		return nil, err
	}
	for field, selected := range nullFields {
		if selected {
			fields[field] = json.RawMessage("null")
		}
	}
	return json.Marshal(fields)
}

// RelationType values documented by Plane.
type RelationType string

const (
	RelationBlocking     RelationType = "blocking"
	RelationBlockedBy    RelationType = "blocked_by"
	RelationDuplicate    RelationType = "duplicate"
	RelationRelatesTo    RelationType = "relates_to"
	RelationStartBefore  RelationType = "start_before"
	RelationStartAfter   RelationType = "start_after"
	RelationFinishBefore RelationType = "finish_before"
	RelationFinishAfter  RelationType = "finish_after"
)

// CreateWorkItemRelationRequest is the documented relation-create body.
type CreateWorkItemRelationRequest struct {
	RelationType RelationType `json:"relation_type"`
	Issues       []string     `json:"issues"`
}

func (r CreateWorkItemRelationRequest) validate() error {
	if !validRelationType(r.RelationType) {
		return fmt.Errorf("relation_type %q is not documented", r.RelationType)
	}
	if len(r.Issues) == 0 {
		return fmt.Errorf("relation issues are required")
	}
	return nil
}

// WorkItemRelationCreateResponse preserves the documented nested 201 JSON
// response shape instead of flattening it into a single relation.
type WorkItemRelationCreateResponse [][]WorkItemRelation

// ListOptions are the documented project collection controls.
type ListOptions struct {
	Cursor         string
	PerPage        int
	Fields         string
	Expand         string
	ExternalID     string
	ExternalSource string
	OrderBy        string
}

func (o ListOptions) Query() (url.Values, error) {
	query, err := plane.WithPagination(nil, plane.PaginationOptions{
		Cursor: o.Cursor, PerPage: o.PerPage, Fields: o.Fields, Expand: o.Expand,
	})
	if err != nil {
		return nil, err
	}
	setOptionalQuery(query, "external_id", o.ExternalID)
	setOptionalQuery(query, "external_source", o.ExternalSource)
	setOptionalQuery(query, "order_by", o.OrderBy)
	return query, nil
}

// DetailOptions are the documented detail-read controls.
type DetailOptions struct {
	Expand         string
	Fields         string
	ExternalID     string
	ExternalSource string
	OrderBy        string
}

func (o DetailOptions) Query() url.Values {
	query := make(url.Values)
	setOptionalQuery(query, "expand", o.Expand)
	setOptionalQuery(query, "fields", o.Fields)
	setOptionalQuery(query, "external_id", o.ExternalID)
	setOptionalQuery(query, "external_source", o.ExternalSource)
	setOptionalQuery(query, "order_by", o.OrderBy)
	return query
}

// IdentifierOptions contains the only documented identifier lookup control.
type IdentifierOptions struct{ Expand string }

func (o IdentifierOptions) Query() url.Values {
	query := make(url.Values)
	setOptionalQuery(query, "expand", o.Expand)
	return query
}

// SearchOptions contains the exact documented search controls.
type SearchOptions struct {
	Search          string
	Limit           int
	ProjectID       string
	WorkspaceSearch string
}

func (o SearchOptions) Query() (url.Values, error) {
	if strings.TrimSpace(o.Search) == "" {
		return nil, fmt.Errorf("search is required")
	}
	if o.Limit < 0 {
		return nil, fmt.Errorf("limit must not be negative")
	}
	query := url.Values{"search": {o.Search}}
	if o.Limit > 0 {
		query.Set("limit", fmt.Sprint(o.Limit))
	}
	setOptionalQuery(query, "project_id", o.ProjectID)
	setOptionalQuery(query, "workspace_search", o.WorkspaceSearch)
	return query, nil
}

// RelationListOptions are the documented relation collection controls.
type RelationListOptions struct {
	Cursor  string
	PerPage int
	Fields  string
	Expand  string
}

func (o RelationListOptions) Query() (url.Values, error) {
	return plane.WithPagination(nil, plane.PaginationOptions{
		Cursor: o.Cursor, PerPage: o.PerPage, Fields: o.Fields, Expand: o.Expand,
	})
}

func setOptionalQuery(query url.Values, key, value string) {
	if value != "" {
		query.Set(key, value)
	}
}

func validRelationType(value RelationType) bool {
	switch value {
	case RelationBlocking, RelationBlockedBy, RelationDuplicate, RelationRelatesTo,
		RelationStartBefore, RelationStartAfter, RelationFinishBefore, RelationFinishAfter:
		return true
	default:
		return false
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
