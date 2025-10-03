package warlot

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// GetProjectStatus returns a project status map.
func (c *Client) GetProjectStatus(ctx context.Context, projectID string, opts ...CallOption) (ProjectStatus, error) {
	path := fmt.Sprintf("/warlotSql/projects/%s/status", url.PathEscape(projectID))
	var out ProjectStatus
	h := c.authHeaders()
	mergeHeaders(h, buildHeaders(nil, opts...))
	if err := c.doJSON(ctx, http.MethodGet, path, h, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// CommitProject persists recent changes of a project to the blockchain.
func (c *Client) CommitProject(ctx context.Context, projectID string, opts ...CallOption) (CommitResponse, error) {
	path := fmt.Sprintf("/warlotSql/projects/%s/commit", url.PathEscape(projectID))
	var out CommitResponse
	h := c.authHeaders()
	mergeHeaders(h, buildHeaders(nil, opts...))
	if err := c.doJSON(ctx, http.MethodPost, path, h, struct{}{}, &out); err != nil {
		return nil, err
	}
	return out, nil
}
