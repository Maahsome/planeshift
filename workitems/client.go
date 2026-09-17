package workitems

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"planeshift/plane"
)

// Client is a thin Work Item route adapter. The shared plane.Client performs
// URL joining, authentication, JSON headers, response decoding, status/error
// handling, metadata capture, context cancellation, and body closure.
type Client struct {
	client plane.Client
}

// NewClient wraps an injected shared Plane client.
func NewClient(client plane.Client) *Client { return &Client{client: client} }

// New is a concise constructor alias for callers composing resource clients.
func New(client plane.Client) *Client { return NewClient(client) }

func (c *Client) do(ctx context.Context, method, route string, query url.Values, body, destination any) (plane.Response, error) {
	if c == nil || c.client == nil {
		return plane.Response{}, fmt.Errorf("work items client is not initialized")
	}
	return c.client.Do(ctx, method, route, query, body, nil, destination)
}

func oneDetailOption(options []DetailOptions) DetailOptions {
	if len(options) == 0 {
		return DetailOptions{}
	}
	return options[0]
}

func oneIdentifierOption(options []IdentifierOptions) IdentifierOptions {
	if len(options) == 0 {
		return IdentifierOptions{}
	}
	return options[0]
}

// Search searches work items across a workspace. OAuth scope:
// projects.work_items:read.
func (c *Client) Search(ctx context.Context, workspaceSlug string, options SearchOptions) (WorkItemSearchResponse, plane.Response, error) {
	query, err := options.Query()
	if err != nil {
		return WorkItemSearchResponse{}, plane.Response{}, err
	}
	var result WorkItemSearchResponse
	response, err := c.do(ctx, http.MethodGet, currentSearchPath(workspaceSlug), query, nil, &result)
	return result, response, err
}

// GetByIdentifier resolves the documented PROJECT-123 identifier route. The
// two identifier segments are escaped independently before joining.
func (c *Client) GetByIdentifier(ctx context.Context, workspaceSlug, projectIdentifier, issueIdentifier string, options ...IdentifierOptions) (WorkItem, plane.Response, error) {
	var item WorkItem
	response, err := c.do(ctx, http.MethodGet, currentIdentifierPath(workspaceSlug, projectIdentifier, issueIdentifier), oneIdentifierOption(options).Query(), nil, &item)
	return item, response, err
}

// List returns project work items using cursor pagination. OAuth scope:
// projects.work_items:read.
func (c *Client) List(ctx context.Context, workspaceSlug, projectID string, options ListOptions) (WorkItemPage, plane.Response, error) {
	query, err := options.Query()
	if err != nil {
		return WorkItemPage{}, plane.Response{}, err
	}
	var page WorkItemPage
	response, err := c.do(ctx, http.MethodGet, currentCollectionPath(workspaceSlug, projectID), query, nil, &page)
	return page, response, err
}

// Create creates a work item. OAuth scope: projects.work_items:write.
func (c *Client) Create(ctx context.Context, workspaceSlug, projectID string, request CreateWorkItemRequest) (WorkItem, plane.Response, error) {
	if err := request.validate(); err != nil {
		return WorkItem{}, plane.Response{}, err
	}
	var item WorkItem
	response, err := c.do(ctx, http.MethodPost, currentCollectionPath(workspaceSlug, projectID), nil, request, &item)
	return item, response, err
}

// Get retrieves a current work item by UUID. OAuth scope:
// projects.work_items:read.
func (c *Client) Get(ctx context.Context, workspaceSlug, projectID, workItemID string, options ...DetailOptions) (WorkItem, plane.Response, error) {
	var item WorkItem
	response, err := c.do(ctx, http.MethodGet, currentDetailPath(workspaceSlug, projectID, workItemID), oneDetailOption(options).Query(), nil, &item)
	return item, response, err
}

// Update partially updates a current work item. OAuth scope:
// projects.work_items:write.
func (c *Client) Update(ctx context.Context, workspaceSlug, projectID, workItemID string, request UpdateWorkItemRequest, options ...DetailOptions) (WorkItem, plane.Response, error) {
	var item WorkItem
	response, err := c.do(ctx, http.MethodPatch, currentDetailPath(workspaceSlug, projectID, workItemID), oneDetailOption(options).Query(), request, &item)
	return item, response, err
}

// Delete deletes a current work item. OAuth scope: projects.work_items:write.
// Plane returns 204 with no body.
func (c *Client) Delete(ctx context.Context, workspaceSlug, projectID, workItemID string, options ...DetailOptions) (plane.Response, error) {
	return c.do(ctx, http.MethodDelete, currentDetailPath(workspaceSlug, projectID, workItemID), oneDetailOption(options).Query(), nil, nil)
}

// ListRelations lists relations for a current work item. OAuth scope:
// projects.work_items:read.
func (c *Client) ListRelations(ctx context.Context, workspaceSlug, projectID, workItemID string, options RelationListOptions) (WorkItemRelationPage, plane.Response, error) {
	query, err := options.Query()
	if err != nil {
		return WorkItemRelationPage{}, plane.Response{}, err
	}
	var page WorkItemRelationPage
	response, err := c.do(ctx, http.MethodGet, currentRelationsPath(workspaceSlug, projectID, workItemID), query, nil, &page)
	return page, response, err
}

