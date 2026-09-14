// Package projects contains the route-specific contracts and client for the
// documented Plane Projects operations. It depends on the shared plane.Client
// for transport, authentication, pagination, errors, and response metadata.
package projects

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"planeshift/plane"
)

// Project is the documented Project response. Pointer fields retain nullable
// values and presence for response properties. IconProp, IsDeployed, and
// DefaultState remain raw JSON because the documentation permits dynamic
// values (and documents is_deployed as both an integer and a boolean).
// Unknown retains future or undocumented response fields through a typed
// decode and subsequent JSON output.
type Project struct {
	ID                    string                     `json:"id"`
	Name                  string                     `json:"name"`
	Identifier            string                     `json:"identifier"`
	Description           *string                    `json:"description"`
	DescriptionText       *string                    `json:"description_text"`
	DescriptionHTML       *string                    `json:"description_html"`
	TotalMembers          *int                       `json:"total_members"`
	TotalCycles           *int                       `json:"total_cycles"`
	TotalModules          *int                       `json:"total_modules"`
	IsMember              *bool                      `json:"is_member"`
	MemberRole            *int                       `json:"member_role"`
	IsDeployed            json.RawMessage            `json:"is_deployed"`
	CreatedAt             *string                    `json:"created_at"`
	UpdatedAt             *string                    `json:"updated_at"`
	Network               *int                       `json:"network"`
	Emoji                 *string                    `json:"emoji"`
	IconProp              json.RawMessage            `json:"icon_prop"`
	ModuleView            *bool                      `json:"module_view"`
	CycleView             *bool                      `json:"cycle_view"`
	InboxView             *bool                      `json:"inbox_view"`
	PageView              *bool                      `json:"page_view"`
	IssueViewsView        *bool                      `json:"issue_views_view"`
	CoverImage            *string                    `json:"cover_image"`
	ArchiveIn             *int                       `json:"archive_in"`
	CloseIn               *int                       `json:"close_in"`
	CreatedBy             *string                    `json:"created_by"`
	UpdatedBy             *string                    `json:"updated_by"`
	Workspace             *string                    `json:"workspace"`
	DefaultAssignee       *string                    `json:"default_assignee"`
	ProjectLead           *string                    `json:"project_lead"`
	Estimate              *string                    `json:"estimate"`
	DefaultState          json.RawMessage            `json:"default_state"`
	TemplateID            *string                    `json:"template_id"`
	GuestViewAllFeatures  *bool                      `json:"guest_view_all_features"`
	ExternalSource        *string                    `json:"external_source"`
	ExternalID            *string                    `json:"external_id"`
	IsIssueTypeEnabled    *bool                      `json:"is_issue_type_enabled"`
	IsTimeTrackingEnabled *bool                      `json:"is_time_tracking_enabled"`
	Unknown               map[string]json.RawMessage `json:"-"`
	present               map[string]bool            `json:"-"`
}

var projectKnownFields = []string{
	"id", "name", "identifier", "description", "description_text", "description_html",
	"total_members", "total_cycles", "total_modules", "is_member", "member_role",
	"is_deployed", "created_at", "updated_at", "network", "emoji", "icon_prop",
	"module_view", "cycle_view", "inbox_view", "page_view", "issue_views_view",
	"cover_image", "archive_in", "close_in", "created_by", "updated_by", "workspace",
	"default_assignee", "project_lead", "estimate", "default_state", "template_id",
	"guest_view_all_features", "external_source", "external_id", "is_issue_type_enabled",
	"is_time_tracking_enabled",
}

// UnmarshalJSON decodes known fields while retaining every unmodeled field.
func (p *Project) UnmarshalJSON(data []byte) error {
	type projectAlias Project
	var decoded projectAlias
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	var allFields map[string]json.RawMessage
	if err := json.Unmarshal(data, &allFields); err != nil {
		return err
	}
	raw := make(map[string]json.RawMessage, len(allFields))
	for key, value := range allFields {
		raw[key] = value
	}
	for _, key := range projectKnownFields {
		delete(raw, key)
	}
	decoded.Unknown = raw
	decoded.present = make(map[string]bool, len(projectKnownFields))
	for key := range allFields {
		if containsString(projectKnownFields, key) {
			decoded.present[key] = true
		}
	}
	*p = Project(decoded)
	return nil
}

