package warlot

import (
	"context"
	"net/http"
)

// InitProject initializes a new project and returns its identifiers.
func (c *Client) InitProject(ctx context.Context, req InitProjectRequest, opts ...CallOption) (*InitProjectResponse, error) {
	var out InitProjectResponse
	if err := c.doJSON(ctx, http.MethodPost, "/warlotSql/projects/init", buildHeaders(nil, opts...), req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// IssueAPIKey creates an API key for a project, returning the key and URL.
func (c *Client) IssueAPIKey(ctx context.Context, req IssueKeyRequest, opts ...CallOption) (*IssueKeyResponse, error) {
	var out IssueKeyResponse
	if err := c.doJSON(ctx, http.MethodPost, "/auth/issue", buildHeaders(nil, opts...), req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ResolveProject resolves a project by holder and name.
// Legacy fields are normalized to the modern shape if necessary.
func (c *Client) ResolveProject(ctx context.Context, req ResolveProjectRequest, opts ...CallOption) (*ResolveProjectResponse, error) {
	var out ResolveProjectResponse
	if err := c.doJSON(ctx, http.MethodPost, "/warlotSql/projects/resolve", buildHeaders(nil, opts...), req, &out); err != nil {
		return nil, err
	}
	if out.ProjectID == "" && out.LegacyProjectID != "" {
		out.ProjectID = out.LegacyProjectID
	}
	if out.DBID == "" && out.LegacyDBID != "" {
		out.DBID = out.LegacyDBID
	}
	return &out, nil
}
