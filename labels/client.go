package labels

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"planeshift/plane"
)

// The official project-scoped label operation matrix is kept beside the
// client so documentation naming cannot alter the retained wire contract.
//
//	List:   GET    /api/v1/workspaces/{slug}/projects/{project_id}/labels/                    200 JSON, projects.labels:read
//	Create: POST   /api/v1/workspaces/{slug}/projects/{project_id}/labels/                    201 JSON, projects.labels:write
//	Get:    GET    /api/v1/workspaces/{slug}/projects/{project_id}/labels/{label_id}/         200 JSON, projects.labels:read
//	Update: PATCH  /api/v1/workspaces/{slug}/projects/{project_id}/labels/{label_id}/         200 JSON, projects.labels:write
//	Delete: DELETE /api/v1/workspaces/{slug}/projects/{project_id}/labels/{label_id}/         204, projects.labels:write
//
// Client owns only route-specific paths and models. Authentication, request
// construction, output dispatch, pagination transport, and errors remain in
// the shared plane package.
type Client struct {
	client plane.Client
}

// NewClient wraps an injected shared Plane client without constructing a
// transport, reading configuration, or making a request.
func NewClient(client plane.Client) *Client {
	return &Client{client: client}
}

// New is a concise constructor alias for callers composing resource clients.
func New(client plane.Client) *Client {
	return NewClient(client)
}

func (c *Client) do(ctx context.Context, method, route string, query url.Values, body any, destination any) (plane.Response, error) {
	if c == nil || c.client == nil {
		return plane.Response{}, fmt.Errorf("labels client is not initialized")
	}
	return c.client.Do(ctx, method, route, query, body, nil, destination)
}

func collectionPath(workspaceSlug, projectID string) string {
	return "/workspaces/" + plane.EscapePathSegment(workspaceSlug) + "/projects/" + plane.EscapePathSegment(projectID) + "/labels/"
}

func detailPath(workspaceSlug, projectID, labelID string) string {
	return collectionPath(workspaceSlug, projectID) + plane.EscapePathSegment(labelID) + "/"
}

// Create creates a project-scoped label. OAuth scope: projects.labels:write.
func (c *Client) Create(ctx context.Context, workspaceSlug, projectID string, request CreateLabelRequest) (Label, plane.Response, error) {
	if err := request.validate(); err != nil {
		return Label{}, plane.Response{}, err
	}
	var label Label
	response, err := c.do(ctx, http.MethodPost, collectionPath(workspaceSlug, projectID), nil, request, &label)
	return label, response, err
}

// List returns project-scoped labels. OAuth scope: projects.labels:read.
func (c *Client) List(ctx context.Context, workspaceSlug, projectID string, options ListOptions) (LabelPage, plane.Response, error) {
	query, err := options.Query()
	if err != nil {
		return LabelPage{}, plane.Response{}, err
	}
	var page LabelPage
	response, err := c.do(ctx, http.MethodGet, collectionPath(workspaceSlug, projectID), query, nil, &page)
	return page, response, err
}

// Get retrieves one project-scoped label. OAuth scope: projects.labels:read.
// The operation has no documented query parameters.
func (c *Client) Get(ctx context.Context, workspaceSlug, projectID, labelID string) (Label, plane.Response, error) {
	var label Label
	response, err := c.do(ctx, http.MethodGet, detailPath(workspaceSlug, projectID, labelID), nil, nil, &label)
	return label, response, err
}

// Update partially updates a project-scoped label. OAuth scope:
// projects.labels:write.
func (c *Client) Update(ctx context.Context, workspaceSlug, projectID, labelID string, request UpdateLabelRequest) (Label, plane.Response, error) {
	var label Label
	response, err := c.do(ctx, http.MethodPatch, detailPath(workspaceSlug, projectID, labelID), nil, request, &label)
	return label, response, err
}

// Delete removes a project-scoped label. OAuth scope: projects.labels:write.
// Successful deletion returns 204 with no response body.
func (c *Client) Delete(ctx context.Context, workspaceSlug, projectID, labelID string) (plane.Response, error) {
	return c.do(ctx, http.MethodDelete, detailPath(workspaceSlug, projectID, labelID), nil, nil, nil)
}