// MarshalJSON emits typed fields and merges unknown fields retained by
// UnmarshalJSON. Explicit nulls in typed pointer/raw fields are preserved.
func (p Project) MarshalJSON() ([]byte, error) {
	type projectAlias Project
	data, err := json.Marshal(projectAlias(p))
	if err != nil {
		return nil, err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return nil, err
	}
	if p.present != nil {
		presentFields := make(map[string]json.RawMessage, len(fields))
		for key := range p.present {
			presentFields[key] = fields[key]
		}
		for key, value := range fields {
			if _, present := p.present[key]; !present && string(value) != "null" {
				presentFields[key] = value
			}
		}
		fields = presentFields
	}
	for key, value := range p.Unknown {
		if _, exists := fields[key]; !exists {
			fields[key] = append(json.RawMessage(nil), value...)
		}
	}
	return json.Marshal(fields)
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

// ProjectPage is the cursor page returned by list projects. The embedded
// shared cursor envelope retains the documented cursor/count fields; the
// Project-specific grouping/count keys and unknown keys are retained too.
type ProjectPage struct {
	plane.CursorPage[Project]
	GroupedBy    *string `json:"grouped_by"`
	SubGroupedBy *string `json:"sub_grouped_by"`
	TotalCount   *int    `json:"total_count"`
}

var projectPageKnownFields = []string{
	"next_cursor", "prev_cursor", "next_page_results", "prev_page_results", "count",
	"total_pages", "total_results", "extra_stats", "results", "grouped_by",
	"sub_grouped_by", "total_count",
}

// UnmarshalJSON decodes the shared cursor envelope and retains the additional
// grouping/count metadata and any future page keys.
func (p *ProjectPage) UnmarshalJSON(data []byte) error {
	var cursor plane.CursorPage[Project]
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
	for _, key := range projectPageKnownFields {
		delete(unknown, key)
	}
	cursor.Unknown = unknown
	*p = ProjectPage{
		CursorPage:   cursor,
		GroupedBy:    grouped.GroupedBy,
		SubGroupedBy: grouped.SubGroupedBy,
		TotalCount:   grouped.TotalCount,
	}
	return nil
}

// MarshalJSON emits the shared cursor envelope, project grouping metadata,
// and all unknown page keys retained by UnmarshalJSON.
func (p ProjectPage) MarshalJSON() ([]byte, error) {
	type cursorAlias plane.CursorPage[Project]
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

// CreateProjectRequest contains the normal-create body. Name and Identifier
// are required; pointer optionals preserve explicit false and zero values.
type CreateProjectRequest struct {
	Name                  string           `json:"name"`
	Description           *string          `json:"description,omitempty"`
	ProjectLead           *string          `json:"project_lead,omitempty"`
	DefaultAssignee       *string          `json:"default_assignee,omitempty"`
	Identifier            string           `json:"identifier"`
	IconProp              *json.RawMessage `json:"icon_prop,omitempty"`
	Emoji                 *string          `json:"emoji,omitempty"`
	CoverImage            *string          `json:"cover_image,omitempty"`
	ModuleView            *bool            `json:"module_view,omitempty"`
	CycleView             *bool            `json:"cycle_view,omitempty"`
	IssueViewsView        *bool            `json:"issue_views_view,omitempty"`
	PageView              *bool            `json:"page_view,omitempty"`
	IntakeView            *bool            `json:"intake_view,omitempty"`
	GuestViewAllFeatures  *bool            `json:"guest_view_all_features,omitempty"`
	ArchiveIn             *int             `json:"archive_in,omitempty"`
	CloseIn               *int             `json:"close_in,omitempty"`
	Timezone              *string          `json:"timezone,omitempty"`
	ExternalSource        *string          `json:"external_source,omitempty"`
	ExternalID            *string          `json:"external_id,omitempty"`
	IsIssueTypeEnabled    *bool            `json:"is_issue_type_enabled,omitempty"`
	IsTimeTrackingEnabled *bool            `json:"is_time_tracking_enabled,omitempty"`
}

func (r CreateProjectRequest) validate() error {
	if strings.TrimSpace(r.Name) == "" {
		return fmt.Errorf("project name is required")
	}
	if strings.TrimSpace(r.Identifier) == "" {
		return fmt.Errorf("project identifier is required")
	}
	return nil
}

// CreateProjectFromTemplateRequest contains the template-create body.
type CreateProjectFromTemplateRequest struct {
	TemplateID  string  `json:"template_id"`
	Name        *string `json:"name,omitempty"`
	Identifier  *string `json:"identifier,omitempty"`
	Description *string `json:"description,omitempty"`
	Network     *int    `json:"network,omitempty"`
	ProjectLead *string `json:"project_lead,omitempty"`
}

func (r CreateProjectFromTemplateRequest) validate() error {
	if strings.TrimSpace(r.TemplateID) == "" {
		return fmt.Errorf("project template_id is required")
	}
	return nil
}

// UpdateProjectRequest contains only the documented optional PATCH fields.
// Pointers are required so explicit false, zero, empty string, and JSON icon
// values are not confused with omission.
type UpdateProjectRequest struct {
	Name                  *string          `json:"name,omitempty"`
	Description           *string          `json:"description,omitempty"`
	ProjectLead           *string          `json:"project_lead,omitempty"`
	DefaultAssignee       *string          `json:"default_assignee,omitempty"`
	Identifier            *string          `json:"identifier,omitempty"`
	IconProp              *json.RawMessage `json:"icon_prop,omitempty"`
	Emoji                 *string          `json:"emoji,omitempty"`
	CoverImage            *string          `json:"cover_image,omitempty"`
	ModuleView            *bool            `json:"module_view,omitempty"`
	CycleView             *bool            `json:"cycle_view,omitempty"`
	IssueViewsView        *bool            `json:"issue_views_view,omitempty"`
	PageView              *bool            `json:"page_view,omitempty"`
	IntakeView            *bool            `json:"intake_view,omitempty"`
	GuestViewAllFeatures  *bool            `json:"guest_view_all_features,omitempty"`
	ArchiveIn             *int             `json:"archive_in,omitempty"`
	CloseIn               *int             `json:"close_in,omitempty"`
	Timezone              *string          `json:"timezone,omitempty"`
	ExternalSource        *string          `json:"external_source,omitempty"`
	ExternalID            *string          `json:"external_id,omitempty"`
	IsIssueTypeEnabled    *bool            `json:"is_issue_type_enabled,omitempty"`
	IsTimeTrackingEnabled *bool            `json:"is_time_tracking_enabled,omitempty"`
	DefaultState          *string          `json:"default_state,omitempty"`
	Estimate              *string          `json:"estimate,omitempty"`
}

// ListOptions are the documented list-project query parameters. A zero
// PerPage omits the parameter and lets Plane use its default of 20.
type ListOptions struct {
	Cursor  string
	PerPage int
	Fields  string
	Expand  string
	OrderBy string
}

// Query returns a copied, validated query for the list operation. The shared
// foundation enforces the documented 1–100 per_page range and preserves the
// documented pass-through -field order syntax.
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
