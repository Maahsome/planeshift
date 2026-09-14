package projectfeatures

import (
	"context"
	"fmt"
	"net/http"

	"planeshift/plane"
)

// Operation matrix from the official Project Features pages:
//
//	Get:    GET   /api/v1/workspaces/{workspace_slug}/projects/{project_id}/features/  200 JSON, projects.features:read
//	Update: PATCH /api/v1/workspaces/{workspace_slug}/projects/{project_id}/features/  200 JSON, projects.features:write
//
// Both operations have the same required path parameters and no query
// parameters. Update accepts any subset of the seven optional boolean fields.
// Shared authentication, JSON headers, status/error handling, 204 behavior,
// response closure, rate-limit metadata, and context behavior remain owned by
// plane.Client. The official source pages are:
//   - https://developers.plane.so/api-reference/project-features/get-project-features
//   - https://developers.plane.so/api-reference/project-features/update-project-features
type Client struct {
	client plane.Client
}

// NewClient wraps an injected shared Plane client without constructing a
// transport, reading configuration, or performing a request.
func NewClient(client plane.Client) *Client {
	return &Client{client: client}
}

func (c *Client) do(ctx context.Context, method, route string, body any, destination any) (plane.Response, error) {
	if c == nil || c.client == nil {
		return plane.Response{}, fmt.Errorf("project features client is not initialized")
	}
	return c.client.Do(ctx, method, route, nil, body, nil, destination)
}

func featuresPath(workspaceSlug, projectID string) string {
	return "/workspaces/" + plane.EscapePathSegment(workspaceSlug) +
		"/projects/" + plane.EscapePathSegment(projectID) + "/features/"
}

// Get retrieves a project's feature flags. OAuth scope: projects.features:read.
func (c *Client) Get(ctx context.Context, workspaceSlug, projectID string) (ProjectFeatures, plane.Response, error) {
	var features ProjectFeatures
	response, err := c.do(ctx, http.MethodGet, featuresPath(workspaceSlug, projectID), nil, &features)
	return features, response, err
}

// Update partially updates a project's feature flags. OAuth scope:
// projects.features:write. An empty request is sent as {} when no fields are
// selected, preserving the documented partial-update shape without inventing
// a client-side requirement for at least one flag.
func (c *Client) Update(ctx context.Context, workspaceSlug, projectID string, request UpdateProjectFeaturesRequest) (ProjectFeatures, plane.Response, error) {
	var features ProjectFeatures
	response, err := c.do(ctx, http.MethodPatch, featuresPath(workspaceSlug, projectID), request, &features)
	return features, response, err
}
