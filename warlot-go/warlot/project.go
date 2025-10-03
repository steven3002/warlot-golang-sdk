package warlot

import "context"

// Project is a light-weight handle bound to a specific project ID.
// It exposes ergonomic helpers that forward to Client methods.
type Project struct {
	ID     string
	Client *Client
}

// Project returns a handle for a given project ID.
func (c *Client) Project(id string) Project { return Project{ID: id, Client: c} }

// SQL executes a SQL statement in the bound project.
func (p Project) SQL(ctx context.Context, sql string, params []any, opts ...CallOption) (*SQLResponse, error) {
	return p.Client.ExecSQL(ctx, p.ID, SQLRequest{SQL: sql, Params: params}, opts...)
}

// Tables lists all tables in the bound project.
func (p Project) Tables(ctx context.Context, opts ...CallOption) (*ListTablesResponse, error) {
	return p.Client.ListTables(ctx, p.ID, opts...)
}

// Browse returns a page of rows for a table in the bound project.
func (p Project) Browse(ctx context.Context, table string, limit, offset int, opts ...CallOption) (*BrowseRowsResponse, error) {
	return p.Client.BrowseRows(ctx, p.ID, table, limit, offset, opts...)
}

// Schema returns a table schema from the bound project.
func (p Project) Schema(ctx context.Context, table string, opts ...CallOption) (TableSchema, error) {
	return p.Client.GetTableSchema(ctx, p.ID, table, opts...)
}

// Count returns the table count for the bound project.
func (p Project) Count(ctx context.Context, opts ...CallOption) (*TableCountResponse, error) {
	return p.Client.GetTableCount(ctx, p.ID, opts...)
}

// Status returns the project status map.
func (p Project) Status(ctx context.Context, opts ...CallOption) (ProjectStatus, error) {
	return p.Client.GetProjectStatus(ctx, p.ID, opts...)
}

// Commit triggers a commit for the bound project.
func (p Project) Commit(ctx context.Context, opts ...CallOption) (CommitResponse, error) {
	return p.Client.CommitProject(ctx, p.ID, opts...)
}
