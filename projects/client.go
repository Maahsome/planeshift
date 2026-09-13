package projects

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"planeshift/plane"
)

// The official Projects operation matrix is intentionally kept next to the
// route implementation. The docs use resource_id for retrieve/update/delete
// and project_id for archive/unarchive; the Go API consistently calls that
// argument projectID while preserving each wire path.
//
//	Create:          POST   /api/v1/workspaces/{workspace_slug}/projects/                  201 JSON, projects:write
//	Create template: POST   /api/v1/workspaces/{workspace_slug}/projects/templates/use/    201 JSON, write or projects:write
//	List:            GET    /api/v1/workspaces/{workspace_slug}/projects/                  200 page, projects:read
//	Retrieve:        GET    /api/v1/workspaces/{workspace_slug}/projects/{resource_id}/     200 JSON, projects:read
//	Update:          PATCH  /api/v1/workspaces/{workspace_slug}/projects/{resource_id}/     200 JSON, projects:write
//	Archive:         POST   /api/v1/workspaces/{workspace_slug}/projects/{project_id}/archive/ 204, projects:write
//	Unarchive:       DELETE /api/v1/workspaces/{workspace_slug}/projects/{project_id}/archive/ 204, projects:write
//	Delete:          DELETE /api/v1/workspaces/{workspace_slug}/projects/{resource_id}/     204, projects:write
//
// The resource-package boundary remains the uncharted, pending-review
// decision recorded in PLANE_CLIENT.md. This package therefore owns only
// route-specific models and paths; authentication, configuration, output
// dispatch, pagination transport, and errors remain in their existing owners.
type Client struct {
	client plane.Client
}

// NewClient wraps an injected shared Plane client. It never constructs a
// transport, reads configuration, or performs a request.
func NewClient(client plane.Client) *Client {
	return &Client{client: client}
}

// New is a concise constructor alias for callers composing resource clients.
func New(client plane.Client) *Client {
	return NewClient(client)
}

func (c *Client) do(ctx context.Context, method, route string, query url.Values, body any, destination any) (plane.Response, error) {
	if c == nil || c.client == nil {
		return plane.Response{}, fmt.Errorf("projects client is not initialized")
	}
	return c.client.Do(ctx, method, route, query, body, nil, destination)
}

func workspaceProjectsPath(workspaceSlug string) string {
	return "/workspaces/" + plane.EscapePathSegment(workspaceSlug) + "/projects/"
}

func projectPath(workspaceSlug, projectID string) string {
	return workspaceProjectsPath(workspaceSlug) + plane.EscapePathSegment(projectID) + "/"
}

func archivePath(workspaceSlug, projectID string) string {
	return projectPath(workspaceSlug, projectID) + "archive/"
}

// Create creates a project. OAuth scope: projects:write. The request is
// validated locally before the shared client is invoked.
func (c *Client) Create(ctx context.Context, workspaceSlug string, request CreateProjectRequest) (Project, plane.Response, error) {
	if err := request.validate(); err != nil {
		return Project{}, plane.Response{}, err
	}
	var project Project
	response, err := c.do(ctx, http.MethodPost, workspaceProjectsPath(workspaceSlug), nil, request, &project)
	return project, response, err
}

// CreateFromTemplate creates a project from a template. OAuth scope: write or
// projects:write as documented by Plane.
func (c *Client) CreateFromTemplate(ctx context.Context, workspaceSlug string, request CreateProjectFromTemplateRequest) (Project, plane.Response, error) {
	if err := request.validate(); err != nil {
		return Project{}, plane.Response{}, err
	}
	var project Project
	response, err := c.do(ctx, http.MethodPost, workspaceProjectsPath(workspaceSlug)+"templates/use/", nil, request, &project)
	return project, response, err
}

// List returns the cursor page of projects. OAuth scope: projects:read.
func (c *Client) List(ctx context.Context, workspaceSlug string, options ListOptions) (ProjectPage, plane.Response, error) {
	query, err := options.Query()
	if err != nil {
		return ProjectPage{}, plane.Response{}, err
	}
	var page ProjectPage
	response, err := c.clientDo(ctx, http.MethodGet, workspaceProjectsPath(workspaceSlug), query, nil, &page)
	return page, response, err
}

// Get retrieves a project using the official resource_id wire path. OAuth
// scope: projects:read. The operation has no documented query parameters.
func (c *Client) Get(ctx context.Context, workspaceSlug, projectID string) (Project, plane.Response, error) {
	var project Project
	response, err := c.clientDo(ctx, http.MethodGet, projectPath(workspaceSlug, projectID), nil, nil, &project)
	return project, response, err
}

// Update partially updates a project using the official resource_id wire
// path. OAuth scope: projects:write.
func (c *Client) Update(ctx context.Context, workspaceSlug, projectID string, request UpdateProjectRequest) (Project, plane.Response, error) {
	var project Project
	response, err := c.clientDo(ctx, http.MethodPatch, projectPath(workspaceSlug, projectID), nil, request, &project)
	return project, response, err
}

// Archive archives a project. It sends no request body and expects 204. OAuth
// scope: projects:write.
func (c *Client) Archive(ctx context.Context, workspaceSlug, projectID string) (plane.Response, error) {
	return c.clientDo(ctx, http.MethodPost, archivePath(workspaceSlug, projectID), nil, nil, nil)
}

// Unarchive restores a project. It sends no request body and expects 204.
// OAuth scope: projects:write.
func (c *Client) Unarchive(ctx context.Context, workspaceSlug, projectID string) (plane.Response, error) {
	return c.clientDo(ctx, http.MethodDelete, archivePath(workspaceSlug, projectID), nil, nil, nil)
}

// Delete permanently deletes a project using the official resource_id wire
// path. It sends no request body and expects 204. OAuth scope: projects:write.
func (c *Client) Delete(ctx context.Context, workspaceSlug, projectID string) (plane.Response, error) {
	return c.clientDo(ctx, http.MethodDelete, projectPath(workspaceSlug, projectID), nil, nil, nil)
}

// clientDo is the typed version of the shared client call used by every
// operation. Only the shared plane.Client receives the request.
func (c *Client) clientDo(ctx context.Context, method, route string, query url.Values, body any, destination any) (plane.Response, error) {
	if c == nil || c.client == nil {
		return plane.Response{}, fmt.Errorf("projects client is not initialized")
	}
	return c.client.Do(ctx, method, route, query, body, nil, destination)
}
