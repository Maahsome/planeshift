package projectlabels

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"planeshift/plane"
)

// The official Project Labels operation matrix is kept beside the routes so
// documentation directory names cannot change the wire contract.
//
//	Create: POST   /api/v1/workspaces/{workspace_slug}/project-labels/              201 JSON, projects.labels:write
//	List:   GET    /api/v1/workspaces/{workspace_slug}/project-labels/              200 page, projects.labels:read
//	Get:    GET    /api/v1/workspaces/{workspace_slug}/project-labels/{label_id}/    200 JSON, projects.labels:read
//	Update: PATCH  /api/v1/workspaces/{workspace_slug}/project-labels/{label_id}/    200 JSON, projects.labels:write
//	Delete: DELETE /api/v1/workspaces/{workspace_slug}/project-labels/{label_id}/    204, projects.labels:write
//
// The resource package owns only these route-specific paths and contracts.
// Authentication, request construction, pagination, safe errors, and response
// metadata remain in the shared plane.Client.
type Client struct {
	client plane.Client
}

// NewClient wraps an injected shared Plane client and performs no request.
func NewClient(client plane.Client) *Client {
	return &Client{client: client}
}

// New is a concise constructor alias for callers composing resource clients.
func New(client plane.Client) *Client {
	return NewClient(client)
}

func (c *Client) do(ctx context.Context, method, route string, query url.Values, body any, destination any) (plane.Response, error) {
	if c == nil || c.client == nil {
		return plane.Response{}, fmt.Errorf("project labels client is not initialized")
	}
	return c.client.Do(ctx, method, route, query, body, nil, destination)
}

func projectLabelsPath(workspaceSlug string) string {
	return "/workspaces/" + plane.EscapePathSegment(workspaceSlug) + "/project-labels/"
}

func projectLabelPath(workspaceSlug, labelID string) string {
	return projectLabelsPath(workspaceSlug) + plane.EscapePathSegment(labelID) + "/"
}

// Create adds a project label to a workspace. OAuth scope:
// projects.labels:write.
func (c *Client) Create(ctx context.Context, workspaceSlug string, request CreateProjectLabelRequest) (ProjectLabel, plane.Response, error) {
	if err := request.validate(); err != nil {
		return ProjectLabel{}, plane.Response{}, err
	}
	var label ProjectLabel
	response, err := c.do(ctx, http.MethodPost, projectLabelsPath(workspaceSlug), nil, request, &label)
	return label, response, err
}

// List returns the workspace project-label page. OAuth scope:
// projects.labels:read.
func (c *Client) List(ctx context.Context, workspaceSlug string, options ListOptions) (ProjectLabelPage, plane.Response, error) {
	query, err := options.Query()
	if err != nil {
		return ProjectLabelPage{}, plane.Response{}, err
	}
	var page ProjectLabelPage
	response, err := c.do(ctx, http.MethodGet, projectLabelsPath(workspaceSlug), query, nil, &page)
	return page, response, err
}

// Get retrieves one workspace project label. OAuth scope:
// projects.labels:read. The operation has no documented query parameters.
func (c *Client) Get(ctx context.Context, workspaceSlug, labelID string) (ProjectLabel, plane.Response, error) {
	var label ProjectLabel
	response, err := c.do(ctx, http.MethodGet, projectLabelPath(workspaceSlug, labelID), nil, nil, &label)
	return label, response, err
}

// Update partially updates one workspace project label. OAuth scope:
// projects.labels:write.
func (c *Client) Update(ctx context.Context, workspaceSlug, labelID string, request UpdateProjectLabelRequest) (ProjectLabel, plane.Response, error) {
	var label ProjectLabel
	response, err := c.do(ctx, http.MethodPatch, projectLabelPath(workspaceSlug, labelID), nil, request, &label)
	return label, response, err
}

// Delete removes one workspace project label. OAuth scope:
// projects.labels:write. The response carries status and safe metadata; the
// documented success status is 204 with no response body.
func (c *Client) Delete(ctx context.Context, workspaceSlug, labelID string) (plane.Response, error) {
	return c.do(ctx, http.MethodDelete, projectLabelPath(workspaceSlug, labelID), nil, nil, nil)
}

// ListOptions contains the documented project-label list query parameters. A
// zero PerPage omits the value and lets Plane use its default of 20.
type ListOptions struct {
	Cursor  string
	PerPage int
	Fields  string
	Expand  string
	OrderBy string
}

// Query validates and builds list query values through the shared pagination
// helper, retaining the resource-specific order_by parameter.
func (o ListOptions) Query() (url.Values, error) {
	query := make(url.Values)
	if o.OrderBy != "" {
		query.Set("order_by", o.OrderBy)
	}
	return plane.WithPagination(query, plane.PaginationOptions{
		Cursor: o.Cursor, PerPage: o.PerPage, Fields: o.Fields, Expand: o.Expand,
	})
}