// CreateRelation creates relations for a current work item. OAuth scope:
// projects.work_items:write. Plane returns a flat relation array with HTTP 201.
func (c *Client) CreateRelation(ctx context.Context, workspaceSlug, projectID, workItemID string, request CreateWorkItemRelationRequest) (WorkItemRelationCreateResponse, plane.Response, error) {
	if err := request.validate(); err != nil {
		return nil, plane.Response{}, err
	}
	var result WorkItemRelationCreateResponse
	response, err := c.do(ctx, http.MethodPost, currentRelationsPath(workspaceSlug, projectID, workItemID), nil, request, &result)
	return result, response, err
}

// The following seven methods are explicitly inventoried deprecated aliases.
// They intentionally use /issues/ routes and are never called by primary
// command runners.

func (c *Client) LegacySearch(ctx context.Context, workspaceSlug string, options SearchOptions) (WorkItemSearchResponse, plane.Response, error) {
	query, err := options.Query()
	if err != nil {
		return WorkItemSearchResponse{}, plane.Response{}, err
	}
	var result WorkItemSearchResponse
	response, err := c.do(ctx, http.MethodGet, legacySearchPath(workspaceSlug), query, nil, &result)
	return result, response, err
}

func (c *Client) LegacyGetByIdentifier(ctx context.Context, workspaceSlug, projectIdentifier, issueIdentifier string, options ...IdentifierOptions) (WorkItem, plane.Response, error) {
	var item WorkItem
	response, err := c.do(ctx, http.MethodGet, legacyIdentifierPath(workspaceSlug, projectIdentifier, issueIdentifier), oneIdentifierOption(options).Query(), nil, &item)
	return item, response, err
}

func (c *Client) LegacyList(ctx context.Context, workspaceSlug, projectID string, options ListOptions) (WorkItemPage, plane.Response, error) {
	query, err := options.Query()
	if err != nil {
		return WorkItemPage{}, plane.Response{}, err
	}
	var page WorkItemPage
	response, err := c.do(ctx, http.MethodGet, legacyCollectionPath(workspaceSlug, projectID), query, nil, &page)
	return page, response, err
}

func (c *Client) LegacyCreate(ctx context.Context, workspaceSlug, projectID string, request CreateWorkItemRequest) (WorkItem, plane.Response, error) {
	if err := request.validate(); err != nil {
		return WorkItem{}, plane.Response{}, err
	}
	var item WorkItem
	response, err := c.do(ctx, http.MethodPost, legacyCollectionPath(workspaceSlug, projectID), nil, request, &item)
	return item, response, err
}

func (c *Client) LegacyGet(ctx context.Context, workspaceSlug, projectID, issueID string, options ...DetailOptions) (WorkItem, plane.Response, error) {
	var item WorkItem
	response, err := c.do(ctx, http.MethodGet, legacyDetailPath(workspaceSlug, projectID, issueID), oneDetailOption(options).Query(), nil, &item)
	return item, response, err
}

func (c *Client) LegacyUpdate(ctx context.Context, workspaceSlug, projectID, issueID string, request UpdateWorkItemRequest, options ...DetailOptions) (WorkItem, plane.Response, error) {
	var item WorkItem
	response, err := c.do(ctx, http.MethodPatch, legacyDetailPath(workspaceSlug, projectID, issueID), oneDetailOption(options).Query(), request, &item)
	return item, response, err
}

func (c *Client) LegacyDelete(ctx context.Context, workspaceSlug, projectID, issueID string, options ...DetailOptions) (plane.Response, error) {
	return c.do(ctx, http.MethodDelete, legacyDetailPath(workspaceSlug, projectID, issueID), oneDetailOption(options).Query(), nil, nil)
}

func currentSearchPath(workspaceSlug string) string {
	return "/workspaces/" + plane.EscapePathSegment(workspaceSlug) + "/work-items/search/"
}

func legacySearchPath(workspaceSlug string) string {
	return "/workspaces/" + plane.EscapePathSegment(workspaceSlug) + "/issues/search/"
}

func currentIdentifierPath(workspaceSlug, projectIdentifier, issueIdentifier string) string {
	return "/workspaces/" + plane.EscapePathSegment(workspaceSlug) + "/work-items/" +
		plane.EscapePathSegment(projectIdentifier) + "-" + plane.EscapePathSegment(issueIdentifier) + "/"
}

func legacyIdentifierPath(workspaceSlug, projectIdentifier, issueIdentifier string) string {
	return "/workspaces/" + plane.EscapePathSegment(workspaceSlug) + "/issues/" +
		plane.EscapePathSegment(projectIdentifier) + "-" + plane.EscapePathSegment(issueIdentifier) + "/"
}

func currentCollectionPath(workspaceSlug, projectID string) string {
	return "/workspaces/" + plane.EscapePathSegment(workspaceSlug) + "/projects/" + plane.EscapePathSegment(projectID) + "/work-items/"
}

func legacyCollectionPath(workspaceSlug, projectID string) string {
	return "/workspaces/" + plane.EscapePathSegment(workspaceSlug) + "/projects/" + plane.EscapePathSegment(projectID) + "/issues/"
}

func currentDetailPath(workspaceSlug, projectID, workItemID string) string {
	return currentCollectionPath(workspaceSlug, projectID) + plane.EscapePathSegment(workItemID) + "/"
}

func legacyDetailPath(workspaceSlug, projectID, issueID string) string {
	return legacyCollectionPath(workspaceSlug, projectID) + plane.EscapePathSegment(issueID) + "/"
}

func currentRelationsPath(workspaceSlug, projectID, workItemID string) string {
	return currentDetailPath(workspaceSlug, projectID, workItemID) + "relations/"
}
