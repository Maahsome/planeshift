package links

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"planeshift/plane"
)

// The link operation contract is intentionally recorded beside the adapter:
// primary routes use /work-items/, the five listed compatibility routes use
// /issues/, create is 201, list/detail/update are 200, delete is 204, and the
// operation scopes are projects.work_items.links:{read,write}. The shared
// client owns auth, JSON, bounded errors, and response metadata.
type Client struct {
	client plane.Client
}

func NewClient(client plane.Client) *Client { return &Client{client: client} }

func New(client plane.Client) *Client { return NewClient(client) }

func (c *Client) do(ctx context.Context, method, route string, query url.Values, body, destination any, expectedStatus int) (plane.Response, error) {
	if c == nil || c.client == nil {
		return plane.Response{}, fmt.Errorf("links client is not initialized")
	}
	response, err := c.client.Do(ctx, method, route, query, body, nil, destination)
	if err != nil {
		return response, err
	}
	if response.StatusCode != expectedStatus {
		return response, fmt.Errorf("links %s returned HTTP status %d, want %d", method, response.StatusCode, expectedStatus)
	}
	return response, nil
}

func oneDetailOption(options []DetailOptions) DetailOptions {
	if len(options) == 0 {
		return DetailOptions{}
	}
	return options[0]
}

// List returns the current work-item link page.
func (c *Client) List(ctx context.Context, workspaceSlug, projectID, workItemID string, options ListOptions) (LinkPage, plane.Response, error) {
	query, err := options.Query()
	if err != nil {
		return LinkPage{}, plane.Response{}, err
	}
	var page LinkPage
	response, err := c.do(ctx, http.MethodGet, currentCollectionPath(workspaceSlug, projectID, workItemID), query, nil, &page, http.StatusOK)
	return page, response, err
}

// Create adds a link to the current work item.
func (c *Client) Create(ctx context.Context, workspaceSlug, projectID, workItemID string, request CreateLinkRequest) (Link, plane.Response, error) {
	if err := request.validate(); err != nil {
		return Link{}, plane.Response{}, err
	}
	var link Link
	response, err := c.do(ctx, http.MethodPost, currentCollectionPath(workspaceSlug, projectID, workItemID), nil, request, &link, http.StatusCreated)
	return link, response, err
}

// Get retrieves a link using the documented page-shaped detail envelope.
func (c *Client) Get(ctx context.Context, workspaceSlug, projectID, workItemID, linkID string, options ...DetailOptions) (LinkDetailResponse, plane.Response, error) {
	query, err := oneDetailOption(options).Query()
	if err != nil {
		return LinkDetailResponse{}, plane.Response{}, err
	}
	var result LinkDetailResponse
	response, err := c.do(ctx, http.MethodGet, currentDetailPath(workspaceSlug, projectID, workItemID, linkID), query, nil, &result, http.StatusOK)
	return result, response, err
}

// Update partially updates a current link.
func (c *Client) Update(ctx context.Context, workspaceSlug, projectID, workItemID, linkID string, request UpdateLinkRequest) (Link, plane.Response, error) {
	var link Link
	response, err := c.do(ctx, http.MethodPatch, currentDetailPath(workspaceSlug, projectID, workItemID, linkID), nil, request, &link, http.StatusOK)
	return link, response, err
}

// Delete removes a current link. Plane returns 204 with no body.
func (c *Client) Delete(ctx context.Context, workspaceSlug, projectID, workItemID, linkID string) (plane.Response, error) {
	return c.do(ctx, http.MethodDelete, currentDetailPath(workspaceSlug, projectID, workItemID, linkID), nil, nil, nil, http.StatusNoContent)
}

// The following five methods are explicitly inventoried /issues/ compatibility
// operations. They are never selected by the primary command runners.
func (c *Client) LegacyList(ctx context.Context, workspaceSlug, projectID, issueID string, options ListOptions) (LinkPage, plane.Response, error) {
	query, err := options.Query()
	if err != nil {
		return LinkPage{}, plane.Response{}, err
	}
	var page LinkPage
	response, err := c.do(ctx, http.MethodGet, legacyCollectionPath(workspaceSlug, projectID, issueID), query, nil, &page, http.StatusOK)
	return page, response, err
}

func (c *Client) LegacyCreate(ctx context.Context, workspaceSlug, projectID, issueID string, request CreateLinkRequest) (Link, plane.Response, error) {
	if err := request.validate(); err != nil {
		return Link{}, plane.Response{}, err
	}
	var link Link
	response, err := c.do(ctx, http.MethodPost, legacyCollectionPath(workspaceSlug, projectID, issueID), nil, request, &link, http.StatusCreated)
	return link, response, err
}

func (c *Client) LegacyGet(ctx context.Context, workspaceSlug, projectID, issueID, linkID string, options ...DetailOptions) (LinkDetailResponse, plane.Response, error) {
	query, err := oneDetailOption(options).Query()
	if err != nil {
		return LinkDetailResponse{}, plane.Response{}, err
	}
	var result LinkDetailResponse
	response, err := c.do(ctx, http.MethodGet, legacyDetailPath(workspaceSlug, projectID, issueID, linkID), query, nil, &result, http.StatusOK)
	return result, response, err
}

func (c *Client) LegacyUpdate(ctx context.Context, workspaceSlug, projectID, issueID, linkID string, request UpdateLinkRequest) (Link, plane.Response, error) {
	var link Link
	response, err := c.do(ctx, http.MethodPatch, legacyDetailPath(workspaceSlug, projectID, issueID, linkID), nil, request, &link, http.StatusOK)
	return link, response, err
}

func (c *Client) LegacyDelete(ctx context.Context, workspaceSlug, projectID, issueID, linkID string) (plane.Response, error) {
	return c.do(ctx, http.MethodDelete, legacyDetailPath(workspaceSlug, projectID, issueID, linkID), nil, nil, nil, http.StatusNoContent)
}

func currentCollectionPath(workspaceSlug, projectID, workItemID string) string {
	return "/workspaces/" + plane.EscapePathSegment(workspaceSlug) + "/projects/" + plane.EscapePathSegment(projectID) + "/work-items/" + plane.EscapePathSegment(workItemID) + "/links/"
}

func currentDetailPath(workspaceSlug, projectID, workItemID, linkID string) string {
	return currentCollectionPath(workspaceSlug, projectID, workItemID) + plane.EscapePathSegment(linkID) + "/"
}

func legacyCollectionPath(workspaceSlug, projectID, issueID string) string {
	return "/workspaces/" + plane.EscapePathSegment(workspaceSlug) + "/projects/" + plane.EscapePathSegment(projectID) + "/issues/" + plane.EscapePathSegment(issueID) + "/links/"
}

func legacyDetailPath(workspaceSlug, projectID, issueID, linkID string) string {
	return legacyCollectionPath(workspaceSlug, projectID, issueID) + plane.EscapePathSegment(linkID) + "/"
}
