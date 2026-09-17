package states

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"planeshift/plane"
)

// The official Work Item States operation matrix is kept beside the client so
// documentation naming cannot alter the retained wire contract.
//
//	List:   GET    /api/v1/workspaces/{slug}/projects/{project_id}/states/                       200 JSON, projects.states:read
//	Create: POST   /api/v1/workspaces/{slug}/projects/{project_id}/states/                       200 JSON, projects.states:write
//	Get:    GET    /api/v1/workspaces/{slug}/projects/{project_id}/states/{state_id}/            200 JSON, projects.states:read
//	Update: PATCH  /api/v1/workspaces/{slug}/projects/{project_id}/states/{state_id}/            200 JSON, projects.states:write
//	Delete: DELETE /api/v1/workspaces/{slug}/projects/{project_id}/states/{state_id}/            204, projects.states:write
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
		return plane.Response{}, fmt.Errorf("states client is not initialized")
	}
	return c.client.Do(ctx, method, route, query, body, nil, destination)
}

func collectionPath(workspaceSlug, projectID string) string {
	return "/workspaces/" + plane.EscapePathSegment(workspaceSlug) + "/projects/" + plane.EscapePathSegment(projectID) + "/states/"
}

func detailPath(workspaceSlug, projectID, stateID string) string {
	return collectionPath(workspaceSlug, projectID) + plane.EscapePathSegment(stateID) + "/"
}

// Create creates a state. OAuth scope: projects.states:write.
func (c *Client) Create(ctx context.Context, workspaceSlug, projectID string, request CreateStateRequest) (State, plane.Response, error) {
	if err := request.validate(); err != nil {
		return State{}, plane.Response{}, err
	}
	var state State
	response, err := c.do(ctx, http.MethodPost, collectionPath(workspaceSlug, projectID), nil, request, &state)
	return state, response, err
}

// List returns project states. OAuth scope: projects.states:read.
func (c *Client) List(ctx context.Context, workspaceSlug, projectID string, options ListOptions) (StatePage, plane.Response, error) {
	query, err := options.Query()
	if err != nil {
		return StatePage{}, plane.Response{}, err
	}
	var page StatePage
	response, err := c.do(ctx, http.MethodGet, collectionPath(workspaceSlug, projectID), query, nil, &page)
	return page, response, err
}

// Get retrieves one state. OAuth scope: projects.states:read. The operation
// has no documented query parameters.
func (c *Client) Get(ctx context.Context, workspaceSlug, projectID, stateID string) (State, plane.Response, error) {
	var state State
	response, err := c.do(ctx, http.MethodGet, detailPath(workspaceSlug, projectID, stateID), nil, nil, &state)
	return state, response, err
}

// Update partially updates a state. OAuth scope: projects.states:write.
func (c *Client) Update(ctx context.Context, workspaceSlug, projectID, stateID string, request UpdateStateRequest) (State, plane.Response, error) {
	if err := request.validate(); err != nil {
		return State{}, plane.Response{}, err
	}
	var state State
	response, err := c.do(ctx, http.MethodPatch, detailPath(workspaceSlug, projectID, stateID), nil, request, &state)
	return state, response, err
}

// Delete removes a state. OAuth scope: projects.states:write. Successful
// deletion returns 204 with no response body.
func (c *Client) Delete(ctx context.Context, workspaceSlug, projectID, stateID string) (plane.Response, error) {
	return c.do(ctx, http.MethodDelete, detailPath(workspaceSlug, projectID, stateID), nil, nil, nil)
}
