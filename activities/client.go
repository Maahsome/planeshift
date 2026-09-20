package activities

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"planeshift/plane"
)

// Client adapts the four read-only Work Item Activity operations. The
// operations use HTTP 200 and the projects.work_items.activities:read OAuth
// scope; the shared client owns authentication, decoding, safe errors, and
// response metadata.
type Client struct {
	client plane.Client
}

// NewClient wraps an injected shared Plane client.
func NewClient(client plane.Client) *Client { return &Client{client: client} }

// New is a concise constructor alias for callers composing resource clients.
func New(client plane.Client) *Client { return NewClient(client) }

func (c *Client) do(ctx context.Context, route string, query url.Values, destination any) (plane.Response, error) {
	if c == nil || c.client == nil {
		return plane.Response{}, fmt.Errorf("activities client is not initialized")
	}
	response, err := c.client.Do(ctx, http.MethodGet, route, query, nil, nil, destination)
	if err != nil {
		return response, err
	}
	if response.StatusCode != http.StatusOK {
		return response, fmt.Errorf("activities GET returned HTTP status %d, want %d", response.StatusCode, http.StatusOK)
	}
	return response, nil
}

// List returns the current work-item activity page.
func (c *Client) List(ctx context.Context, workspaceSlug, projectID, workItemID string, options ListOptions) (ActivityPage, plane.Response, error) {
	query, err := options.Query()
	if err != nil {
		return ActivityPage{}, plane.Response{}, err
	}
	var page ActivityPage
	response, err := c.do(ctx, currentCollectionPath(workspaceSlug, projectID, workItemID), query, &page)
	return page, response, err
}

// Get retrieves a current activity using the documented page-shaped detail
// envelope. The public route uses activity_id even though the operation page
// currently labels that path parameter resource_id.
func (c *Client) Get(ctx context.Context, workspaceSlug, projectID, workItemID, activityID string, options ...DetailOptions) (ActivityDetailResponse, plane.Response, error) {
	query, err := oneDetailOption(options).Query()
	if err != nil {
		return ActivityDetailResponse{}, plane.Response{}, err
	}
	var result ActivityDetailResponse
	response, err := c.do(ctx, currentDetailPath(workspaceSlug, projectID, workItemID, activityID), query, &result)
	return result, response, err
}

// LegacyList returns the explicitly inventoried deprecated /issues activity
// page. It is compatibility-only and never selected by primary runners.
func (c *Client) LegacyList(ctx context.Context, workspaceSlug, projectID, issueID string, options ListOptions) (ActivityPage, plane.Response, error) {
	query, err := options.Query()
	if err != nil {
		return ActivityPage{}, plane.Response{}, err
	}
	var page ActivityPage
	response, err := c.do(ctx, legacyCollectionPath(workspaceSlug, projectID, issueID), query, &page)
	return page, response, err
}

// LegacyGet returns the explicitly inventoried deprecated /issues activity
// detail response. It is compatibility-only and never selected by primary
// runners.
func (c *Client) LegacyGet(ctx context.Context, workspaceSlug, projectID, issueID, activityID string, options ...DetailOptions) (ActivityDetailResponse, plane.Response, error) {
	query, err := oneDetailOption(options).Query()
	if err != nil {
		return ActivityDetailResponse{}, plane.Response{}, err
	}
	var result ActivityDetailResponse
	response, err := c.do(ctx, legacyDetailPath(workspaceSlug, projectID, issueID, activityID), query, &result)
	return result, response, err
}

func oneDetailOption(options []DetailOptions) DetailOptions {
	if len(options) == 0 {
		return DetailOptions{}
	}
	return options[0]
}

func currentCollectionPath(workspaceSlug, projectID, workItemID string) string {
	return "/workspaces/" + plane.EscapePathSegment(workspaceSlug) + "/projects/" + plane.EscapePathSegment(projectID) + "/work-items/" + plane.EscapePathSegment(workItemID) + "/activities/"
}

func currentDetailPath(workspaceSlug, projectID, workItemID, activityID string) string {
	return currentCollectionPath(workspaceSlug, projectID, workItemID) + plane.EscapePathSegment(activityID) + "/"
}

func legacyCollectionPath(workspaceSlug, projectID, issueID string) string {
	return "/workspaces/" + plane.EscapePathSegment(workspaceSlug) + "/projects/" + plane.EscapePathSegment(projectID) + "/issues/" + plane.EscapePathSegment(issueID) + "/activities/"
}

func legacyDetailPath(workspaceSlug, projectID, issueID, activityID string) string {
	return legacyCollectionPath(workspaceSlug, projectID, issueID) + plane.EscapePathSegment(activityID) + "/"
}
